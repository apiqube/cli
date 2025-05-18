package store

import (
	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

func buildBleveMapping() *mapping.IndexMappingImpl {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultAnalyzer = "standard"

	manifestMapping := bleve.NewDocumentMapping()

	idMapping := bleve.NewTextFieldMapping()
	idMapping.Analyzer = "keyword"
	idMapping.Store = true
	manifestMapping.AddFieldMappingsAt(index.ID, idMapping)

	manifestMapping.AddFieldMappingsAt(index.Version, bleve.NewNumericFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaVersion, bleve.NewNumericFieldMapping())

	exactMatchFields := []string{
		index.Kind,
		index.DependsOn,
		index.MetaCreatedBy,
		index.MetaUpdatedBy,
		index.MetaUsedBy,
	}

	for _, field := range exactMatchFields {
		fm := bleve.NewTextFieldMapping()
		fm.Analyzer = "keyword"
		manifestMapping.AddFieldMappingsAt(field, fm)
	}

	nameMapping := bleve.NewTextFieldMapping()
	nameMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(index.Name, nameMapping)

	namespaceMapping := bleve.NewTextFieldMapping()
	namespaceMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(index.Namespace, namespaceMapping)

	hashMapping := bleve.NewTextFieldMapping()
	hashMapping.Analyzer = "keyword"
	hashMapping.Store = true
	manifestMapping.AddFieldMappingsAt(index.MetaHash, hashMapping)

	dateTimeFieldMapping := bleve.NewTextFieldMapping()
	dateTimeFieldMapping.Analyzer = "keyword"
	dateTimeFieldMapping.Store = true
	dateTimeFieldMapping.IncludeTermVectors = false
	dateTimeFieldMapping.IncludeInAll = false

	dateFields := []string{
		index.MetaCreatedAt,
		index.MetaUpdatedAt,
		index.MetaLastApplied,
	}

	for _, field := range dateFields {
		manifestMapping.AddFieldMappingsAt(field, dateTimeFieldMapping)
	}

	indexMapping.DefaultMapping = manifestMapping

	return indexMapping
}
