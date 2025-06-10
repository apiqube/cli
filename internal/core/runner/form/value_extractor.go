package form

import (
	"strconv"
	"strings"
)

// DefaultValueExtractor implements ValueExtractor interface
type DefaultValueExtractor struct{}

func NewDefaultValueExtractor() *DefaultValueExtractor {
	return &DefaultValueExtractor{}
}

func (e *DefaultValueExtractor) Extract(parts []string, data any, indexStack []int) (any, bool) {
	current := data
	stackIndex := 0

	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch val := current.(type) {
		case map[string]any:
			v, ok := val[part]
			if !ok {
				return nil, false
			}
			current = v

		case []any:
			if part == "#" {
				// Use index from stack
				if stackIndex >= len(indexStack) {
					return nil, false
				}
				idx := indexStack[stackIndex]
				stackIndex++

				if idx < 0 || idx >= len(val) {
					return nil, false
				}
				current = val[idx]
			} else {
				// Direct array index
				idx, err := strconv.Atoi(part)
				if err != nil || idx < 0 || idx >= len(val) {
					return nil, false
				}
				current = val[idx]
			}

		default:
			// If we still have parts to process but reached a non-container type
			return nil, false
		}
	}

	return current, true
}
