package store

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/apiqube/cli/internal/manifests"
	"github.com/apiqube/cli/internal/ui"
	parcer "github.com/apiqube/cli/internal/yaml"
	"github.com/dgraph-io/badger/v4"
	"gopkg.in/yaml.v3"
)

const (
	BadgerDatabaseDirPath = "qube/storage"
)

var (
	manifestListKeyPrefix = "manifest_list:"
)

var (
	instance *Storage
	once     sync.Once
)

type Storage struct {
	db          *badger.DB
	enabled     bool
	initialized bool
}

func Init() {
	once.Do(func() {
		path, err := xdg.DataFile(BadgerDatabaseDirPath)
		if err != nil {
			ui.Errorf("Failed to open database: %v", err)
			return
		}

		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			ui.Errorf("Failed to create database: %v", err)
			return
		}

		db, err := badger.Open(badger.DefaultOptions(path).WithLogger(nil))
		if err != nil {
			ui.Errorf("Failed to open database: %v", err)
			return
		}

		instance = &Storage{
			db:          db,
			enabled:     true,
			initialized: true,
		}
	})
}

func Stop() {
	if instance != nil && instance.initialized {
		instance.enabled = false
		instance.initialized = false
		if err := instance.db.Close(); err != nil {
			ui.Errorf("Failed to close database: %v", err)
		}
		instance = nil
	}
}

func IsEnabled() bool {
	return instance != nil && instance.enabled
}

func LoadManifestList() ([]string, error) {
	if !IsEnabled() {
		return nil, nil
	}

	var manifestList []string

	err := instance.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(manifestListKeyPrefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			key := it.Item().Key()
			manifestList = append(manifestList, strings.TrimPrefix(string(key), manifestListKeyPrefix))
		}

		return nil
	})

	return manifestList, err
}

func SaveManifests(mans ...manifests.Manifest) error {
	if !IsEnabled() {
		return nil
	}

	return instance.db.Update(func(txn *badger.Txn) error {
		var data []byte
		var err error

		for _, m := range mans {
			data, err = yaml.Marshal(m)
			if err != nil {
				return err
			}

			if err = txn.Set(genManifestKey(m.GetID()), data); err != nil {
				return err
			}

			if err = txn.Set(genManifestListKey(m.GetID()), nil); err != nil {
				return err
			}
		}

		return nil
	})
}

func LoadManifests(ids ...string) ([]manifests.Manifest, error) {
	if !IsEnabled() {
		return nil, nil
	}

	var results []manifests.Manifest
	var rErr error

	err := instance.db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		var err error

		for _, id := range ids {
			item, err = txn.Get(genManifestKey(id))
			if errors.Is(err, badger.ErrKeyNotFound) {
				rErr = errors.Join(rErr, fmt.Errorf("manifest %s not found", id))
				continue
			} else if err != nil {
				rErr = errors.Join(rErr, err)
				continue
			}

			var mans []manifests.Manifest

			if err = item.Value(func(data []byte) error {
				if mans, err = parcer.ParseManifests(data); err != nil {
					return err
				}

				results = append(results, mans...)
				return nil
			}); err != nil {
				rErr = errors.Join(rErr, err)
			}
		}

		return nil
	})

	return results, errors.Join(rErr, err)
}

func genManifestKey(id string) []byte {
	return []byte(id)
}

func genManifestListKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", manifestListKeyPrefix, id))
}
