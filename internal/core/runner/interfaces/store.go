package interfaces

import (
	"reflect"

	"github.com/apiqube/cli/internal/core/manifests"
)

type ManifestStore interface {
	GetAllManifests() []manifests.Manifest
	GetManifestsByKind(kind string) ([]manifests.Manifest, error)
	GetManifestByID(id string) (manifests.Manifest, error)
}

type DataStore interface {
	Set(key string, value any)
	Get(key string) (any, bool)
	Delete(key string)
	All() map[string]any

	SetTyped(key string, value any, kind reflect.Kind)
	GetTyped(key string) (any, reflect.Kind, bool)

	AsString(key string) (string, error)
	AsInt(key string) (int64, error)
	AsFloat(key string) (float64, error)
	AsBool(key string) (bool, error)
	AsStringSlice(key string) ([]string, error)
	AsIntSlice(key string) ([]int, error)
	AsMap(key string) (map[string]any, error)
}

type PassStore interface {
	Channel(key string) chan any
	ChannelT(key string, kind reflect.Kind) chan any
	SafeSend(key string, val any)
}

type OutputStore interface {
	SendOutput(msg any)
	GetOutput() Output
	SetOutput(out Output)
}
