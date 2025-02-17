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
