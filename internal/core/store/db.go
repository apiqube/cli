package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apiqube/cli/internal/operations"

	"github.com/apiqube/cli/internal/core/manifests/kinds"
	"github.com/apiqube/cli/internal/core/manifests/utils"

	"github.com/adrg/xdg"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v4"
)

const (
	StorageDirPath = "ApiQube/storage"
)

const (
	MaxManifestVersion = math.MaxUint8 - 230
)

type LoadOptions struct {
	IDs      []string
	Versions map[string]int // id -> version
	Hash     string
	Latest   bool
	All      bool
	SortBy   []string
	Limit    int
}

type Storage struct {
	db    *badger.DB
	index bleve.Index
}

func NewStorage() (*Storage, error) {
	return NewStorageWithPath(StorageDirPath)
}

func NewStorageWithPath(path string) (*Storage, error) {
	path, err := xdg.DataFile(path)
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

	var bIndex bleve.Index
	bIndex, err = bleve.Open(blevePath)
	if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
		mapping := buildBleveMapping()
		bIndex, err = bleve.New(blevePath, mapping)
	}
	if err != nil {
		return nil, fmt.Errorf("error opening index: %v", err)
	}

	return &Storage{
		db:    db,
		index: bIndex,
	}, nil
}

func (s *Storage) Save(mans ...manifests.Manifest) error {
	return instance.db.Update(func(txn *badger.Txn) error {
		for _, m := range mans {
			currentVer, err := s.getCurrentVersion(txn, m.GetID())
			if err != nil {
				return fmt.Errorf("version check failed: %v", err)
			}

			newVersion := currentVer + 1

			meta := m.GetMeta()
			meta.SetVersion(uint8(newVersion))
			meta.SetUpdatedAt(time.Now())
			meta.SetUpdatedBy("qube")

			versionedKey := genVersionedKey(m.GetID(), newVersion)
			latestKey := genLatestKey(m.GetID())

			data, err := json.Marshal(m)
			if err != nil {
				return fmt.Errorf("marshaling error: %v", err)
			}

			if err = txn.Set(versionedKey, data); err != nil {
				return err
			}

			if err = txn.Set(latestKey, versionedKey); err != nil {
				return err
			}

			if err = s.indexCurrentManifest(m, string(versionedKey), currentVer); err != nil {
				return err
			}

			if newVersion > MaxManifestVersion {
				if err = s.cleanupOldVersions(txn, m.GetID(), MaxManifestVersion); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *Storage) Load(opts LoadOptions) ([]manifests.Manifest, error) {
	var results []manifests.Manifest

	if opts.Hash != "" {
		m, err := s.loadByHash(opts.Hash)
		if err != nil {
			return nil, err
		}
		return []manifests.Manifest{m}, nil
	}

	if len(opts.Versions) > 0 {
		for id, version := range opts.Versions {
			m, err := s.loadVersion(id, version)
			if err != nil {
				return nil, fmt.Errorf("failed to load %s@v%d: %w", id, version, err)
			}
			results = append(results, m)
		}
		return results, nil
	}

	if len(opts.IDs) > 0 {
		return s.loadBulk(opts)
	}

	if opts.All || opts.Latest {
		return s.loadLatest()
	}

	return nil, fmt.Errorf("no valid load options specified")
}

func (s *Storage) Search(query Query) ([]manifests.Manifest, error) {
	searchResult, err := query.Execute(s.index)
	if err != nil {
		return nil, fmt.Errorf("failed to search manifests: %w", err)
	}

	return s.parseSearchResults(searchResult)
}

func (s *Storage) Rollback(id string, targetVersion int) error {
	return instance.db.Update(func(txn *badger.Txn) error {
		key := genVersionedKey(id, targetVersion)
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("old version for manifest %s not found", id)
		}

		val, _ := item.ValueCopy(nil)
		newVersion, err := s.getCurrentVersion(txn, id)
		if err != nil {
			return err
		}

		newVersion += 1
		newKey := genVersionedKey(id, newVersion)

		if err = txn.Set(newKey, val); err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("old version for manifest %s not found", id)
		} else if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		latestKey := genLatestKey(id)
		return txn.Set(latestKey, newKey)
	})
}

func (s *Storage) CleanupOldVersions(id string, keep int) error {
	return instance.db.Update(func(txn *badger.Txn) error {
		versions, err := s.getAllVersions(txn, id)
		if err != nil {
			return err
		}

		if len(versions) <= keep {
			return nil
		}

		sort.Ints(versions)
		toDelete := versions[:len(versions)-keep]

		for _, ver := range toDelete {
			key := genVersionedKey(id, ver)

			if err = txn.Delete(key); err != nil {
				return fmt.Errorf("error deleting old versions: %w", err)
			}

			if err = s.index.Delete(string(key)); err != nil {
				return fmt.Errorf("failed to delete from index: %v", err)
			}
		}

		latestVer := versions[len(versions)-1]
		if latestVer > MaxManifestVersion {
			return s.resetVersions(txn, id)
		}

		return nil
	})
}

func (s *Storage) loadByHash(hash string) (manifests.Manifest, error) {
	var manifest manifests.Manifest

	hashQuery := bleve.NewTermQuery(hash)
	hashQuery.SetField(kinds.MetaHash)

	searchResult, err := s.index.Search(bleve.NewSearchRequest(hashQuery))
	if err != nil {
		return nil, fmt.Errorf("bleve search failed: %w", err)
	}

	if searchResult.Total == 0 {
		return nil, fmt.Errorf("manifest with hash %q not found", hash)
	}

	hit := searchResult.Hits[0]

	err = s.db.View(func(txn *badger.Txn) error {
		var latestItem *badger.Item

		latestKey := genLatestKey(extractBaseID(hit.ID))
		latestItem, err = txn.Get(latestKey)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		var versionedKey []byte
		versionedKey, err = latestItem.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to get versioned key: %w", err)
		}

		var manifestItem *badger.Item
		manifestItem, err = txn.Get(versionedKey)
		if err != nil {
			return fmt.Errorf("failed to get manifest data: %w", err)
		}

		return manifestItem.Value(func(data []byte) error {
			var parseErr error
			manifest, parseErr = operations.Parse(operations.JSONFormat, data)
			return parseErr
		})
	})
	if err != nil {
		return nil, fmt.Errorf("storage error: %w", err)
	}

	if manifest.GetMeta().GetHash() != hash {
		return nil, fmt.Errorf("hash mismatch after load")
	}

	return manifest, nil
}

func (s *Storage) loadVersion(id string, version int) (manifests.Manifest, error) {
	var m manifests.Manifest
	versionedKey := genVersionedKey(id, version)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(versionedKey)
		if err != nil {
			return err
		}
		return item.Value(func(data []byte) error {
			m, err = operations.Parse(operations.JSONFormat, data)
			return err
		})
	})

	return m, err
}

func (s *Storage) loadLatest() ([]manifests.Manifest, error) {
	var results []manifests.Manifest

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(manifestsLatestKey)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			versionedKey, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("failed to get versioned key for %s: %w", item.Key(), err)
			}

			versionedItem, err := txn.Get(versionedKey)
			if err != nil {
				return fmt.Errorf("failed to get manifest at %s: %w", versionedKey, err)
			}

			var m manifests.Manifest
			err = versionedItem.Value(func(data []byte) error {
				m, err = operations.Parse(operations.JSONFormat, data)
				if err != nil {
					return fmt.Errorf("failed to parse manifest %s: %w", versionedKey, err)
				}
				results = append(results, m)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load results: %w", err)
	}

	return results, nil
}

func (s *Storage) loadBulk(opts LoadOptions) ([]manifests.Manifest, error) {
	var (
		results []manifests.Manifest
		errs    error
	)

	err := instance.db.View(func(txn *badger.Txn) error {
		for _, id := range opts.IDs {
			namespace, kind, name, err := utils.ParseManifestIDWithError(id)
			if err != nil {
				return err
			}

			id = utils.FormManifestID(namespace, kind, name)

			item, err := txn.Get(genLatestKey(id))
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					errs = errors.Join(errs, fmt.Errorf("manifest %q not found", id))
					continue
				}
				errs = errors.Join(errs, fmt.Errorf("failed to get latest pointer for %q: %w", id, err))
				continue
			}

			var latestKey []byte
			if err = item.Value(func(val []byte) error {
				latestKey = append([]byte{}, val...)
				return nil
			}); err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to read latest key for %q: %w", id, err))
				continue
			}

			item, err = txn.Get(latestKey)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to get manifest data for %q: %w", id, err))
				continue
			}

			var manifest manifests.Manifest
			if err = item.Value(func(data []byte) error {
				manifest, err = operations.Parse(operations.JSONFormat, data)
				return err
			}); err != nil {
				errs = errors.Join(errs, fmt.Errorf("parsing failed for %q: %w", id, err))
				continue
			}

			results = append(results, manifest)
		}

		if len(results) == 0 && errs != nil {
			return fmt.Errorf("no manifests loaded: %w", errs)
		}
		return nil
	})
	if err != nil {
		return results, fmt.Errorf("transaction failed: %w", errors.Join(err, errs))
	}

	return results, errs
}

func (s *Storage) parseSearchResults(searchResults *bleve.SearchResult) ([]manifests.Manifest, error) {
	if searchResults == nil || searchResults.Total == 0 {
		return nil, nil
	}

	var (
		manifestsList = make([]manifests.Manifest, 0, len(searchResults.Hits))
		errorsList    []error
	)

	err := s.db.View(func(txn *badger.Txn) error {
		for _, hit := range searchResults.Hits {
			latestKey := genLatestKey(extractBaseID(hit.ID))
			item, err := txn.Get(latestKey)
			if err != nil {
				errorsList = append(errorsList, fmt.Errorf("latest key for %q: %w", hit.ID, err))
				continue
			}

			versionedKey, err := item.ValueCopy(nil)
			if err != nil {
				errorsList = append(errorsList, fmt.Errorf("value copy for %q: %w", hit.ID, err))
				continue
			}

			manifestItem, err := txn.Get(versionedKey)
			if err != nil {
				errorsList = append(errorsList, fmt.Errorf("manifest %q: %w", string(versionedKey), err))
				continue
			}

			var m manifests.Manifest
			err = manifestItem.Value(func(data []byte) error {
				m, err = operations.Parse(operations.JSONFormat, data)
				if err != nil {
					return err
				}
				manifestsList = append(manifestsList, m)
				return nil
			})
			if err != nil {
				errorsList = append(errorsList, fmt.Errorf("parse manifest %q: %w", string(versionedKey), err))
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return manifestsList, errors.Join(errorsList...)
}

func (s *Storage) getCurrentVersion(txn *badger.Txn, id string) (int, error) {
	item, err := txn.Get(genLatestKey(id))
	if errors.Is(err, badger.ErrKeyNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return 0, err
	}

	valStr := string(val)
	parts := strings.Split(valStr, "@v")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid versioned key format")
	}

	return strconv.Atoi(parts[1])
}

func (s *Storage) indexCurrentManifest(m manifests.Manifest, newKey string, oldVersion int) error {
	if oldVersion > 0 {
		oldKey := genVersionedKey(m.GetID(), oldVersion)
		if err := s.index.Delete(string(oldKey)); err != nil {
			return fmt.Errorf("could not delete manifest: %s old version %v: %v", m.GetID(), oldVersion, err)
		}
	}

	m.GetMeta().SetIsCurrent(true)

	if err := s.index.Index(newKey, m.Index()); err != nil {
		return fmt.Errorf("could not index manifest: %s old version %v: %v", m.GetID(), oldVersion, err)
	}

	return nil
}

func (s *Storage) cleanupOldVersions(txn *badger.Txn, id string, keepVersions int) error {
	versions, err := s.getAllVersions(txn, id)
	if err != nil {
		return err
	}

	sort.Ints(versions)

	for i := 0; i < len(versions)-keepVersions; i++ {
		key := genVersionedKey(id, versions[i])

		if err = txn.Delete(key); err != nil {
			return fmt.Errorf("could not delete manifest: %s old version %v: %v", id, versions[i], err)
		}

		if err = s.index.Delete(string(key)); err != nil {
			return fmt.Errorf("could not delete manifest: %s old version %v: %v", id, versions[i], err)
		}
	}

	if versions[len(versions)-1] > MaxManifestVersion {
		return s.resetVersions(txn, id)
	}

	return nil
}

func (s *Storage) resetVersions(txn *badger.Txn, id string) error {
	latestKey := genLatestKey(id)
	item, err := txn.Get(latestKey)
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	latestVal, err := item.ValueCopy(nil)
	if err != nil {
		return fmt.Errorf("failed to read latest value: %w", err)
	}

	item, err = txn.Get(latestVal)
	if err != nil {
		return fmt.Errorf("failed to get manifest data: %w", err)
	}

	manifestData, err := item.ValueCopy(nil)
	if err != nil {
		return fmt.Errorf("failed to copy manifest data: %w", err)
	}

	manifest, err := operations.Parse(operations.JSONFormat, manifestData)
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}

	meta := manifest.GetMeta()
	meta.SetVersion(1)
	meta.SetUpdatedAt(time.Now())
	meta.SetUpdatedBy("qube")
	meta.SetIsCurrent(true)

	newData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshaling failed: %w", err)
	}

	newKey := genVersionedKey(id, 1)
	if err = txn.Set(newKey, newData); err != nil {
		return fmt.Errorf("failed to save new version: %w", err)
	}

	if err = txn.Set(latestKey, newKey); err != nil {
		return fmt.Errorf("failed to update latest pointer: %w", err)
	}

	if err = s.cleanupIndexForReset(id, newKey); err != nil {
		return fmt.Errorf("index cleanup failed: %w", err)
	}

	return s.index.Index(string(newKey), manifest.Index())
}

func (s *Storage) cleanupIndexForReset(id string, newKey []byte) error {
	query := bleve.NewTermQuery(id)
	searchReq := bleve.NewSearchRequest(query)
	searchReq.Fields = []string{"*"}

	searchResults, err := s.index.Search(searchReq)
	if err != nil {
		return err
	}

	var delErrors error
	for _, hit := range searchResults.Hits {
		if hit.ID != string(newKey) {
			if err := s.index.Delete(hit.ID); err != nil {
				delErrors = errors.Join(delErrors,
					fmt.Errorf("failed to delete %s: %w", hit.ID, err))
			}
		}
	}
	return delErrors
}

func (s *Storage) getAllVersions(txn *badger.Txn, id string) ([]int, error) {
	var versions []int
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := []byte(id + "@v")
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		key := string(it.Item().Key())
		parts := strings.Split(key, "@v")
		if len(parts) != 2 {
			continue
		}
		ver, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		versions = append(versions, ver)
	}
	return versions, nil
}
