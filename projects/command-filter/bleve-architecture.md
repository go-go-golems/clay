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
func (ci *CommandIndex) Search(ctx context.Context, filter *FilterBuilder, commands []*CommandDescription) ([]*CommandDescription, error) {
    searchRequest := bleve.NewSearchRequest(filter.Build())
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

### 3. Query Builder Interface

```go
// QueryBuilder provides methods for creating filters
type QueryBuilder interface {
    // Type filters
    Type(type_ string) *FilterBuilder
    Types(types ...string) *FilterBuilder
    
    // Tag filters
    Tag(tag string) *FilterBuilder
    Tags(tags ...string) *FilterBuilder
    AllTags(tags ...string) *FilterBuilder
    AnyTags(tags ...string) *FilterBuilder
    
    // Path filters
    Path(path string) *FilterBuilder
    PathGlob(pattern string) *FilterBuilder
    PathPrefix(prefix string) *FilterBuilder
    
    // Name filters
    Name(name string) *FilterBuilder
    NamePattern(pattern string) *FilterBuilder
    
    // Metadata filters
    Metadata(key string, value interface{}) *FilterBuilder
    MetadataMatch(matches map[string]interface{}) *FilterBuilder
}

// FilterBuilder provides methods for combining filters
type FilterBuilder struct {
    query query.Query
}

func (f *FilterBuilder) And(others ...*FilterBuilder) *FilterBuilder
func (f *FilterBuilder) Or(others ...*FilterBuilder) *FilterBuilder
func (f *FilterBuilder) Not() *FilterBuilder
func (f *FilterBuilder) Build() query.Query
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
    builder := command.NewBuilder()

    // Create filter
    filter := builder.
        Type("http").
        Tag("api").
        Build()

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

    builder := command.NewBuilder()
    ctx := context.Background()

    // First search: HTTP commands
    httpFilter := builder.Type("http").Build()
    httpCmds, err := index.Search(ctx, httpFilter, commands)
    if err != nil {
        panic(err)
    }

    // Second search: GRPC commands with tags
    grpcFilter := builder.
        Type("grpc").
        AllTags("api", "stable").
        Build()
    grpcCmds, err := index.Search(ctx, grpcFilter, commands)
    if err != nil {
        panic(err)
    }

    // Third search: Commands with specific path and metadata
    complexFilter := builder.
        PathGlob("service/*/api").
        Metadata("version", "2.0.0").
        Build()
    complexCmds, err := index.Search(ctx, complexFilter, commands)
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

    builder := command.NewBuilder()
    ctx := context.Background()

    // Build complex filter
    filter := builder.Or(
        builder.Type("http"),
        builder.Type("grpc"),
    ).And(
        builder.AllTags("api", "v2", "stable"),
        builder.PathGlob("service/*/api"),
        builder.Metadata("version", "2.0.0"),
    ).Build()

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

### Using Helper Functions

```go
func ExampleHelperFunctions(commands []*CommandDescription) {
    builder := command.NewBuilder()

    // Using Must helper for error handling
    filter := Must(
        builder.Type("http").
        And(
            builder.Tags("api", "stable"),
            builder.PathPrefix("service"),
        ),
        nil,
    ).Build()

    // Using NewFilter for raw queries
    customFilter := NewFilter(
        bleve.NewTermQuery("custom_field"),
    ).Build()
}
```

## Future Enhancements

1. **Performance**
   - Query caching
   - Result caching
   - Batch indexing

2. **Features**
   - Fuzzy matching
   - Regular expressions
   - Custom scoring
   - Sorting options

3. **Advanced**
   - Query string parsing
   - Faceted search
   - Custom analyzers
