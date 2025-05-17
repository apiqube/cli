package store

import (
	"sync"
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/ui"
)

var (
	instance             *Storage
	once                 sync.Once
	enabled, initialized bool
)

func Init() {
	once.Do(func() {
		db, err := NewStorage()
		if err != nil {
			ui.Errorf("Error initializing storage: %v", err)
		}

		instance = db
		enabled = true
		initialized = true
	})
}

func Stop() {
	if instance != nil && initialized {
		enabled = false
		initialized = false
		if err := instance.db.Close(); err != nil {
			ui.Errorf("Failed to close database: %v", err)
		}
		instance = nil
	}
}

func IsEnabled() bool {
	return instance != nil && enabled
}

func SaveManifests(mans ...manifests.Manifest) error {
	if !isEnabled() {
		return nil
	}

	return instance.SaveManifests(mans...)
}

func LoadManifests(ids ...string) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.LoadManifests()
}

func LoadManifest(id string) (manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.LoadManifest(id)
}

func FindManifestsByKind(kind string) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByKind(kind)
}

func FindManifestsByVersion(lowVersion, heightVersion uint8) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByVersion(lowVersion, heightVersion)
}

func FindManifestByName(name string) (manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestByName(name)
}

func FindManifestsByNameWildcard(namePattern string) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByNameWildcard(namePattern)
}

func FindManifestByNamespace(namespace string) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestByNamespace(namespace)
}

func FindManifestByDependencies(dependencies []string, requireAll bool) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestByDependencies(dependencies, requireAll)
}

func FindManifestByHash(hash string) (manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestByHash(hash)
}

func FindManifestsByCreatedAtRange(start, end time.Time) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByCreatedAtRange(start, end)
}

func FindManifestsByCreatedBy(createdBy string) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByCreatedBy(createdBy)
}

func FindManifestsByUpdatedAtRange(start, end time.Time) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByUpdatedAtRange(start, end)
}

func FindManifestsByUpdatedBy(updatedBy string) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByUpdatedBy(updatedBy)
}

func FindManifestsByUsedBy(usedBy string) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByUsedBy(usedBy)
}

func FindManifestsByLastAppliedRange(start, end time.Time) ([]manifests.Manifest, error) {
	if !isEnabled() {
		return nil, nil
	}

	return instance.FindManifestsByLastAppliedRange(start, end)
}

func isEnabled() bool {
	if !IsEnabled() {
		ui.Errorf("Database instance not ready")
		return false
	}
	return true
}
