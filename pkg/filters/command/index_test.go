package command

import (
	"context"
	"testing"

	"github.com/go-go-golems/clay/pkg/filters/command/builder"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandIndex_SimpleSearch(t *testing.T) {
	// Create test commands
	commands := []*cmds.CommandDescription{
		{
			Name: "http-server",
			Type: "http",
			Tags: []string{"api", "server"},
			Metadata: map[string]interface{}{
				"version": "1.0.0",
			},
		},
		{
			Name: "grpc-server",
			Type: "grpc",
			Tags: []string{"api", "server"},
			Metadata: map[string]interface{}{
				"version": "1.0.0",
			},
		},
		{
			Name: "cli-tool",
			Type: "cli",
			Tags: []string{"tool"},
			Metadata: map[string]interface{}{
				"version": "2.0.0",
			},
		},
	}

	// Create index
	index, err := NewCommandIndex(commands)
	require.NoError(t, err)
	defer index.Close()

	// Test cases
	tests := []struct {
		name          string
		buildFilter   func(*builder.Builder) *builder.FilterBuilder
		expectedNames []string
		expectedCount int
		errorExpected bool
	}{
		{
			name: "search by type",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Type("http")
			},
			expectedNames: []string{"http-server"},
			expectedCount: 1,
		},
		{
			name: "search by tag",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Tag("api")
			},
			expectedNames: []string{"http-server", "grpc-server"},
			expectedCount: 2,
		},
		{
			name: "search by metadata",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Metadata("version", "2.0.0")
			},
			expectedNames: []string{"cli-tool"},
			expectedCount: 1,
		},
		{
			name: "search with multiple tags",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.AllTags("api", "server")
			},
			expectedNames: []string{"http-server", "grpc-server"},
			expectedCount: 2,
		},
		{
			name: "search with type and tag combination",
			buildFilter: func(b *builder.Builder) *builder.FilterBuilder {
				return b.Type("http").And(b.Tag("api"))
			},
			expectedNames: []string{"http-server"},
			expectedCount: 1,
		},
	}

	ctx := context.Background()
	b := builder.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.buildFilter(b)
			results, err := index.Search(ctx, filter, commands)

			if tt.errorExpected {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)

			resultNames := make([]string, len(results))
			for i, cmd := range results {
				resultNames[i] = cmd.Name
			}
			assert.ElementsMatch(t, tt.expectedNames, resultNames)
		})
	}
}

// TestCommandIndex_Creation tests the creation and closing of the index
func TestCommandIndex_Creation(t *testing.T) {
	commands := []*cmds.CommandDescription{
		{
			Name: "test-cmd",
			Type: "test",
		},
	}

	// Test successful creation
	index, err := NewCommandIndex(commands)
	require.NoError(t, err)
	require.NotNil(t, index)

	// Test successful closing
	err = index.Close()
	require.NoError(t, err)

	// Test creation with nil commands
	_, err = NewCommandIndex(nil)
	require.NoError(t, err)

	// Test creation with empty commands
	_, err = NewCommandIndex([]*cmds.CommandDescription{})
	require.NoError(t, err)
}
