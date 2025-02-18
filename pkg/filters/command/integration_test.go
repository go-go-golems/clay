package command

import (
	"context"
	"testing"

	"github.com/go-go-golems/clay/pkg/filters/command/builder"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexQueries_TypeAndTag(t *testing.T) {
	// Create test commands
	commands := []*cmds.CommandDescription{
		{
			Name: "http-api",
			Type: "http",
			Tags: []string{"api", "stable"},
		},
		{
			Name: "grpc-api",
			Type: "grpc",
			Tags: []string{"api", "stable"},
		},
		{
			Name: "http-web",
			Type: "http",
			Tags: []string{"web"},
		},
		{
			Name: "grpc-internal",
			Type: "grpc",
			Tags: []string{"internal"},
		},
	}

	// Create index
	index, err := NewCommandIndex(commands)
	require.NoError(t, err)
	defer index.Close()

	ctx := context.Background()
	b := builder.New()

	tests := []struct {
		name          string
		buildFilter   func(*builder.Builder) *builder.FilterBuilder
		expectedNames []string
	}{
		{
			name: "type AND tag",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Type("http").And(b.Tag("api"))
			},
			expectedNames: []string{"http-api"},
		},
		{
			name: "multiple types AND multiple tags",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Types("http", "grpc").And(b.AllTags("api", "stable"))
			},
			expectedNames: []string{"http-api", "grpc-api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.buildFilter(b)
			results, err := index.Search(ctx, filter, commands)
			require.NoError(t, err)

			resultNames := make([]string, len(results))
			for i, cmd := range results {
				resultNames[i] = cmd.Name
			}
			assert.ElementsMatch(t, tt.expectedNames, resultNames)
		})
	}
}

func TestComplexQueries_PathBased(t *testing.T) {
	// Create test commands
	commands := []*cmds.CommandDescription{
		{
			Name:    "http-api",
			Type:    "http",
			Tags:    []string{"stable"},
			Parents: []string{"service", "api"},
		},
		{
			Name:    "grpc-api",
			Type:    "grpc",
			Tags:    []string{"stable"},
			Parents: []string{"service", "api"},
		},
		{
			Name:    "http-web",
			Type:    "http",
			Parents: []string{"service", "web"},
		},
		{
			Name:    "internal-tool",
			Type:    "cli",
			Parents: []string{"tools", "internal"},
		},
	}

	// Create index
	index, err := NewCommandIndex(commands)
	require.NoError(t, err)
	defer index.Close()

	ctx := context.Background()
	b := builder.New()

	tests := []struct {
		name          string
		buildFilter   func(*builder.Builder) *builder.FilterBuilder
		expectedNames []string
	}{
		{
			name: "path prefix AND type",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.PathPrefix("service/").And(b.Type("http"))
			},
			expectedNames: []string{"http-api", "http-web"},
		},
		{
			name: "path glob AND tag",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.PathGlob("*/api/*").And(b.Tag("stable"))
			},
			expectedNames: []string{"http-api", "grpc-api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.buildFilter(b)
			results, err := index.Search(ctx, filter, commands)
			require.NoError(t, err)

			resultNames := make([]string, len(results))
			for i, cmd := range results {
				resultNames[i] = cmd.Name
			}
			assert.ElementsMatch(t, tt.expectedNames, resultNames)
		})
	}
}

func TestComplexQueries_Metadata(t *testing.T) {
	// Create test commands
	commands := []*cmds.CommandDescription{
		{
			Name: "http-api",
			Type: "http",
			Tags: []string{"api", "stable"},
			Metadata: map[string]interface{}{
				"version": "2.0.0",
				"stage":   "prod",
			},
		},
		{
			Name: "grpc-api",
			Type: "grpc",
			Tags: []string{"api", "stable"},
			Metadata: map[string]interface{}{
				"version": "1.0.0",
				"stage":   "prod",
			},
		},
		{
			Name: "http-web",
			Type: "http",
			Tags: []string{"web"},
			Metadata: map[string]interface{}{
				"version": "2.0.0",
				"stage":   "dev",
			},
		},
	}

	// Create index
	index, err := NewCommandIndex(commands)
	require.NoError(t, err)
	defer index.Close()

	ctx := context.Background()
	b := builder.New()

	tests := []struct {
		name          string
		buildFilter   func(*builder.Builder) *builder.FilterBuilder
		expectedNames []string
	}{
		{
			name: "metadata AND tag",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Metadata("version", "2.0.0").And(b.Tag("stable"))
			},
			expectedNames: []string{"http-api"},
		},
		{
			name: "metadata match AND type",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.MetadataMatch(map[string]interface{}{
					"version": "2.0.0",
					"stage":   "prod",
				}).And(b.Type("http"))
			},
			expectedNames: []string{"http-api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.buildFilter(b)
			results, err := index.Search(ctx, filter, commands)
			require.NoError(t, err)

			resultNames := make([]string, len(results))
			for i, cmd := range results {
				resultNames[i] = cmd.Name
			}
			assert.ElementsMatch(t, tt.expectedNames, resultNames)
		})
	}
}

func TestComplexQueries_NamePattern(t *testing.T) {
	// Create test commands
	commands := []*cmds.CommandDescription{
		{
			Name: "serve-api",
			Type: "http",
			Tags: []string{"api"},
		},
		{
			Name: "serve-web",
			Type: "http",
			Tags: []string{"web"},
		},
		{
			Name: "api-server",
			Type: "api",
			Tags: []string{"api"},
		},
		{
			Name: "web-server",
			Type: "api",
			Tags: []string{"web"},
		},
		{
			Name: "process-data",
			Type: "worker",
			Tags: []string{"background"},
		},
	}

	// Create index
	index, err := NewCommandIndex(commands)
	require.NoError(t, err)
	defer index.Close()

	ctx := context.Background()
	b := builder.New()

	tests := []struct {
		name          string
		buildFilter   func(*builder.Builder) *builder.FilterBuilder
		expectedNames []string
	}{
		{
			name: "name pattern AND type",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.NamePattern("serve*").And(b.Type("http"))
			},
			expectedNames: []string{"serve-api", "serve-web"},
		},
		{
			name: "name OR name",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Name("api-server").Or(b.Name("web-server"))
			},
			expectedNames: []string{"api-server", "web-server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.buildFilter(b)
			results, err := index.Search(ctx, filter, commands)
			require.NoError(t, err)

			resultNames := make([]string, len(results))
			for i, cmd := range results {
				resultNames[i] = cmd.Name
			}
			assert.ElementsMatch(t, tt.expectedNames, resultNames)
		})
	}
}

func TestComplexQueries_NestedCombinations(t *testing.T) {
	// Create test commands
	commands := []*cmds.CommandDescription{
		{
			Name: "prod-api",
			Type: "http",
			Tags: []string{"api", "stable"},
			Metadata: map[string]interface{}{
				"version": "2.0.0",
				"stage":   "prod",
			},
		},
		{
			Name: "dev-api",
			Type: "http",
			Tags: []string{"api", "experimental"},
			Metadata: map[string]interface{}{
				"version": "2.1.0",
				"stage":   "dev",
			},
		},
		{
			Name: "prod-web",
			Type: "http",
			Tags: []string{"web", "stable"},
			Metadata: map[string]interface{}{
				"version": "2.0.0",
				"stage":   "prod",
			},
		},
		{
			Name: "test-service",
			Type: "test",
			Tags: []string{"test", "experimental"},
			Metadata: map[string]interface{}{
				"version": "1.0.0",
				"stage":   "test",
			},
		},
	}

	// Create index
	index, err := NewCommandIndex(commands)
	require.NoError(t, err)
	defer index.Close()

	ctx := context.Background()
	b := builder.New()

	tests := []struct {
		name          string
		buildFilter   func(*builder.Builder) *builder.FilterBuilder
		expectedNames []string
	}{
		{
			name: "complex nested - type AND (tag OR tag) AND metadata",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				tagFilter := b.Tag("api").Or(b.Tag("web"))
				return b.Type("http").And(tagFilter).And(
					b.MetadataMatch(map[string]interface{}{
						"version": "2.0.0",
						"stage":   "prod",
					}),
				)
			},
			expectedNames: []string{"prod-api", "prod-web"},
		},
		{
			name: "NOT combination - type AND NOT tag",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Type("http").And(b.Tag("experimental").Not())
			},
			expectedNames: []string{"prod-api", "prod-web"},
		},
		{
			name: "multi-level combination",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				httpFilter := b.Type("http").And(b.Tag("api")).And(b.Metadata("stage", "prod"))
				testFilter := b.Type("test").And(b.Tag("experimental")).And(b.Metadata("stage", "test"))
				return httpFilter.Or(testFilter)
			},
			expectedNames: []string{"prod-api", "test-service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.buildFilter(b)
			results, err := index.Search(ctx, filter, commands)
			require.NoError(t, err)

			resultNames := make([]string, len(results))
			for i, cmd := range results {
				resultNames[i] = cmd.Name
			}
			assert.ElementsMatch(t, tt.expectedNames, resultNames)
		})
	}
}
