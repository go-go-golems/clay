package command

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandDocument_Conversion(t *testing.T) {
	tests := []struct {
		name     string
		input    *cmds.CommandDescription
		wantErr  bool
		validate func(*testing.T, *commandDocument)
	}{
		{
			name: "basic command conversion",
			input: &cmds.CommandDescription{
				Name:    "test-cmd",
				Type:    "test",
				Parents: []string{"parent1", "parent2"},
				Tags:    []string{"tag1", "tag2"},
				Metadata: map[string]interface{}{
					"version": "1.0.0",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, doc *commandDocument) {
				assert.Equal(t, "test-cmd", doc.Name)
				assert.Equal(t, "test-cmd", doc.NamePattern)
				assert.Equal(t, "test", doc.Type)
				assert.Equal(t, []string{"parent1", "parent2"}, doc.Parents)
				assert.Equal(t, []string{"tag1", "tag2"}, doc.Tags)
				assert.Equal(t, "1.0.0", doc.Metadata["version"])
			},
		},
		{
			name: "command with empty fields",
			input: &cmds.CommandDescription{
				Name: "test-cmd",
				Type: "test",
			},
			wantErr: false,
			validate: func(t *testing.T, doc *commandDocument) {
				assert.Equal(t, "test-cmd", doc.Name)
				assert.Equal(t, "test", doc.Type)
				assert.Empty(t, doc.Parents)
				assert.Empty(t, doc.Tags)
				assert.Empty(t, doc.Metadata)
			},
		},
		{
			name: "command with empty name",
			input: &cmds.CommandDescription{
				Type: "test",
			},
			wantErr: true,
		},
		{
			name: "command with empty type",
			input: &cmds.CommandDescription{
				Name: "test-cmd",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := newCommandDocument(tt.input)
			err := doc.validate()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, doc)
			}
		})
	}
}

func TestCommandDocument_Validation(t *testing.T) {
	tests := []struct {
		name    string
		doc     *commandDocument
		wantErr bool
	}{
		{
			name: "valid document",
			doc: &commandDocument{
				Name: "test-cmd",
				Type: "test",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			doc: &commandDocument{
				Type: "test",
			},
			wantErr: true,
		},
		{
			name: "missing type",
			doc: &commandDocument{
				Name: "test-cmd",
			},
			wantErr: true,
		},
		{
			name:    "missing both name and type",
			doc:     &commandDocument{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
