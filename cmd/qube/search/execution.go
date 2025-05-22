package search

import (
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/store"
)

func executeSearch(opts *Options) ([]manifests.Manifest, error) {
	if opts.all {
		return store.Load(store.LoadOptions{All: true})
	}

	query := buildSearchQuery(opts)
	return store.Search(query)
}

func buildSearchQuery(opts *Options) store.Query {
	query := store.NewQuery()

	if opts.flagsSet["name"] {
		query.WithExactName(opts.name)
	} else if opts.flagsSet["name-wildcard"] {
		query.WithWildcardName(opts.nameWildcard)
	} else if opts.flagsSet["name-regex"] {
		query.WithRegexName(opts.nameRegex)
	}

	if opts.flagsSet["namespace"] {
		query.WithNamespace(opts.namespace)
	}

	if opts.flagsSet["kind"] {
		query.WithKind(opts.kind)
	}

	if opts.flagsSet["version"] {
		query.WithVersion(opts.version)
	}

	if opts.flagsSet["created-by"] {
		query.WithCreatedBy(opts.createdBy)
	}

	if opts.flagsSet["user-by"] {
		query.WithUsedBy(opts.usedBy)
	}

	if opts.flagsSet["hash"] {
		query.WithHashPrefix(opts.hashPrefix)
	}

	if opts.flagsSet["depends"] {
		query.WithDependencies(opts.dependsOn)
	} else if opts.flagsSet["depends-all"] {
		query.WithAllDependencies(opts.dependsOnAll)
	}

	if opts.flagsSet["created-after"] {
		query.WithCreatedAfter(opts.createdAfter)
	}

	if opts.flagsSet["created-before"] {
		query.WithCreatedBefore(opts.createdBefore)
	}

	if opts.flagsSet["updated-after"] {
		query.WithUpdatedAfter(opts.updatedAfter)
	}

	if opts.flagsSet["updated-before"] {
		query.WithUpdatedBefore(opts.updatedBefore)
	}

	if opts.flagsSet["last-applied"] {
		query.WithLastApplied(opts.lastApplied)
	}

	return query
}
