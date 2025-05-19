package context

import (
	"context"
	"reflect"
	"sync"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

type ValuePair struct {
	Key   string
	Value any
	Kind  reflect.Kind
}

type CtxBuilder struct {
	ctx       context.Context
	manifests map[string]manifests.Manifest
	values    map[string]any
	kinds     map[string]reflect.Kind
	passChans map[string]chan any
	passKinds map[string]reflect.Kind
	passDone  map[string]bool
	output    interfaces.Output
}

func NewCtxBuilder() *CtxBuilder {
	return &CtxBuilder{
		ctx:       context.Background(),
		manifests: make(map[string]manifests.Manifest),
		values:    make(map[string]any),
		kinds:     make(map[string]reflect.Kind),
		passChans: make(map[string]chan any),
		passKinds: make(map[string]reflect.Kind),
		passDone:  make(map[string]bool),
	}
}

func (b *CtxBuilder) WithContext(ctx context.Context) *CtxBuilder {
	b.ctx = ctx
	return b
}

func (b *CtxBuilder) WithManifests(manifests ...manifests.Manifest) *CtxBuilder {
	for _, m := range manifests {
		b.manifests[m.GetID()] = m
	}
	return b
}

func (b *CtxBuilder) WithValue(key string, value any, kind reflect.Kind) *CtxBuilder {
	b.values[key] = value
	b.kinds[key] = kind
	return b
}

func (b *CtxBuilder) WithValues(valuesPairs ...ValuePair) *CtxBuilder {
	for _, pair := range valuesPairs {
		b.values[pair.Key] = pair.Value
		b.kinds[pair.Key] = pair.Kind
	}
	return b
}

func (b *CtxBuilder) WithPassChan(key string, ch chan any, kind reflect.Kind) *CtxBuilder {
	b.passChans[key] = ch
	b.passKinds[key] = kind
	return b
}

func (b *CtxBuilder) WithOutput(output interfaces.Output) *CtxBuilder {
	b.output = output
	return b
}

func (b *CtxBuilder) Build() interfaces.ExecutionContext {
	return &ctxBaseImpl{
		Context:        b.ctx,
		manifests:      b.manifests,
		values:         b.values,
		kinds:          b.kinds,
		passChans:      b.passChans,
		passKinds:      b.passKinds,
		passDone:       b.passDone,
		output:         b.output,
		manifestsMutex: sync.RWMutex{},
		storeMutex:     sync.RWMutex{},
		chansMutex:     sync.RWMutex{},
		outputMutex:    sync.RWMutex{},
	}
}

func (b *CtxBuilder) Reset() {
	*b = CtxBuilder{ctx: b.ctx}
}
