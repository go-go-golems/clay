package command

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/rs/zerolog/log"
)

// commandDocument represents the structure used for indexing commands in Bleve
type commandDocument struct {
	Name     string                 `json:"name"`
	FullPath string                 `json:"full_path"`
	Parents  []string               `json:"parents"`
	Type     string                 `json:"type"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

// newCommandDocument creates a new commandDocument from a CommandDescription
func newCommandDocument(cmd *cmds.CommandDescription) *commandDocument {
	fullPath := cmd.FullPath()

	log.Debug().
		Str("name", cmd.Name).
		Str("fullPath", fullPath).
		Strs("parents", cmd.Parents).
		Str("type", cmd.Type).
		Strs("tags", cmd.Tags).
		Msg("Creating document for command")

	return &commandDocument{
		Name:     cmd.Name,
		FullPath: fullPath,
		Parents:  cmd.Parents,
		Type:     cmd.Type,
		Tags:     cmd.Tags,
		Metadata: cmd.Metadata,
	}
}

// validate checks if the document has all required fields
func (d *commandDocument) validate() error {
	if d.Name == "" {
		return fmt.Errorf("command document %s must have a name", d.FullPath)
	}
	if d.Type == "" {
		return fmt.Errorf("command document %s (name: %s) must have a type", d.FullPath, d.Name)
	}
	return nil
}
