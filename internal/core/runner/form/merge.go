package form

// mergeProcessedData merges two maps (base and overlay), overlaying keys over base.
func mergeProcessedData(base map[string]any, overlay map[string]any) map[string]any {
	if base == nil && overlay == nil {
		return nil
	}
	if base == nil {
		return overlay
	}
	if overlay == nil {
		return base
	}
	result := make(map[string]any, len(base)+len(overlay))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		result[k] = v
	}
	return result
}
