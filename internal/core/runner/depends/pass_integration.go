package depends

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/apiqube/cli/internal/core/runner/depends/rules"
	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// PassManager handles automatic data passing between tests
type PassManager struct {
	mx               sync.RWMutex
	saveRequirements map[string]SaveRequirement
	graphResult      *Result
	channels         map[string]chan any
}

func NewPassManager(graphResult *Result) *PassManager {
	return &PassManager{
		mx:               sync.RWMutex{},
		saveRequirements: graphResult.SaveRequirements,
		graphResult:      graphResult,
		channels:         make(map[string]chan any),
	}
}

func (m *PassManager) Close() {
	for _, ch := range m.channels {
		close(ch)
	}
}

// ShouldSaveResult determines if a test result should be saved for passing
func (m *PassManager) ShouldSaveResult(manifestID string) bool {
	req, exists := m.saveRequirements[manifestID]
	return exists && req.Required
}

// SaveTestResult saves test result data for future use
func (m *PassManager) SaveTestResult(ctx interfaces.ExecutionContext, manifestID string, data TestData) error {
	req, exists := m.saveRequirements[manifestID]
	if !exists || !req.Required {
		return nil // No need to save
	}

	// Save specific paths if required
	for _, path := range req.Paths {
		value, err := m.extractValueByPath(data, path)
		if err != nil {
			ctx.GetOutput().Logf(interfaces.ErrorLevel, "failed to extract value for path: %s", path)
			// Log warning but don't fail - the path might be optional
			continue
		}

		key := fmt.Sprintf("%s.%s", manifestID, path)
		ctx.SetTyped(key, value, reflect.TypeOf(value).Kind())

		// Send to PassStore channels for any waiting consumers
		m.notifyConsumers(ctx, manifestID, data)
	}

	return nil
}

// TestData represents the result of a test execution
type TestData struct {
	Request  RequestData       `json:"request"`
	Response ResponseData      `json:"response"`
	Status   int               `json:"status"`
	Headers  map[string]string `json:"headers"`
	Duration time.Duration     `json:"duration"`
	Error    string            `json:"error,omitempty"`
}

type RequestData struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
}

type ResponseData struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
}

// extractValueByPath extracts a value from test result using gjson
func (m *PassManager) extractValueByPath(result TestData, path string) (any, error) {
	if path == "" || path == "*" {
		return result, nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	res := gjson.ParseBytes(data).Get(path)
	if !res.Exists() {
		return result, nil
	}

	return res.Value(), nil
}

// notifyConsumers sends data to PassStore channels
func (m *PassManager) notifyConsumers(ctx interfaces.ExecutionContext, manifestID string, result TestData) {
	// Get dependents of this manifest
	dependents := m.graphResult.GetDependentsOf(manifestID)

	m.mx.Lock()
	defer m.mx.Unlock()

	for _, dep := range dependents {
		if dep.Type == rules.DependencyTypeTemplate {
			from := dep.From
			if dep.Metadata.Alias != "" {
				from = dep.Metadata.Alias
			}

			_, exists := m.channels[from]
			if !exists {
				m.channels[from] = ctx.Channel(from)
			}

			// Send complete result
			ctx.SafeSend(from, result)

			// Send specific paths if specified in metadata
			for _, path := range dep.Metadata.Paths {
				if value, err := m.extractValueByPath(result, path); err == nil {
					key := fmt.Sprintf("%s.%s", from, path)
					ctx.SafeSend(key, value)
				}
			}
		}
	}
}

// WaitForDependency waits for a dependency to be available
func (m *PassManager) WaitForDependency(ctx interfaces.ExecutionContext, dependencyAlias, path string) (any, error) {
	// Try to get from DataStore first (synchronous)
	key := fmt.Sprintf("%s.%s", dependencyAlias, path)
	if value, exists := ctx.Get(key); exists {
		return value, nil
	}

	// If not available, wait on channel (asynchronous)
	ch := ctx.Channel(key)
	select {
	case value := <-ch:
		return value, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for dependency %s", key)
	}
}

// GetDependencyValue gets a dependency value with fallback to waiting
func (m *PassManager) GetDependencyValue(ctx interfaces.ExecutionContext, dependencyAlias, path string) (any, error) {
	// First try direct access
	fullKey := fmt.Sprintf("%s.%s", dependencyAlias, path)
	if value, exists := ctx.Get(fullKey); exists {
		return value, nil
	}

	// Try getting the full result and extracting the path
	if result, exists := ctx.Get(dependencyAlias); exists {
		if testResult, ok := result.(TestData); ok {
			return m.extractValueByPath(testResult, path)
		}
	}

	// Last resort: wait for the dependency
	return m.WaitForDependency(ctx, dependencyAlias, path)
}

// ResolveTemplateValue resolves a template value like "{{ users-list.response.body.data[0].id }}"
func (m *PassManager) ResolveTemplateValue(ctx interfaces.ExecutionContext, templateStr string) (any, error) {
	// Parse template string to extract alias and path
	alias, path, err := m.parseTemplateString(templateStr)
	if err != nil {
		return nil, err
	}

	return m.GetDependencyValue(ctx, alias, path)
}

// parseTemplateString parses "{{ alias.path }}" format
func (m *PassManager) parseTemplateString(templateStr string) (alias, path string, err error) {
	// Remove {{ and }} and trim spaces
	content := strings.TrimSpace(templateStr)
	if strings.HasPrefix(content, "{{") && strings.HasSuffix(content, "}}") {
		content = strings.TrimSpace(content[2 : len(content)-2])
	}

	// Split on first dot
	parts := strings.SplitN(content, ".", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid template format: %s", templateStr)
	}

	return parts[0], parts[1], nil
}

// GetSaveRequirement returns save requirement for a manifest
func (m *PassManager) GetSaveRequirement(manifestID string) (SaveRequirement, bool) {
	req, exists := m.saveRequirements[manifestID]
	return req, exists
}

// ListRequiredSaves returns all manifests that need to save data
func (m *PassManager) ListRequiredSaves() map[string]SaveRequirement {
	result := make(map[string]SaveRequirement)
	for id, req := range m.saveRequirements {
		if req.Required {
			result[id] = req
		}
	}
	return result
}
