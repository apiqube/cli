package store

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/apiqube/cli/internal/core/manifests"
	parcer "github.com/apiqube/cli/internal/core/yaml"
	"github.com/dgraph-io/badger/v4"
	"gopkg.in/yaml.v3"
)

const (
	BadgerDatabaseDirPath = "qube/storage"
)

const (
	manifestListKeyPrefix = "manifest_list:"
)

type Storage struct {
	db *badger.DB
}

func NewStorage() (*Storage, error) {
	path, err := xdg.DataFile(BadgerDatabaseDirPath)
	if err != nil {
		return nil, fmt.Errorf("error getting data file path: %v", err)
	}

	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating data file path: %v", err)
	}

	db, err := badger.Open(badger.DefaultOptions(path).WithLogger(nil))
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) LoadManifestList() ([]string, error) {
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

func (s *Storage) SaveManifests(mans ...manifests.Manifest) error {
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

func (s *Storage) LoadManifests(ids ...string) ([]manifests.Manifest, error) {
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
