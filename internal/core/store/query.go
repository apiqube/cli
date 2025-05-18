package store

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

type Query interface {
	Execute(index bleve.Index) (*bleve.SearchResult, error)
	WithAll() Query
	WithExactName(name string) Query
	WithWildcardName(pattern string) Query
	WithRegexName(regex string) Query
	WithKind(kind string) Query
	WithNamespace(namespace string) Query
	WithVersion(version int) Query
	WithCreatedBy(by string) Query
	WithUsedBy(by string) Query
	WithHashPrefix(prefix string) Query
	WithDependencies(deps []string) Query
	WithAllDependencies(deps []string) Query
	WithCreatedAfter(t time.Time) Query
	WithCreatedBefore(t time.Time) Query
	WithUpdatedAfter(t time.Time) Query
	WithUpdatedBefore(t time.Time) Query
	WithLastApplied(t time.Time) Query
}

type manifestQuery struct {
	bleveQuery query.Query
	limit      int
}

func NewQuery() Query {
	return &manifestQuery{
		bleveQuery: bleve.NewMatchAllQuery(),
	}
}

func (q *manifestQuery) Execute(index bleve.Index) (*bleve.SearchResult, error) {
	searchRequest := bleve.NewSearchRequest(q.bleveQuery)
	if q.limit > 0 {
		searchRequest.Size = q.limit
	}

	return index.Search(searchRequest)
}

func (q *manifestQuery) WithAll() Query {
	q.bleveQuery = bleve.NewMatchAllQuery()
	return q
}

func (q *manifestQuery) WithExactName(name string) Query {
	termQuery := bleve.NewTermQuery(name)
	termQuery.SetField(index.Name)
	q.addQuery(termQuery)
	return q
}

func (q *manifestQuery) WithWildcardName(pattern string) Query {
	wildcardQuery := bleve.NewWildcardQuery(pattern)
	wildcardQuery.SetField(index.Name)
	q.addQuery(wildcardQuery)
	return q
}

func (q *manifestQuery) WithRegexName(regex string) Query {
	regexpQuery := bleve.NewRegexpQuery(regex)
	regexpQuery.SetField(index.Name)
	q.addQuery(regexpQuery)
	return q
}

func (q *manifestQuery) WithKind(kind string) Query {
	termQuery := bleve.NewTermQuery(kind)
	termQuery.SetField(index.Kind)
	q.addQuery(termQuery)
	return q
}

func (q *manifestQuery) WithNamespace(namespace string) Query {
	termQuery := bleve.NewTermQuery(namespace)
	termQuery.SetField(index.Namespace)
	q.addQuery(termQuery)
	return q
}

func (q *manifestQuery) WithVersion(version int) Query {
	val := float64(version)
	numericQuery := bleve.NewNumericRangeQuery(&val, nil)
	numericQuery.SetField(index.MetaVersion)
	return q
}

func (q *manifestQuery) WithCreatedBy(by string) Query {
	termQuery := bleve.NewTermQuery(by)
	termQuery.SetField(index.MetaCreatedBy)
	q.addQuery(termQuery)
	return q
}

func (q *manifestQuery) WithUsedBy(by string) Query {
	termQuery := bleve.NewTermQuery(by)
	termQuery.SetField(index.MetaUsedBy)
	q.addQuery(termQuery)
	return q
}

func (q *manifestQuery) WithHashPrefix(prefix string) Query {
	prefixQuery := bleve.NewPrefixQuery(prefix)
	prefixQuery.SetField(index.MetaHash)
	q.addQuery(prefixQuery)
	return q
}

func (q *manifestQuery) WithDependencies(deps []string) Query {
	disjunctionQuery := bleve.NewDisjunctionQuery()
	for _, dep := range deps {
		termQuery := bleve.NewTermQuery(dep)
		termQuery.SetField(index.DependsOn)
		disjunctionQuery.AddQuery(termQuery)
	}
	q.addQuery(disjunctionQuery)
	return q
}

func (q *manifestQuery) WithAllDependencies(deps []string) Query {
	conjunctionQuery := bleve.NewConjunctionQuery()
	for _, dep := range deps {
		termQuery := bleve.NewTermQuery(dep)
		termQuery.SetField(index.DependsOn)
		conjunctionQuery.AddQuery(termQuery)
	}
	q.addQuery(conjunctionQuery)
	return q
}

func (q *manifestQuery) WithCreatedAfter(t time.Time) Query {
	dateQuery := bleve.NewTermRangeQuery(t.Format(time.RFC3339Nano), "")
	dateQuery.SetField(index.MetaCreatedAt)
	q.addQuery(dateQuery)
	return q
}

func (q *manifestQuery) WithCreatedBefore(t time.Time) Query {
	dateQuery := bleve.NewTermRangeQuery("", t.Format(time.RFC3339Nano))
	dateQuery.SetField(index.MetaCreatedAt)
	q.addQuery(dateQuery)
	return q
}

func (q *manifestQuery) WithUpdatedAfter(t time.Time) Query {
	dateQuery := bleve.NewTermRangeQuery(t.Format(time.RFC3339Nano), "")
	dateQuery.SetField(index.MetaUpdatedAt)
	q.addQuery(dateQuery)
	return q
}

func (q *manifestQuery) WithUpdatedBefore(t time.Time) Query {
	dateQuery := bleve.NewTermRangeQuery("", t.Format(time.RFC3339Nano))
	dateQuery.SetField(index.MetaUpdatedAt)
	q.addQuery(dateQuery)
	return q
}

func (q *manifestQuery) WithLastApplied(t time.Time) Query {
	dateQuery := bleve.NewTermRangeQuery("", t.Format(time.RFC3339Nano))
	dateQuery.SetField(index.MetaLastApplied)
	q.addQuery(dateQuery)
	return q
}

func (q *manifestQuery) addQuery(newQuery query.Query) {
	if _, ok := q.bleveQuery.(*query.MatchAllQuery); ok {
		q.bleveQuery = newQuery
	} else {
		q.bleveQuery = bleve.NewConjunctionQuery(q.bleveQuery, newQuery)
	}
}
