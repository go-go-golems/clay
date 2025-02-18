package command

import (
	"github.com/blevesearch/bleve/v2/search/query"
)

// BleveFilter wraps a Bleve query for command filtering
type BleveFilter struct {
	query query.Query
}

// NewBleveFilter creates a new BleveFilter with the given query
func NewBleveFilter(q query.Query) *BleveFilter {
	return &BleveFilter{query: q}
}

// And combines multiple filters with AND logic
func (f *BleveFilter) And(filters ...*BleveFilter) *BleveFilter {
	queries := make([]query.Query, len(filters)+1)
	queries[0] = f.query
	for i, filter := range filters {
		queries[i+1] = filter.query
	}
	return NewBleveFilter(query.NewConjunctionQuery(queries))
}

// Or combines multiple filters with OR logic
func (f *BleveFilter) Or(filters ...*BleveFilter) *BleveFilter {
	queries := make([]query.Query, len(filters)+1)
	queries[0] = f.query
	for i, filter := range filters {
		queries[i+1] = filter.query
	}
	return NewBleveFilter(query.NewDisjunctionQuery(queries))
}

// Not negates the filter
func (f *BleveFilter) Not() *BleveFilter {
	mustNotQueries := []query.Query{f.query}
	return NewBleveFilter(query.NewBooleanQuery(nil, nil, mustNotQueries))
}

// GetQuery returns the underlying Bleve query
func (f *BleveFilter) GetQuery() query.Query {
	return f.query
}
