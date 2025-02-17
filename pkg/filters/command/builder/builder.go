package builder

import (
	"fmt"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

// Builder provides methods for creating command filters
type Builder struct {
	opts *Options
}

// New creates a new Builder with the given options
func New(opts ...Option) *Builder {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return &Builder{opts: options}
}

// Type creates a filter that matches commands by type
func (b *Builder) Type(type_ string) *FilterBuilder {
	q := bleve.NewTermQuery(type_)
	q.SetField("type")
	return NewFilterBuilder(q, b.opts)
}

// Types creates a filter that matches commands by any of the given types
func (b *Builder) Types(types ...string) *FilterBuilder {
	queries := make([]query.Query, len(types))
	for i, t := range types {
		q := bleve.NewTermQuery(t)
		q.SetField("type")
		queries[i] = q
	}
	return NewFilterBuilder(
		query.NewDisjunctionQuery(queries),
		b.opts,
	)
}

// Tag creates a filter that matches commands having a specific tag
func (b *Builder) Tag(tag string) *FilterBuilder {
	q := bleve.NewTermQuery(tag)
	q.SetField("tags")
	return NewFilterBuilder(q, b.opts)
}

// Tags creates a filter that matches commands having any of the given tags
func (b *Builder) Tags(tags ...string) *FilterBuilder {
	queries := make([]query.Query, len(tags))
	for i, tag := range tags {
		q := bleve.NewTermQuery(tag)
		q.SetField("tags")
		queries[i] = q
	}
	return NewFilterBuilder(
		query.NewDisjunctionQuery(queries),
		b.opts,
	)
}

// AllTags creates a filter that matches commands having all the given tags
func (b *Builder) AllTags(tags ...string) *FilterBuilder {
	queries := make([]query.Query, len(tags))
	for i, tag := range tags {
		q := bleve.NewTermQuery(tag)
		q.SetField("tags")
		queries[i] = q
	}
	return NewFilterBuilder(
		query.NewConjunctionQuery(queries),
		b.opts,
	)
}

// AnyTags is an alias for Tags
func (b *Builder) AnyTags(tags ...string) *FilterBuilder {
	return b.Tags(tags...)
}

// Path creates a filter that matches commands by exact path
func (b *Builder) Path(path string) *FilterBuilder {
	q := bleve.NewTermQuery(path)
	q.SetField("full_path")
	return NewFilterBuilder(q, b.opts)
}

// PathGlob creates a filter that matches commands by path glob pattern
func (b *Builder) PathGlob(pattern string) *FilterBuilder {
	// Convert glob pattern to a prefix query where possible
	if matches, _ := filepath.Match(pattern, ""); matches {
		return b.PathPrefix(pattern)
	}
	q := bleve.NewWildcardQuery(pattern)
	q.SetField("full_path")
	return NewFilterBuilder(q, b.opts)
}

// PathPrefix creates a filter that matches commands by path prefix
func (b *Builder) PathPrefix(prefix string) *FilterBuilder {
	q := bleve.NewPrefixQuery(prefix)
	q.SetField("full_path")
	return NewFilterBuilder(q, b.opts)
}

// Name creates a filter that matches commands by exact name
func (b *Builder) Name(name string) *FilterBuilder {
	q := bleve.NewTermQuery(name)
	q.SetField("name")
	return NewFilterBuilder(q, b.opts)
}

// NamePattern creates a filter that matches commands by name pattern
func (b *Builder) NamePattern(pattern string) *FilterBuilder {
	q := bleve.NewWildcardQuery(pattern)
	q.SetField("name_pattern")
	return NewFilterBuilder(q, b.opts)
}

// Metadata creates a filter that matches commands by metadata field value
func (b *Builder) Metadata(key string, value interface{}) *FilterBuilder {
	q := bleve.NewTermQuery(fmt.Sprintf("%v", value))
	q.SetField("metadata." + key)
	return NewFilterBuilder(q, b.opts)
}

// MetadataMatch creates a filter that matches commands by multiple metadata fields
func (b *Builder) MetadataMatch(matches map[string]interface{}) *FilterBuilder {
	queries := make([]query.Query, 0, len(matches))
	for key, value := range matches {
		q := bleve.NewTermQuery(fmt.Sprintf("%v", value))
		q.SetField("metadata." + key)
		queries = append(queries, q)
	}
	return NewFilterBuilder(
		query.NewConjunctionQuery(queries),
		b.opts,
	)
}
