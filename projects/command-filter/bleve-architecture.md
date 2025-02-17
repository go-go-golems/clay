# Bleve-Based Command Filter Architecture

## Overview

A command filtering system using Bleve with an in-memory index. The index is created once and can be used for multiple searches.

## Core Components

### 1. Command Index

```go
// CommandIndex manages the in-memory Bleve index
type CommandIndex struct {
    index bleve.Index
}

// NewCommandIndex creates a new index from a list of commands
func NewCommandIndex(commands []*CommandDescription) (*CommandIndex, error) {
    // Create memory-only index
    indexMapping := bleve.NewIndexMapping()
    index, err := bleve.NewMemOnly(indexMapping)
    if err != nil {
        return nil, err
    }

    // Index all commands
    for _, cmd := range commands {
        doc := &commandDocument{
            Name:        cmd.Name,
            NamePattern: cmd.Name,  // For pattern matching
            FullPath:    cmd.FullPath(),
            Parents:     cmd.Parents,
            Type:        cmd.Type,
            Tags:        cmd.Tags,
            Metadata:    cmd.Metadata,
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
func (ci *CommandIndex) Search(ctx context.Context, filter *BleveFilter, commands []*CommandDescription) ([]*CommandDescription, error) {
    searchRequest := bleve.NewSearchRequest(filter.query)
    searchRequest.Size = len(commands) // Get all matches
    
    searchResult, err := ci.index.SearchInContext(ctx, searchRequest)
    if err != nil {
        return nil, err
    }

    // Collect matching commands
    matches := make([]*CommandDescription, 0, len(searchResult.Hits))
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
```

### 2. Document Structure

```go
// Document structure for indexing
type commandDocument struct {
    Name        string                 `json:"name"`
    NamePattern string                 `json:"name_pattern"` // For pattern matching
    FullPath    string                 `json:"full_path"`
    Parents     []string               `json:"parents"`
    Type        string                 `json:"type"`
    Tags        []string               `json:"tags"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

### 3. Filter Types

```go
// BleveFilter wraps a Bleve query
type BleveFilter struct {
    query bleve.Query
}

// QueryBuilder provides methods for creating filters
type QueryBuilder struct{}

// Name-based filters
func (b *QueryBuilder) ExactName(name string) *BleveFilter {
    return &BleveFilter{
        query: bleve.NewTermQuery(name),
    }
}

// ... other filter methods remain the same ...
```

## Usage Examples

### Basic Usage

```go
func ExampleBasicUsage(commands []*CommandDescription) {
    // Create index
    index, err := NewCommandIndex(commands)
    if err != nil {
        panic(err)
    }
    defer index.Close()

    // Create builder
    builder := &QueryBuilder{}

    // Create filter
    filter := builder.And(
        builder.Type("http"),
        builder.HasTag("api"),
    )

    // Search
    ctx := context.Background()
    matches, err := index.Search(ctx, filter, commands)
    if err != nil {
        panic(err)
    }

    for _, cmd := range matches {
        fmt.Printf("Found command: %s\n", cmd.Name)
    }
}
```

### Multiple Searches

```go
func ExampleMultipleSearches(commands []*CommandDescription) {
    // Create index once
    index, err := NewCommandIndex(commands)
    if err != nil {
        panic(err)
    }
    defer index.Close()

    builder := &QueryBuilder{}
    ctx := context.Background()

    // First search: HTTP commands
    httpFilter := builder.Type("http")
    httpCmds, err := index.Search(ctx, httpFilter, commands)
    if err != nil {
        panic(err)
    }

    // Second search: GRPC commands
    grpcFilter := builder.Type("grpc")
    grpcCmds, err := index.Search(ctx, grpcFilter, commands)
    if err != nil {
        panic(err)
    }

    // Third search: Commands with specific tag
    tagFilter := builder.HasTag("experimental")
    taggedCmds, err := index.Search(ctx, tagFilter, commands)
    if err != nil {
        panic(err)
    }

    // Process results...
}
```

### Complex Filtering

```go
func ExampleComplexFiltering(commands []*CommandDescription) {
    // Create index
    index, err := NewCommandIndex(commands)
    if err != nil {
        panic(err)
    }
    defer index.Close()

    builder := &QueryBuilder{}
    ctx := context.Background()

    // Build complex filter
    filter := builder.And(
        // Must be HTTP or GRPC
        builder.Or(
            builder.Type("http"),
            builder.Type("grpc"),
        ),
        // Must have all these tags
        builder.HasAllTags("api", "v2", "stable"),
        // Must be in specific path
        builder.ParentsGlob("service/*/api"),
        // Must have specific metadata
        builder.MetadataField("version", "2.0.0"),
    )

    // Search
    matches, err := index.Search(ctx, filter, commands)
    if err != nil {
        panic(err)
    }

    for _, cmd := range matches {
        fmt.Printf("Found command: %s\n", cmd.FullPath())
    }
}
```

## Future Enhancements

2. **Features**
   - Fuzzy matching
   - Regular expressions
   - Custom scoring
   - Sorting options

3. **Advanced**
   - Query string parsing
   - Faceted search
