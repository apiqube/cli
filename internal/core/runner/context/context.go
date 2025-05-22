package context

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

var _ interfaces.ExecutionContext = (*ctxBaseImpl)(nil)

type ctxBaseImpl struct {
	context.Context

	manifestsMutex sync.RWMutex
	manifests      map[string]manifests.Manifest

	storeMutex sync.RWMutex
	values     map[string]any
	kinds      map[string]reflect.Kind

	chansMutex sync.RWMutex
	passChans  map[string]chan any
	passKinds  map[string]reflect.Kind
	passDone   map[string]bool

	outputMutex sync.RWMutex
	output      interfaces.Output
}

func (c *ctxBaseImpl) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *ctxBaseImpl) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *ctxBaseImpl) Err() error {
	return c.Context.Err()
}

func (c *ctxBaseImpl) Value(key any) any {
	return c.Context.Value(key)
}

func (c *ctxBaseImpl) GetAllManifests() []manifests.Manifest {
	c.manifestsMutex.RLock()
	defer c.manifestsMutex.RUnlock()

	ret := make([]manifests.Manifest, 0, len(c.manifests))
	for _, m := range c.manifests {
		ret = append(ret, m)
	}

	return ret
}

func (c *ctxBaseImpl) GetManifestsByKind(kind string) ([]manifests.Manifest, error) {
	c.manifestsMutex.RLock()
	defer c.manifestsMutex.RUnlock()

	ret := make([]manifests.Manifest, 0, len(c.manifests))
	for _, m := range c.manifests {
		if m.GetKind() == kind {
			ret = append(ret, m)
		}
	}

	if len(ret) == 0 {
		return nil, fmt.Errorf("no such manifest: %s", kind)
	}

	return ret, nil
}

func (c *ctxBaseImpl) GetManifestByID(id string) (manifests.Manifest, error) {
	c.manifestsMutex.RLock()
	defer c.manifestsMutex.RUnlock()

	return c.manifests[id], nil
}

func (c *ctxBaseImpl) Set(key string, value any) {
	c.storeMutex.Lock()
	defer c.storeMutex.Unlock()
	c.values[key] = value
}

func (c *ctxBaseImpl) Get(key string) (any, bool) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()

	if v, ok := c.values[key]; ok {
		return v, true
	}

	return nil, false
}

func (c *ctxBaseImpl) Delete(key string) {
	c.storeMutex.Lock()
	defer c.storeMutex.Unlock()
	delete(c.values, key)
}

func (c *ctxBaseImpl) All() map[string]any {
	c.storeMutex.Lock()
	defer c.storeMutex.Unlock()
	return deepCopyMap(c.values, c.kinds)
}

func (c *ctxBaseImpl) SetTyped(key string, value any, kind reflect.Kind) {
	c.storeMutex.Lock()
	defer c.storeMutex.Unlock()
	c.values[key] = value
	c.kinds[key] = kind
}

func (c *ctxBaseImpl) GetTyped(key string) (any, reflect.Kind, bool) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()

	if v, ok := c.values[key]; ok {
		return v, c.kinds[key], true
	}

	return nil, reflect.Invalid, false
}

func (c *ctxBaseImpl) AsString(key string) (string, error) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return "", fmt.Errorf("value by %s not found", key)
	}

	val, is := v.(string)
	if !is {
		return "", fmt.Errorf("value by %s is not a string", key)
	}

	return val, nil
}

func (c *ctxBaseImpl) AsInt(key string) (int64, error) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return 0, fmt.Errorf("value by %s not found", key)
	}

	val, is := v.(int64)
	if !is {
		return 0, fmt.Errorf("value by %s is not a int", key)
	}

	return val, nil
}

func (c *ctxBaseImpl) AsFloat(key string) (float64, error) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return 0, fmt.Errorf("value by %s not found", key)
	}

	val, is := v.(float64)
	if !is {
		return 0, fmt.Errorf("value by %s is not a float", key)
	}

	return val, nil
}

func (c *ctxBaseImpl) AsBool(key string) (bool, error) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return false, fmt.Errorf("value by %s not found", key)
	}

	val, is := v.(bool)
	if !is {
		return false, fmt.Errorf("value by %s is not a bool", key)
	}

	return val, nil
}

func (c *ctxBaseImpl) AsStringSlice(key string) ([]string, error) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return nil, fmt.Errorf("value by %s not found", key)
	}

	val, is := v.([]string)
	if !is {
		return nil, fmt.Errorf("value by %s is not a string slice", key)
	}

	return val, nil
}

func (c *ctxBaseImpl) AsIntSlice(key string) ([]int, error) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return nil, fmt.Errorf("value by %s not found", key)
	}

	val, is := v.([]int)
	if !is {
		return nil, fmt.Errorf("value by %s is not a int slice", key)
	}

	return val, nil
}

func (c *ctxBaseImpl) AsMap(key string) (map[string]any, error) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return nil, fmt.Errorf("value by %s not found", key)
	}

	val, is := v.(map[string]any)
	if !is {
		return nil, fmt.Errorf("value by %s is not a map", key)
	}

	return val, nil
}

func (c *ctxBaseImpl) Channel(key string) chan any {
	c.chansMutex.Lock()
	defer c.chansMutex.Unlock()

	ch, ok := c.passChans[key]
	if !ok {
		ch = make(chan any, 1)
		c.passChans[key] = ch
	}
	return ch
}

func (c *ctxBaseImpl) ChannelT(key string, kind reflect.Kind) chan any {
	c.chansMutex.Lock()
	defer c.chansMutex.Unlock()

	ch, ok := c.passChans[key]
	if !ok {
		ch = make(chan any, 1)
		c.passChans[key] = ch
		c.passKinds[key] = kind
	}
	return ch
}

func (c *ctxBaseImpl) SafeSend(key string, val any) {
	c.chansMutex.RLock()
	ch, ok := c.passChans[key]
	c.chansMutex.RUnlock()

	if !ok {
		return
	}

	select {
	case ch <- val:
	default:
	}
}

func (c *ctxBaseImpl) SendOutput(msg any) {
	c.outputMutex.RLock()
	defer c.outputMutex.RUnlock()
	if c.output != nil {
		go c.output.ReceiveMsg(msg)
	}
}

func (c *ctxBaseImpl) GetOutput() interfaces.Output {
	c.outputMutex.Lock()
	defer c.outputMutex.Unlock()
	return c.output
}

func (c *ctxBaseImpl) SetOutput(out interfaces.Output) {
	c.outputMutex.Lock()
	defer c.outputMutex.Unlock()
	c.output = out
}

func deepCopyMap(m map[string]any, kinds map[string]reflect.Kind) map[string]any {
	if m == nil {
		return nil
	}

	cache := make(map[uintptr]any)
	newMap := make(map[string]any, len(m))

	for key, val := range m {
		if expectedKind, exists := kinds[key]; exists {
			actualKind := reflect.TypeOf(val).Kind()
			if actualKind != expectedKind {
				continue
			}
		}
		newMap[key] = deepCopyValue(val, cache)
	}
	return newMap
}

func deepCopyValue(value any, cache map[uintptr]any) any {
	if value == nil {
		return nil
	}

	val := reflect.ValueOf(value)

	if val.CanAddr() {
		ptr := val.UnsafeAddr()
		if cached, exists := cache[ptr]; exists {
			return cached
		}
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}

		elem := val.Elem()
		newPtr := reflect.New(elem.Type())

		if val.CanAddr() {
			cache[val.UnsafeAddr()] = newPtr.Interface()
		}

		newPtr.Elem().Set(reflect.ValueOf(deepCopyValue(elem.Interface(), cache)))
		return newPtr.Interface()

	case reflect.Map:
		newMap := reflect.MakeMap(val.Type())

		if val.CanAddr() {
			cache[val.UnsafeAddr()] = newMap.Interface()
		}

		iter := val.MapRange()
		for iter.Next() {
			newKey := deepCopyValue(iter.Key().Interface(), cache)
			newValue := deepCopyValue(iter.Value().Interface(), cache)
			newMap.SetMapIndex(
				reflect.ValueOf(newKey),
				reflect.ValueOf(newValue),
			)
		}
		return newMap.Interface()

	case reflect.Slice:
		newSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())

		if val.CanAddr() {
			cache[val.UnsafeAddr()] = newSlice.Interface()
		}

		for i := 0; i < val.Len(); i++ {
			newSlice.Index(i).Set(
				reflect.ValueOf(deepCopyValue(val.Index(i).Interface(), cache)),
			)
		}
		return newSlice.Interface()

	case reflect.Struct:
		newStruct := reflect.New(val.Type()).Elem()

		if val.CanAddr() {
			cache[val.UnsafeAddr()] = newStruct.Interface()
		}

		for i := 0; i < val.NumField(); i++ {
			if newStruct.Field(i).CanSet() {
				newStruct.Field(i).Set(
					reflect.ValueOf(deepCopyValue(val.Field(i).Interface(), cache)),
				)
			}
		}
		return newStruct.Interface()

	case reflect.Interface:
		if val.IsNil() {
			return nil
		}
		return deepCopyValue(val.Elem().Interface(), cache)

	default:
		return value
	}
}
