package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apiqube/cli/ui"

	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/apiqube/cli/internal/core/manifests/parsing"
	bleceQuery "github.com/blevesearch/bleve/v2/search/query"

	"github.com/adrg/xdg"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v4"
)

const (
	StorageDirPath = "ApiQube/Storage"
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

func (s *Storage) SaveManifests(mans ...manifests.Manifest) error {
	var err error
	err = instance.db.Update(func(txn *badger.Txn) error {
		var data []byte

		for _, m := range mans {
			data, err = json.Marshal(m)
			if err != nil {
				return fmt.Errorf("error marshalling manifest: %v", err)
			}

			if err = txn.Set(genManifestKey(m.GetID()), data); err != nil {
				return err
			}

			if err = s.index.Index(m.GetID(), m.Index()); err != nil {
				ui.Errorf("Failed to index manifest %s: %v", m.GetID(), err)
				return err
			}
		}

		return nil
	})

	return err
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

			var man manifests.Manifest

			if err = item.Value(func(data []byte) error {
				if man, err = parsing.ParseManifestAsJSON(data); err == nil {
					results = append(results, man)
				}

				return err
			}); err != nil {
				rErr = errors.Join(rErr, err)
			}
		}

		return nil
	})

	return results, errors.Join(rErr, err)
}

func (s *Storage) LoadManifest(id string) (manifests.Manifest, error) {
	var result manifests.Manifest
	var rErr error

	err := instance.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(genManifestKey(id))
		if errors.Is(err, badger.ErrKeyNotFound) {
			rErr = errors.Join(rErr, fmt.Errorf("manifest %s not found", id))
			return nil
		} else if err != nil {
			rErr = errors.Join(rErr, err)
		}

		if err = item.Value(func(data []byte) error {
			if result, err = parsing.ParseManifestAsJSON(data); err != nil {
				return err
			}

			return nil
		}); err != nil {
			rErr = errors.Join(rErr, err)
		}

		return nil
	})

	return result, errors.Join(rErr, err)
}

func (s *Storage) FindManifestsByKind(kind string) ([]manifests.Manifest, error) {
	query := bleve.NewTermQuery(kind)
	query.SetField(index.Kind)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error searching manifests: %v", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestsByVersion(lowVersion, heightVersion uint8) ([]manifests.Manifest, error) {
	low, height := float64(lowVersion), float64(heightVersion)

	query := bleve.NewNumericRangeQuery(&low, &height)
	query.SetField(index.Version)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error searching manifests: %v", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestByName(name string) (manifests.Manifest, error) {
	query := bleve.NewMatchPhraseQuery(name)
	query.SetField(index.Name)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 1

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	results, err := s.parseSearResults(searchResults)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no matching manifest found")
	} else {
		return results[0], nil
	}
}

func (s *Storage) FindManifestsByNameWildcard(namePattern string) ([]manifests.Manifest, error) {
	query := bleve.NewWildcardQuery(fmt.Sprintf("*%s*", namePattern))
	query.SetField(index.Name)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestsByNamespace(namespace string) ([]manifests.Manifest, error) {
	query := bleve.NewTermQuery(namespace)
	query.SetField(index.Namespace)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestByDependencies(dependencies []string, requireAll bool) ([]manifests.Manifest, error) {
	var query bleceQuery.Query

	if requireAll {
		q := bleve.NewConjunctionQuery()
		for _, dep := range dependencies {
			termQuery := bleve.NewTermQuery(dep)
			termQuery.SetField(index.DependsOn)
			q.AddQuery(termQuery)
		}
		query = q
	} else {
		q := bleve.NewDisjunctionQuery()
		for _, dep := range dependencies {
			termQuery := bleve.NewTermQuery(dep)
			termQuery.SetField(index.DependsOn)
			q.AddQuery(termQuery)
		}
		query = q
	}

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestByHash(hash string) (manifests.Manifest, error) {
	query := bleve.NewTermQuery(hash)
	query.SetField(index.MetaHash)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 1

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	results, err := s.parseSearResults(searchResults)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no matching manifest found")
	} else {
		return results[0], nil
	}
}

func (s *Storage) FindManifestsByCreatedAtRange(start, end time.Time) ([]manifests.Manifest, error) {
	query := bleve.NewDateRangeQuery(start, end)
	query.SetField(index.MetaCreatedAt)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error searching manifests: %v", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestsByCreatedBy(createdBy string) ([]manifests.Manifest, error) {
	query := bleve.NewTermQuery(createdBy)
	query.SetField(index.MetaCreatedBy)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestsByUpdatedAtRange(start, end time.Time) ([]manifests.Manifest, error) {
	query := bleve.NewDateRangeQuery(start, end)
	query.SetField(index.MetaUpdatedAt)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error searching manifests: %v", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestsByUpdatedBy(updatedBy string) ([]manifests.Manifest, error) {
	query := bleve.NewTermQuery(updatedBy)
	query.SetField(index.MetaUpdatedBy)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestsByUsedBy(usedBy string) ([]manifests.Manifest, error) {
	query := bleve.NewTermQuery(usedBy)
	query.SetField(index.MetaUsedBy)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) FindManifestsByLastAppliedRange(start, end time.Time) ([]manifests.Manifest, error) {
	query := bleve.NewDateRangeQuery(start, end)
	query.SetField(index.MetaLastApplied)

	searchRequest := bleve.NewSearchRequest(query)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error searching manifests: %v", err)
	}

	return s.parseSearResults(searchResults)
}

func (s *Storage) parseSearResults(searchResults *bleve.SearchResult) ([]manifests.Manifest, error) {
	var results []manifests.Manifest
	var rErr error

	if searchResults.Total > 0 {
		for _, hit := range searchResults.Hits {
			var man manifests.Manifest

			err := s.db.View(func(txn *badger.Txn) error {
				var item *badger.Item
				var err error

				item, err = txn.Get(genManifestKey(hit.ID))
				if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
					rErr = errors.Join(rErr, fmt.Errorf("could not find manifest by ID: %v", hit.ID))
				} else if err != nil {
					rErr = errors.Join(rErr, fmt.Errorf("failed to load manifest by ID: %v", hit.ID))
				}

				return item.Value(func(val []byte) error {
					man, err = parsing.ParseManifest(parsing.JSONMethod, val)
					if err != nil {
						return err
					}
					return nil
				})
			})
			if err != nil {
				rErr = errors.Join(rErr, err)
			}

			results = append(results, man)
		}
	}

	return results, rErr
}
