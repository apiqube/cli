package depends

import (
	"encoding/json"
	"fmt"
	"github.com/apiqube/cli/internal/core/runner/depends/rules"
	"strings"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// PassManager handles automatic data passing between tests
type PassManager struct {
	ctx              interfaces.ExecutionContext
	saveRequirements map[string]SaveRequirement
	graphResult      *Result
}

func NewPassManager(ctx interfaces.ExecutionContext, graphResult *Result) *PassManager {
	return &PassManager{
		ctx:              ctx,
		saveRequirements: graphResult.SaveRequirements,
		graphResult:      graphResult,
	}
}

// ShouldSaveResult determines if a test result should be saved for passing
func (pm *PassManager) ShouldSaveResult(manifestID string) bool {
	req, exists := pm.saveRequirements[manifestID]
	return exists && req.Required
}

// SaveTestResult saves test result data for future use
func (pm *PassManager) SaveTestResult(manifestID string, result TestResult) error {
	req, exists := pm.saveRequirements[manifestID]
	if !exists || !req.Required {
		return nil // No need to save
	}

	// Save the complete result first
	pm.ctx.Set(manifestID, result)

	// Save specific paths if required
	for _, path := range req.Paths {
		value, err := pm.extractValueByPath(result, path)
		if err != nil {
			// Log warning but don't fail - the path might be optional
			continue
		}

		key := fmt.Sprintf("%s.%s", manifestID, path)
		pm.ctx.Set(key, value)
	}

	// Send to PassStore channels for any waiting consumers
	pm.notifyWaitingConsumers(manifestID, result)

	return nil
}

// TestResult represents the result of a test execution
type TestResult struct {
	Request  RequestData       `json:"request"`
	Response ResponseData      `json:"response"`
	Status   int               `json:"status"`
	Headers  map[string]string `json:"headers"`
	Duration int64             `json:"duration_ms"`
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

// extractValueByPath extracts a value from test result using dot notation path
func (pm *PassManager) extractValueByPath(result TestResult, path string) (any, error) {
	// Convert result to map for easier navigation
	resultMap := pm.structToMap(result)

	// Navigate the path
	return pm.navigatePath(resultMap, path)
}

// structToMap converts struct to map using JSON marshaling
func (pm *PassManager) structToMap(v any) map[string]any {
	data, _ := json.Marshal(v)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}

// navigatePath navigates a dot-notation path in a map structure
func (pm *PassManager) navigatePath(data any, path string) (any, error) {
	if path == "" || path == "*" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		// Handle array indexing like "data[0]" or "data[-1]"
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			current = pm.handleArrayAccess(current, part)
			if current == nil {
				return nil, fmt.Errorf("array access failed for part: %s", part)
			}
		} else {
			// Handle map access
			if m, ok := current.(map[string]any); ok {
				if val, exists := m[part]; exists {
					current = val
				} else {
					return nil, fmt.Errorf("path not found: %s", part)
				}
			} else {
				return nil, fmt.Errorf("cannot access field %s on non-map type", part)
			}
		}
	}

	return current, nil
}

// handleArrayAccess handles array access patterns like "data[0]", "data[-1]", "data[*]"
func (pm *PassManager) handleArrayAccess(data any, part string) any {
	// Extract field name and index
	openBracket := strings.Index(part, "[")
	closeBracket := strings.Index(part, "]")

	if openBracket == -1 || closeBracket == -1 {
		return nil
	}

	fieldName := part[:openBracket]
	indexStr := part[openBracket+1 : closeBracket]

	// Get the field first
	var fieldValue any
	if fieldName != "" {
		if m, ok := data.(map[string]any); ok {
			if val, exists := m[fieldName]; exists {
				fieldValue = val
			} else {
				return nil
			}
		} else {
			return nil
		}
	} else {
		fieldValue = data
	}

	// Handle array access
	if arr, ok := fieldValue.([]any); ok {
		return pm.accessArray(arr, indexStr)
	}

	return nil
}

// accessArray handles different array access patterns
func (pm *PassManager) accessArray(arr []any, indexStr string) any {
	switch indexStr {
	case "*":
		// Return entire array
		return arr
	case "-1":
		// Return last element
		if len(arr) > 0 {
			return arr[len(arr)-1]
		}
		return nil
	default:
		// Try to parse as integer index
		var index int
		if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
			return nil
		}

		// Handle negative indices
		if index < 0 {
			index = len(arr) + index
		}

		if index >= 0 && index < len(arr) {
			return arr[index]
		}
		return nil
	}
}

// notifyWaitingConsumers sends data to PassStore channels
func (pm *PassManager) notifyWaitingConsumers(manifestID string, result TestResult) {
	// Get dependents of this manifest
	dependents := pm.graphResult.GetDependentsOf(manifestID)

	for _, dep := range dependents {
		if dep.Type == rules.DependencyTypeTemplate || dep.Type == rules.DependencyTypeValue {
			// Send complete result
			pm.ctx.SafeSend(manifestID, result)

			// Send specific paths if specified in metadata
			for _, path := range dep.Metadata.Paths {
				if value, err := pm.extractValueByPath(result, path); err == nil {
					key := fmt.Sprintf("%s.%s", manifestID, path)
					pm.ctx.SafeSend(key, value)
				}
			}
		}
	}
}

// WaitForDependency waits for a dependency to be available
func (pm *PassManager) WaitForDependency(manifestID, dependencyAlias, path string) (any, error) {
	// Try to get from DataStore first (synchronous)
	key := fmt.Sprintf("%s.%s", dependencyAlias, path)
	if value, exists := pm.ctx.Get(key); exists {
		return value, nil
	}

	// If not available, wait on channel (asynchronous)
	ch := pm.ctx.Channel(key)
	select {
	case value := <-ch:
		return value, nil
	case <-pm.ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for dependency %s", key)
	}
}

// GetDependencyValue gets a dependency value with fallback to waiting
func (pm *PassManager) GetDependencyValue(manifestID, dependencyAlias, path string) (any, error) {
	// First try direct access
	fullKey := fmt.Sprintf("%s.%s", dependencyAlias, path)
	if value, exists := pm.ctx.Get(fullKey); exists {
		return value, nil
	}

	// Try getting the full result and extracting the path
	if result, exists := pm.ctx.Get(dependencyAlias); exists {
		if testResult, ok := result.(TestResult); ok {
			return pm.extractValueByPath(testResult, path)
		}
	}

	// Last resort: wait for the dependency
	return pm.WaitForDependency(manifestID, dependencyAlias, path)
}

// ResolveTemplateValue resolves a template value like "{{ users-list.response.body.data[0].id }}"
func (pm *PassManager) ResolveTemplateValue(manifestID, templateStr string) (any, error) {
	// Parse template string to extract alias and path
	alias, path, err := pm.parseTemplateString(templateStr)
	if err != nil {
		return nil, err
	}

	return pm.GetDependencyValue(manifestID, alias, path)
}

// parseTemplateString parses "{{ alias.path }}" format
func (pm *PassManager) parseTemplateString(templateStr string) (alias, path string, err error) {
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
func (pm *PassManager) GetSaveRequirement(manifestID string) (SaveRequirement, bool) {
	req, exists := pm.saveRequirements[manifestID]
	return req, exists
}

// ListRequiredSaves returns all manifests that need to save data
func (pm *PassManager) ListRequiredSaves() map[string]SaveRequirement {
	result := make(map[string]SaveRequirement)
	for id, req := range pm.saveRequirements {
		if req.Required {
			result[id] = req
		}
	}
	return result
}
