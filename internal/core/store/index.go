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
		index.Namespace,
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
	nameMapping.Analyzer = "standard"
	manifestMapping.AddFieldMappingsAt(index.Name, nameMapping)

	hashMapping := bleve.NewTextFieldMapping()
	hashMapping.Analyzer = "keyword"
	hashMapping.Store = true
	manifestMapping.AddFieldMappingsAt(index.MetaHash, hashMapping)

	dateTimeMapping := bleve.NewDateTimeFieldMapping()

	dateFields := []string{
		index.MetaCreatedAt,
		index.MetaUpdatedAt,
		index.MetaLastApplied,
	}

	for _, field := range dateFields {
		manifestMapping.AddFieldMappingsAt(field, dateTimeMapping)
	}

	indexMapping.DefaultMapping = manifestMapping

	return indexMapping
}
