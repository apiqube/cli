package accessor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

type DataAccessor interface {
	Get(path string) (any, error)
	GetString(path string) (string, error)
	GetInt(path string) (int64, error)
	GetFloat(path string) (float64, error)
	GetBool(path string) (bool, error)
	GetStringSlice(path string) ([]string, error)
	GetMap(path string) (map[string]any, error)
}

var _ DataAccessor = (*Accessor)(nil)

type Accessor struct {
	store interfaces.DataStore
}

func NewAccessor(store interfaces.DataStore) *Accessor {
	return &Accessor{store: store}
}

func (a *Accessor) Get(path string) (any, error) {
	key, subPath := splitKeyAndPath(path)
	root, ok := a.store.Get(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return walkPath(root, subPath)
}

func (a *Accessor) GetString(path string) (string, error) {
	v, err := a.Get(path)
	if err != nil {
		return "", err
	}

	val, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("not a string at path %s", path)
	}

	return val, nil
}

func (a *Accessor) GetInt(path string) (int64, error) {
	v, err := a.Get(path)
	if err != nil {
		return -1, err
	}

	val, ok := v.(int64)
	if !ok {
		return -1, fmt.Errorf("not a int at path %s", path)
	}

	return val, nil
}

func (a *Accessor) GetFloat(path string) (float64, error) {
	v, err := a.Get(path)
	if err != nil {
		return -1, err
	}

	val, ok := v.(float64)
	if !ok {
		return -1, fmt.Errorf("not a float at path %s", path)
	}

	return val, nil
}

func (a *Accessor) GetBool(path string) (bool, error) {
	v, err := a.Get(path)
	if err != nil {
		return false, err
	}

	val, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("not a bool at path %s", path)
	}

	return val, nil
}

func (a *Accessor) GetStringSlice(path string) ([]string, error) {
	v, err := a.Get(path)
	if err != nil {
		return []string{}, err
	}

	val, ok := v.([]string)
	if !ok {
		return []string{}, fmt.Errorf("not a string slice at path %s", path)
	}

	return val, nil
}

func (a *Accessor) GetMap(path string) (map[string]any, error) {
	v, err := a.Get(path)
	if err != nil {
		return map[string]any{}, err
	}

	val, ok := v.(map[string]any)
	if !ok {
		return map[string]any{}, fmt.Errorf("not a map at path %s", path)
	}

	return val, nil
}

func splitKeyAndPath(full string) (string, string) {
	if idx := strings.LastIndex(full, "."); idx != -1 {
		return full[:idx], full[idx+1:]
	}

	return full, ""
}

func walkPath(v any, path string) (any, error) {
	if path == "" {
		return v, nil
	}

	parts := strings.Split(path, ".")
	cur := v

	for _, part := range parts {
		switch val := cur.(type) {
		case map[string]any:
			cur = val[part]
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx >= len(val) {
				return nil, fmt.Errorf("invalid index: %s", part)
			}
			cur = val[idx]
		default:
			return nil, fmt.Errorf("unsupported or unexpected type at %s", part)
		}
	}
	return cur, nil
}
