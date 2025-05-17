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

	kindMapping := bleve.NewTextFieldMapping()
	kindMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(index.Kind, kindMapping)

	nameMapping := bleve.NewTextFieldMapping()
	manifestMapping.AddFieldMappingsAt(index.Name, nameMapping)

	dependsMapping := bleve.NewTextFieldMapping()
	dependsMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(index.DependsOn, dependsMapping)

	namespaceMapping := bleve.NewTextFieldMapping()
	namespaceMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt(index.Namespace, namespaceMapping)

	// Meta
	manifestMapping.AddFieldMappingsAt(index.MetaHash, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaCreatedAt, bleve.NewDateTimeFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaCreatedBy, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaUpdatedAt, bleve.NewDateTimeFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaUpdatedBy, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaUsedBy, bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt(index.MetaLastApplied, bleve.NewDateTimeFieldMapping())

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = manifestMapping

	return indexMapping
}
