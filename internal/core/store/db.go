package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests/parsing"

	"github.com/adrg/xdg"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v4"
)

const (
	StorageDirPath = "qube/storage"
)

const (
	manifestHashKeyPrefix = "manifest_hash:"
)

type Storage struct {
	db    *badger.DB
	index bleve.Index
}

func NewStorage() (*Storage, error) {
	path, err := xdg.DataFile(StorageDirPath)
	if err != nil {
		return nil, fmt.Errorf("error getting data file path: %v", err)
	}

	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating data file path: %v", err)
	}

	badgerPath := filepath.Join(path, "manifests")
	blevePath := filepath.Join(path, "index")

	db, err := badger.Open(badger.DefaultOptions(badgerPath).WithLogger(nil))
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	var index bleve.Index
	index, err = bleve.Open(blevePath)
	if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
		mapping := buildBleveMapping()
		index, err = bleve.New(blevePath, mapping)
	}
	if err != nil {
		return nil, fmt.Errorf("error opening index: %v", err)
	}

	return &Storage{
		db:    db,
		index: index,
	}, nil
}

func (s *Storage) SaveManifests(mans ...manifests.Manifest) error {
	return instance.db.Update(func(txn *badger.Txn) error {
		var data []byte
		var err error

		for _, m := range mans {
			if man, ok := m.(manifests.Marshaler); ok {
				data, err = man.MarshalJSON()
			} else {
				data, err = json.Marshal(m)
			}

			if err = txn.Set(genManifestKey(m.GetID()), data); err != nil {
				return err
			}

			if err = s.index.Index(m.GetID(), m.Index()); err != nil {
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
				if mans, err = parsing.ParseManifestsAsYAML(data); err != nil {
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

func (s *Storage) FindManifestByHash(hash string) (manifests.Manifest, bool, error) {
	var manifest manifests.Manifest
	found := false

	query := bleve.NewTermQuery(hash)
	query.SetField("hash")

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 1

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return manifest, false, fmt.Errorf("search failed: %w", err)
	}

	if searchResults.Total > 0 {
		hit := searchResults.Hits[0]
		err = s.db.View(func(txn *badger.Txn) error {
			var item *badger.Item
			item, err = txn.Get(genManifestKey(hit.ID))
			if err != nil {
				return err
			}

			return item.Value(func(val []byte) error {
				if err = json.Unmarshal(val, &manifest); err != nil {
					return fmt.Errorf("failed to unmarshal manifest: %w", err)
				}
				found = true
				return nil
			})
		})

		if err != nil {
			return manifest, false, fmt.Errorf("failed to load manifest: %w", err)
		}
	}

	return manifest, found, nil
}

func (s *Storage) LoadManifestHashes() ([]string, error) {
	var results []string
	var rErr error

	err := instance.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(manifestHashKeyPrefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			key := it.Item().Key()
			results = append(results, strings.TrimPrefix(string(key), manifestHashKeyPrefix))
		}

		return nil
	})

	return results, errors.Join(rErr, err)
}

func (s *Storage) SaveManifestHash(hash string) error {
	var rErr error

	err := instance.db.Update(func(txn *badger.Txn) error {
		return txn.Set(genManifestHashKey(hash), []byte(hash))
	})
	if err != nil {
		return fmt.Errorf("error saving manifest hash: %v", err)
	}

	return errors.Join(rErr, err)
}
