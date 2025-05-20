package store

import (
	"github.com/apiqube/cli/internal/core/manifests/kinds"
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
	manifestMapping.AddFieldMappingsAt(kinds.ID, idMapping)

	manifestMapping.AddFieldMappingsAt(kinds.MetaVersion, bleve.NewNumericFieldMapping())

	exactMatchFields := []string{
		kinds.Version,
		kinds.Kind,
		kinds.DependsOn,
		kinds.MetaCreatedBy,
		kinds.MetaUpdatedBy,
		kinds.MetaUsedBy,
	}

	for _, field := range exactMatchFields {
		fm := bleve.NewTextFieldMapping()
		fm.Analyzer = "keyword"
		manifestMapping.AddFieldMappingsAt(field, fm)
	}

	nameMapping := bleve.NewTextFieldMapping()
	nameMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(kinds.Name, nameMapping)

	namespaceMapping := bleve.NewTextFieldMapping()
	namespaceMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(kinds.Namespace, namespaceMapping)

	hashMapping := bleve.NewTextFieldMapping()
	hashMapping.Analyzer = "keyword"
	hashMapping.Store = true
	manifestMapping.AddFieldMappingsAt(kinds.MetaHash, hashMapping)

	dateTimeFieldMapping := bleve.NewTextFieldMapping()
	dateTimeFieldMapping.Analyzer = "keyword"
	dateTimeFieldMapping.Store = true
	dateTimeFieldMapping.IncludeTermVectors = false
	dateTimeFieldMapping.IncludeInAll = false

	dateFields := []string{
		kinds.MetaCreatedAt,
		kinds.MetaUpdatedAt,
		kinds.MetaLastApplied,
	}

	for _, field := range dateFields {
		manifestMapping.AddFieldMappingsAt(field, dateTimeFieldMapping)
	}

	indexMapping.DefaultMapping = manifestMapping

	return indexMapping
}
