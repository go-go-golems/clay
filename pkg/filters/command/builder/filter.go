package builder

import (
	"fmt"

	"github.com/blevesearch/bleve/v2/search/query"
)

// FilterBuilder provides methods for building and combining filters
type FilterBuilder struct {
	query query.Query
	opts  *Options
}

// NewFilterBuilder creates a new FilterBuilder with the given query and options
func NewFilterBuilder(q query.Query, opts *Options) *FilterBuilder {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &FilterBuilder{
		query: q,
		opts:  opts,
	}
}

// And combines this filter with others using AND logic
func (f *FilterBuilder) And(others ...*FilterBuilder) *FilterBuilder {
	queries := make([]query.Query, len(others)+1)
	queries[0] = f.query
	for i, other := range others {
		queries[i+1] = other.query
	}
	fmt.Printf("Creating conjunction query with %d queries\n", len(queries))
	return NewFilterBuilder(
		query.NewConjunctionQuery(queries),
		f.opts,
	)
}

// Or combines this filter with others using OR logic
func (f *FilterBuilder) Or(others ...*FilterBuilder) *FilterBuilder {
	queries := make([]query.Query, len(others)+1)
	queries[0] = f.query
	for i, other := range others {
		queries[i+1] = other.query
	}
	fmt.Printf("Creating disjunction query with %d queries\n", len(queries))
	return NewFilterBuilder(
		query.NewDisjunctionQuery(queries),
		f.opts,
	)
}

// Not negates this filter
func (f *FilterBuilder) Not() *FilterBuilder {
	mustNotQueries := []query.Query{f.query}
	return NewFilterBuilder(
		query.NewBooleanQuery(nil, nil, mustNotQueries),
		f.opts,
	)
}

// Build returns the underlying Bleve query
func (f *FilterBuilder) Build() query.Query {
	return f.query
}

// Must is a helper that panics if err is not nil
func Must(filter *FilterBuilder, err error) *FilterBuilder {
	if err != nil {
		panic(err)
	}
	return filter
}

// NewFilter creates a new FilterBuilder from a raw query
func NewFilter(q query.Query) *FilterBuilder {
	return NewFilterBuilder(q, nil)
}
