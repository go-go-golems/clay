package builder

import (
	"testing"

	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_TypeFilters(t *testing.T) {
	b := New()

	tests := []struct {
		name      string
		filter    *FilterBuilder
		wantType  string
		wantField string
	}{
		{
			name:      "single type filter",
			filter:    b.Type("http"),
			wantType:  "term",
			wantField: "type",
		},
		{
			name:      "multiple types filter",
			filter:    b.Types("http", "grpc"),
			wantType:  "disjunction",
			wantField: "type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.filter.Build()
			require.NotNil(t, q)

			switch tt.wantType {
			case "term":
				termQuery, ok := q.(*query.TermQuery)
				require.True(t, ok, "expected TermQuery")
				assert.Equal(t, tt.wantField, termQuery.Field())
			case "disjunction":
				disjQuery, ok := q.(*query.DisjunctionQuery)
				require.True(t, ok, "expected DisjunctionQuery")
				for _, sq := range disjQuery.Disjuncts {
					termQuery, ok := sq.(*query.TermQuery)
					require.True(t, ok, "expected TermQuery in Disjuncts")
					assert.Equal(t, tt.wantField, termQuery.Field())
				}
			}
		})
	}
}

func TestBuilder_TagFilters(t *testing.T) {
	b := New()

	tests := []struct {
		name      string
		filter    *FilterBuilder
		wantType  string
		wantField string
	}{
		{
			name:      "single tag filter",
			filter:    b.Tag("api"),
			wantType:  "term",
			wantField: "tags",
		},
		{
			name:      "any tags filter",
			filter:    b.Tags("api", "stable"),
			wantType:  "disjunction",
			wantField: "tags",
		},
		{
			name:      "all tags filter",
			filter:    b.AllTags("api", "stable"),
			wantType:  "conjunction",
			wantField: "tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.filter.Build()
			require.NotNil(t, q)

			switch tt.wantType {
			case "term":
				termQuery, ok := q.(*query.TermQuery)
				require.True(t, ok, "expected TermQuery")
				assert.Equal(t, tt.wantField, termQuery.Field())
			case "disjunction":
				disjQuery, ok := q.(*query.DisjunctionQuery)
				require.True(t, ok, "expected DisjunctionQuery")
				for _, sq := range disjQuery.Disjuncts {
					termQuery, ok := sq.(*query.TermQuery)
					require.True(t, ok, "expected TermQuery in Disjuncts")
					assert.Equal(t, tt.wantField, termQuery.Field())
				}
			case "conjunction":
				conjQuery, ok := q.(*query.ConjunctionQuery)
				require.True(t, ok, "expected ConjunctionQuery")
				for _, sq := range conjQuery.Conjuncts {
					termQuery, ok := sq.(*query.TermQuery)
					require.True(t, ok, "expected TermQuery in Conjuncts")
					assert.Equal(t, tt.wantField, termQuery.Field())
				}
			}
		})
	}
}

func TestBuilder_PathFilters(t *testing.T) {
	b := New()

	tests := []struct {
		name      string
		filter    *FilterBuilder
		wantType  string
		wantField string
	}{
		{
			name:      "exact path filter",
			filter:    b.Path("service/api"),
			wantType:  "term",
			wantField: "full_path",
		},
		{
			name:      "path prefix filter",
			filter:    b.PathPrefix("service/"),
			wantType:  "prefix",
			wantField: "full_path",
		},
		{
			name:      "path glob filter",
			filter:    b.PathGlob("service/*/api"),
			wantType:  "wildcard",
			wantField: "full_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.filter.Build()
			require.NotNil(t, q)

			switch tt.wantType {
			case "term":
				termQuery, ok := q.(*query.TermQuery)
				require.True(t, ok, "expected TermQuery")
				assert.Equal(t, tt.wantField, termQuery.Field())
			case "prefix":
				prefixQuery, ok := q.(*query.PrefixQuery)
				require.True(t, ok, "expected PrefixQuery")
				assert.Equal(t, tt.wantField, prefixQuery.Field())
			case "wildcard":
				wildcardQuery, ok := q.(*query.WildcardQuery)
				require.True(t, ok, "expected WildcardQuery")
				assert.Equal(t, tt.wantField, wildcardQuery.Field())
			}
		})
	}
}

func TestBuilder_NameFilters(t *testing.T) {
	b := New()

	tests := []struct {
		name      string
		filter    *FilterBuilder
		wantType  string
		wantField string
	}{
		{
			name:      "exact name filter",
			filter:    b.Name("api-server"),
			wantType:  "term",
			wantField: "name",
		},
		{
			name:      "name pattern filter",
			filter:    b.NamePattern("serve*"),
			wantType:  "wildcard",
			wantField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.filter.Build()
			require.NotNil(t, q)

			switch tt.wantType {
			case "term":
				termQuery, ok := q.(*query.TermQuery)
				require.True(t, ok, "expected TermQuery")
				assert.Equal(t, tt.wantField, termQuery.Field())
			case "wildcard":
				wildcardQuery, ok := q.(*query.WildcardQuery)
				require.True(t, ok, "expected WildcardQuery")
				assert.Equal(t, tt.wantField, wildcardQuery.Field())
			}
		})
	}
}

func TestBuilder_MetadataFilters(t *testing.T) {
	b := New()

	tests := []struct {
		name      string
		filter    *FilterBuilder
		wantType  string
		wantField string
	}{
		{
			name:      "single metadata filter",
			filter:    b.Metadata("version", "2.0.0"),
			wantType:  "term",
			wantField: "metadata.version",
		},
		{
			name: "metadata match filter",
			filter: b.MetadataMatch(map[string]interface{}{
				"version": "2.0.0",
				"stage":   "prod",
			}),
			wantType:  "conjunction",
			wantField: "metadata.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.filter.Build()
			require.NotNil(t, q)

			switch tt.wantType {
			case "term":
				termQuery, ok := q.(*query.TermQuery)
				require.True(t, ok, "expected TermQuery")
				assert.Equal(t, tt.wantField, termQuery.Field())
			case "conjunction":
				conjQuery, ok := q.(*query.ConjunctionQuery)
				require.True(t, ok, "expected ConjunctionQuery")
				for _, sq := range conjQuery.Conjuncts {
					termQuery, ok := sq.(*query.TermQuery)
					require.True(t, ok, "expected TermQuery in Conjuncts")
					field := termQuery.Field()
					assert.True(t, len(field) > len(tt.wantField), "field should start with metadata.")
					assert.True(t, field[:len(tt.wantField)] == tt.wantField, "field should start with metadata.")
				}
			}
		})
	}
}

func TestBuilder_FilterCombinations(t *testing.T) {
	b := New()

	tests := []struct {
		name     string
		filter   *FilterBuilder
		wantType string
	}{
		{
			name: "AND combination",
			filter: b.Type("http").And(
				b.Tag("api"),
			),
			wantType: "conjunction",
		},
		{
			name: "OR combination",
			filter: b.Type("http").Or(
				b.Type("grpc"),
			),
			wantType: "disjunction",
		},
		{
			name: "NOT combination",
			filter: b.Type("http").And(
				b.Tag("deprecated").Not(),
			),
			wantType: "conjunction",
		},
		{
			name: "complex nested combination",
			filter: b.Type("http").And(
				b.Tag("api").Or(
					b.Tag("web"),
				),
			).And(
				b.MetadataMatch(map[string]interface{}{
					"version": "2.0.0",
					"stage":   "prod",
				}),
			),
			wantType: "conjunction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.filter.Build()
			require.NotNil(t, q)

			switch tt.wantType {
			case "conjunction":
				_, ok := q.(*query.ConjunctionQuery)
				require.True(t, ok, "expected ConjunctionQuery")
			case "disjunction":
				_, ok := q.(*query.DisjunctionQuery)
				require.True(t, ok, "expected DisjunctionQuery")
			}
		})
	}
}
