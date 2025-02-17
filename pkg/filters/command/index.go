package command

import (
	"context"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-go-golems/clay/pkg/filters/command/builder"
	"github.com/go-go-golems/glazed/pkg/cmds"
)

// CommandIndex manages the in-memory Bleve index for command filtering
type CommandIndex struct {
	index bleve.Index
}

// NewCommandIndex creates a new index from a list of commands
func NewCommandIndex(commands []*cmds.CommandDescription) (*CommandIndex, error) {
	// Create memory-only index
	indexMapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return nil, err
	}

	// Index all commands
	for _, cmd := range commands {
		doc := newCommandDocument(cmd)
		if err := doc.validate(); err != nil {
			index.Close()
			return nil, err
		}
		if err := index.Index(cmd.Name, doc); err != nil {
			index.Close()
			return nil, err
		}
	}

	return &CommandIndex{index: index}, nil
}

// Close releases the index resources
func (ci *CommandIndex) Close() error {
	return ci.index.Close()
}

// Search executes a query and returns matching commands
func (ci *CommandIndex) Search(ctx context.Context, filter *builder.FilterBuilder, commands []*cmds.CommandDescription) ([]*cmds.CommandDescription, error) {
	searchRequest := bleve.NewSearchRequest(filter.Build())
	searchRequest.Size = len(commands) // Get all matches

	searchResult, err := ci.index.SearchInContext(ctx, searchRequest)
	if err != nil {
		return nil, err
	}

	// Collect matching commands
	matches := make([]*cmds.CommandDescription, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		for _, cmd := range commands {
			if cmd.Name == hit.ID {
				matches = append(matches, cmd)
				break
			}
		}
	}

	return matches, nil
}
