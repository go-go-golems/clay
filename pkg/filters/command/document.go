package command

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
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
	fmt.Printf("Creating document for command %s:\n", cmd.Name)
	fmt.Printf("  Full path: %s\n", fullPath)
	fmt.Printf("  Parents: %v\n", cmd.Parents)
	fmt.Printf("  Type: %s\n", cmd.Type)
	fmt.Printf("  Tags: %v\n", cmd.Tags)

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
		return fmt.Errorf("command document must have a name")
	}
	if d.Type == "" {
		return fmt.Errorf("command document must have a type")
	}
	return nil
}
