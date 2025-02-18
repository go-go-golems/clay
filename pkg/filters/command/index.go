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

	// Create field mappings
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = "keyword" // Use keyword analyzer to prevent tokenization
	fmt.Printf("Created keyword field mapping with analyzer: %s\n", keywordFieldMapping.Analyzer)

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "standard"
	fmt.Printf("Created text field mapping with analyzer: %s\n", textFieldMapping.Analyzer)

	// Create document mapping
	documentMapping := bleve.NewDocumentMapping()

	// Add field mappings
	documentMapping.AddFieldMappingsAt("name", keywordFieldMapping)
	documentMapping.AddFieldMappingsAt("name", textFieldMapping)
	fmt.Printf("Added field mapping for name\n")

	documentMapping.AddFieldMappingsAt("full_path", keywordFieldMapping)
	fmt.Printf("Added field mapping for full_path\n")

	documentMapping.AddFieldMappingsAt("type", keywordFieldMapping)
	fmt.Printf("Added field mapping for type\n")

	documentMapping.AddFieldMappingsAt("tags", keywordFieldMapping)
	fmt.Printf("Added field mapping for tags\n")

	documentMapping.AddFieldMappingsAt("parents", keywordFieldMapping)
	fmt.Printf("Added field mapping for parents\n")

	// Create metadata mapping
	metadataMapping := bleve.NewDocumentMapping()
	metadataMapping.Dynamic = true // Allow dynamic fields in metadata
	metadataMapping.Enabled = true
	documentMapping.AddSubDocumentMapping("metadata", metadataMapping)
	fmt.Printf("Added sub-document mapping for metadata\n")

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
