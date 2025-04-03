package command

import (
	"context"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-go-golems/clay/pkg/filters/command/builder"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/rs/zerolog/log"
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
	log.Debug().Str("analyzer", keywordFieldMapping.Analyzer).Msg("Created keyword field mapping")

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "standard"
	log.Debug().Str("analyzer", textFieldMapping.Analyzer).Msg("Created text field mapping")

	// Create document mapping
	documentMapping := bleve.NewDocumentMapping()

	// Add field mappings
	documentMapping.AddFieldMappingsAt("name", keywordFieldMapping)
	documentMapping.AddFieldMappingsAt("name", textFieldMapping)
	log.Debug().Msg("Added field mapping for name")

	documentMapping.AddFieldMappingsAt("full_path", keywordFieldMapping)
	log.Debug().Msg("Added field mapping for full_path")

	documentMapping.AddFieldMappingsAt("type", keywordFieldMapping)
	log.Debug().Msg("Added field mapping for type")

	documentMapping.AddFieldMappingsAt("tags", keywordFieldMapping)
	log.Debug().Msg("Added field mapping for tags")

	documentMapping.AddFieldMappingsAt("parents", keywordFieldMapping)
	log.Debug().Msg("Added field mapping for parents")

	// Create metadata mapping
	metadataMapping := bleve.NewDocumentMapping()
	metadataMapping.Dynamic = true // Allow dynamic fields in metadata
	metadataMapping.Enabled = true
	documentMapping.AddSubDocumentMapping("metadata", metadataMapping)
	log.Debug().Msg("Added sub-document mapping for metadata")

	// Add document mapping to index
	indexMapping.AddDocumentMapping("_default", documentMapping)
	log.Debug().Msg("Added document mapping to index mapping")

	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return nil, err
	}

	// Index all commands
	for _, cmd := range commands {
		doc := newCommandDocument(cmd)
		if err := doc.validate(); err != nil {
			if closeErr := index.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Error closing index after validation failure")
			}
			return nil, err
		}
		if err := index.Index(cmd.Name, doc); err != nil {
			if closeErr := index.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Error closing index after indexing failure")
			}
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
	log.Debug().Type("queryType", query).Msg("Executing search with query")

	// Print the query details
	log.Debug().Interface("query", query).Msg("Query details")

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = len(commands) // Get all matches

	searchResult, err := ci.index.SearchInContext(ctx, searchRequest)
	if err != nil {
		return nil, err
	}

	log.Debug().Int("hitCount", len(searchResult.Hits)).Msg("Search returned hits")
	for _, hit := range searchResult.Hits {
		log.Debug().Str("id", hit.ID).Float64("score", hit.Score).Msg("Hit found")
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
