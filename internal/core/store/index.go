package store

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

func buildBleveMapping() *mapping.IndexMappingImpl {
	manifestMapping := bleve.NewDocumentMapping()

	// Main
	manifestMapping.AddFieldMappingsAt("version", bleve.NewNumericFieldMapping())
	manifestMapping.AddFieldMappingsAt("kind", bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt("name", bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt("namespace", bleve.NewTextFieldMapping())
	dependsMapping := bleve.NewTextFieldMapping()
	dependsMapping.Analyzer = "keyword"
	manifestMapping.AddFieldMappingsAt("dependsOn", dependsMapping)

	// Meta
	manifestMapping.AddFieldMappingsAt("hash", bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt("createdAt", bleve.NewDateTimeFieldMapping())
	manifestMapping.AddFieldMappingsAt("createdBy", bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt("updatedAt", bleve.NewDateTimeFieldMapping())
	manifestMapping.AddFieldMappingsAt("usedBy", bleve.NewTextFieldMapping())
	manifestMapping.AddFieldMappingsAt("lastApplied", bleve.NewDateTimeFieldMapping())

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("manifest", manifestMapping)

	return indexMapping
}
