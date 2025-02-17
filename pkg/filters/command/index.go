package command

import (
	"context"
	"fmt"

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
	// Create memory-only index with custom mapping
	indexMapping := bleve.NewIndexMapping()

	// Create a keyword field mapping for full_path
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = "keyword" // Use keyword analyzer to prevent tokenization
	fmt.Printf("Created keyword field mapping for full_path with analyzer: %s\n", keywordFieldMapping.Analyzer)

	// Create document mapping
	documentMapping := bleve.NewDocumentMapping()
	documentMapping.AddFieldMappingsAt("full_path", keywordFieldMapping)
	fmt.Printf("Added field mapping for full_path to document mapping\n")

	// Add document mapping to index
	indexMapping.AddDocumentMapping("_default", documentMapping)
	fmt.Printf("Added document mapping to index mapping\n")

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
	query := filter.Build()
	fmt.Printf("Executing search with query type: %T\n", query)

	// Print the query details
	fmt.Printf("Query details: %#v\n", query)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = len(commands) // Get all matches

	searchResult, err := ci.index.SearchInContext(ctx, searchRequest)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Search returned %d hits\n", len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		fmt.Printf("Hit: %s (score: %f)\n", hit.ID, hit.Score)
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
