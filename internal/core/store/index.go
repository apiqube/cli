package store

import (
	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

func buildBleveMapping() *mapping.IndexMappingImpl {
	manifestMapping := bleve.NewDocumentMapping()

	// Main
	manifestMapping.AddFieldMappingsAt(index.Version, bleve.NewNumericFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.Kind, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.Name, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.Namespace, bleve.NewTextFieldMapping())
	dependsMapping := bleve.NewTextFieldMapping()
	dependsMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(index.DependsOn, dependsMapping)

	// Meta
	manifestMapping.AddFieldMappingsAt(index.MetaHash, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaCreatedAt, bleve.NewDateTimeFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaCreatedBy, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaUpdatedAt, bleve.NewDateTimeFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaUpdatedBy, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaUsedBy, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaLastApplied, bleve.NewDateTimeFieldMapping())

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("manifest", manifestMapping)

	return indexMapping
}
