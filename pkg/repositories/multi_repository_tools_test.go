package repositories

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-go-golems/clay/pkg/repositories/mcp"
	"github.com/stretchr/testify/assert"
)

func TestListTools(t *testing.T) {
	tests := []struct {
		name  string
		repos map[string]struct {
			tools []mcp.Tool
			err   error
		}
		expectedTools []string // tool names
		wantErr       bool
	}{
		{
			name: "root mounted tools - no prefix",
			repos: map[string]struct {
				tools []mcp.Tool
				err   error
			}{
				"/": {
					tools: []mcp.Tool{
						{Name: "tool1", Description: "test tool 1"},
						{Name: "tool2", Description: "test tool 2"},
					},
				},
			},
			expectedTools: []string{"tool1", "tool2"}, // No leading slash for root mount
			wantErr:       false,
		},
		{
			name: "mounted repo tools - with prefix",
			repos: map[string]struct {
				tools []mcp.Tool
				err   error
			}{
				"/test": {
					tools: []mcp.Tool{
						{Name: "tool1", Description: "test tool 1"},
					},
				},
			},
			expectedTools: []string{"/test/tool1"}, // Keep prefix for non-root mounts
			wantErr:       false,
		},
		{
			name: "multiple repos",
			repos: map[string]struct {
				tools []mcp.Tool
				err   error
			}{
				"/test1": {
					tools: []mcp.Tool{
						{Name: "tool1", Description: "test tool 1"},
					},
				},
				"/test2": {
					tools: []mcp.Tool{
						{Name: "tool2", Description: "test tool 2"},
					},
				},
			},
			expectedTools: []string{"/test1/tool1", "/test2/tool2"},
			wantErr:       false,
		},
		{
			name: "repo with error",
			repos: map[string]struct {
				tools []mcp.Tool
				err   error
			}{
				"/": {
					tools: nil,
					err:   assert.AnError,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()
			ctx := context.Background()

			for path, repo := range tt.repos {
				mockRepo := NewMockRepository(nil)
				mockRepo.tools = repo.tools
				mockRepo.toolsError = repo.err
				mr.Mount(path, mockRepo)
			}

			tools, _, err := mr.ListTools(ctx, "")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			var toolNames []string
			for _, tool := range tools {
				toolNames = append(toolNames, tool.Name)
			}
			assert.ElementsMatch(t, tt.expectedTools, toolNames)
		})
	}
}

func TestToolSchemaHandling(t *testing.T) {
	// Create a command with a complex schema
	cmd := createTestCommand("test", nil)
	desc := cmd.Description()
	desc.Short = "Test command"
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"param1": map[string]interface{}{
				"type": "string",
			},
		},
	}
	schemaBytes, _ := json.Marshal(schema)

	tests := []struct {
		name     string
		tools    []mcp.Tool
		expected string // expected schema in tool
	}{
		{
			name: "preserve schema",
			tools: []mcp.Tool{
				{
					Name:        "test",
					Description: "Test command",
					InputSchema: schemaBytes,
				},
			},
			expected: string(schemaBytes),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()
			ctx := context.Background()

			mockRepo := NewMockRepository(nil)
			mockRepo.tools = tt.tools
			mr.Mount("/", mockRepo)

			tools, _, err := mr.ListTools(ctx, "")
			assert.NoError(t, err)
			assert.Len(t, tools, 1)
			assert.Equal(t, tt.expected, string(tools[0].InputSchema))
		})
	}
}
