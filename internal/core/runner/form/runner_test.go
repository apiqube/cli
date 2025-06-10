package form

import (
	"context"
	"reflect"
	"testing"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// MockExecutionContext for testing
type MockExecutionContext struct {
	context.Context
	data map[string]any
}

func NewMockExecutionContext() *MockExecutionContext {
	return &MockExecutionContext{
		Context: context.Background(),
		data:    make(map[string]any),
	}
}

func (m *MockExecutionContext) Set(key string, value any) {
	m.data[key] = value
}

func (m *MockExecutionContext) Get(key string) (any, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *MockExecutionContext) Delete(key string) {
	delete(m.data, key)
}

func (m *MockExecutionContext) All() map[string]any {
	return m.data
}

func (m *MockExecutionContext) SetTyped(key string, value any, kind reflect.Kind) {
	m.data[key] = value
}

func (m *MockExecutionContext) GetTyped(key string) (any, reflect.Kind, bool) {
	val, ok := m.data[key]
	return val, reflect.TypeOf(val).Kind(), ok
}

func (m *MockExecutionContext) AsString(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}
	return "", nil
}

func (m *MockExecutionContext) AsInt(key string) (int64, error) {
	if val, ok := m.data[key]; ok {
		if i, ok := val.(int64); ok {
			return i, nil
		}
	}
	return 0, nil
}

func (m *MockExecutionContext) AsFloat(key string) (float64, error) {
	if val, ok := m.data[key]; ok {
		if f, ok := val.(float64); ok {
			return f, nil
		}
	}
	return 0, nil
}

func (m *MockExecutionContext) AsBool(key string) (bool, error) {
	if val, ok := m.data[key]; ok {
		if b, ok := val.(bool); ok {
			return b, nil
		}
	}
	return false, nil
}

func (m *MockExecutionContext) AsStringSlice(key string) ([]string, error) {
	if val, ok := m.data[key]; ok {
		if slice, ok := val.([]string); ok {
			return slice, nil
		}
	}
	return nil, nil
}

func (m *MockExecutionContext) AsIntSlice(key string) ([]int, error) {
	if val, ok := m.data[key]; ok {
		if slice, ok := val.([]int); ok {
			return slice, nil
		}
	}
	return nil, nil
}

func (m *MockExecutionContext) AsMap(key string) (map[string]any, error) {
	if val, ok := m.data[key]; ok {
		if m, ok := val.(map[string]any); ok {
			return m, nil
		}
	}
	return nil, nil
}

// Implement other required interfaces (stubs for testing)
func (m *MockExecutionContext) GetAllManifests() []manifests.Manifest { return nil }

func (m *MockExecutionContext) GetManifestsByKind(kind string) ([]manifests.Manifest, error) {
	return nil, nil
}

func (m *MockExecutionContext) GetManifestByID(id string) (manifests.Manifest, error) {
	return nil, nil
}
func (m *MockExecutionContext) Channel(key string) chan any                     { return nil }
func (m *MockExecutionContext) ChannelT(key string, kind reflect.Kind) chan any { return nil }
func (m *MockExecutionContext) SafeSend(key string, val any)                    {}
func (m *MockExecutionContext) SendOutput(msg any)                              {}
func (m *MockExecutionContext) GetOutput() interfaces.Output                    { return nil }
func (m *MockExecutionContext) SetOutput(out interfaces.Output)                 {}

func TestRunner_Apply(t *testing.T) {
	runner := NewRunner()
	ctx := NewMockExecutionContext()
	ctx.Set("username", "john")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple template",
			input:    "Hello {{ username }}",
			expected: "Hello john",
		},
		{
			name:     "no template",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "multiple templates",
			input:    "{{ username }} says hello to {{ username }}",
			expected: "john says hello to john",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runner.Apply(ctx, tt.input, nil)
			if result != tt.expected {
				t.Errorf("Apply() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRunner_ApplyBody(t *testing.T) {
	runner := NewRunner()
	ctx := NewMockExecutionContext()
	ctx.Set("username", "john")

	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "simple body",
			input: map[string]any{
				"name": "{{ username }}",
				"age":  25,
			},
			expected: map[string]any{
				"name": "john",
				"age":  25,
			},
		},
		{
			name: "nested body",
			input: map[string]any{
				"user": map[string]any{
					"name": "{{ username }}",
					"details": map[string]any{
						"active": true,
					},
				},
			},
			expected: map[string]any{
				"user": map[string]any{
					"name": "john",
					"details": map[string]any{
						"active": true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runner.ApplyBody(ctx, tt.input, nil)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ApplyBody() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRunner_MapHeaders(t *testing.T) {
	runner := NewRunner()
	ctx := NewMockExecutionContext()
	ctx.Set("token", "abc123")

	input := map[string]string{
		"Authorization": "Bearer {{ token }}",
		"Content-Type":  "application/json",
	}

	expected := map[string]string{
		"Authorization": "Bearer abc123",
		"Content-Type":  "application/json",
	}

	result := runner.MapHeaders(ctx, input, nil)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("MapHeaders() = %v, want %v", result, expected)
	}
}
