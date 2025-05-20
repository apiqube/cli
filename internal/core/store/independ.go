package store

import (
	"errors"
	"sync"

	"github.com/apiqube/cli/ui/cli"

	"github.com/apiqube/cli/internal/core/manifests"
)

var errStoreNotInitialized = errors.New("store not initialized")

var (
	instance             *Storage
	once                 sync.Once
	enabled, initialized bool
)

func Init() {
	once.Do(func() {
		db, err := NewStorage()
		if err != nil {
			cli.Errorf("Error initializing storage: %v", err)
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
			cli.Errorf("Failed to close database: %v", err)
		}

		if err := instance.index.Close(); err != nil {
			cli.Errorf("Failed to close index: %v", err)
		}

		instance = nil
	}
}

func IsEnabled() bool {
	return instance != nil && enabled
}

func Save(mans ...manifests.Manifest) error {
	if !IsEnabled() {
		return errStoreNotInitialized
	}

	return instance.Save(mans...)
}

func Load(opts LoadOptions) ([]manifests.Manifest, error) {
	if !IsEnabled() {
		return nil, errStoreNotInitialized
	}

	return instance.Load(opts)
}

func Search(query Query) ([]manifests.Manifest, error) {
	if !IsEnabled() {
		return nil, errStoreNotInitialized
	}

	return instance.Search(query)
}

func Rollback(id string, targetVersion int) error {
	if !IsEnabled() {
		return errStoreNotInitialized
	}

	return instance.Rollback(id, targetVersion)
}

func CleanupOldVersions(id string, keep int) error {
	if !IsEnabled() {
		return errStoreNotInitialized
	}

	return instance.CleanupOldVersions(id, keep)
}
