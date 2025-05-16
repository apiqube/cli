package store

import (
	"sync"

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

func LoadManifestList() ([]string, error) {
	if !IsEnabled() {
		ui.Errorf("Database instance not ready")
		return nil, nil
	}

	return instance.LoadManifestList()
}

func SaveManifests(mans ...manifests.Manifest) error {
	if !IsEnabled() {
		ui.Errorf("Database instance not ready")
		return nil
	}

	return instance.SaveManifests(mans...)
}

func LoadManifests(ids ...string) ([]manifests.Manifest, error) {
	if !IsEnabled() {
		ui.Errorf("Database instance not ready")
		return nil, nil
	}

	return instance.LoadManifests(ids...)
}
