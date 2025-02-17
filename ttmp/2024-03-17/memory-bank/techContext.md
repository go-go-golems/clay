# Command Filter Technical Context

## Technology Stack

1. **Core Technologies**
   - Go 1.23
   - Bleve v2 (search backend)
   - Clay command system
   - Glazed package

2. **Dependencies**
   ```go
   require (
       "github.com/blevesearch/bleve/v2"
       "github.com/go-go-golems/glazed/pkg/cmds"
   )
   ```

## Technical Architecture

1. **Package Structure**
   ```
   pkg/filters/command/
   ├── builder/
   │   ├── builder.go    # Main builder interface
   │   ├── filter.go     # Filter builder implementation
   │   └── options.go    # Builder options
   ├── index.go          # Command index implementation
   ├── document.go       # Document structure
   └── filter.go         # Filter types (legacy)
   ```

2. **Core Components**
   - `CommandIndex`: Manages in-memory Bleve index
   - `Builder`: Fluent interface for query construction
   - `FilterBuilder`: Filter combination and building
   - `commandDocument`: Document structure for indexing

3. **Query Building**
   ```go
   builder := command.NewBuilder()
   filter := builder.
       Type("http").
       Tag("api").
       Build()
   ```

## Technical Constraints

1. **Memory Usage**
   - In-memory index only
   - No persistence layer
   - Optimize for large command sets

2. **Performance**
   - Fast search response times
   - Efficient boolean operations
   - Concurrent search support

3. **Compatibility**
   - Go 1.23 compatibility
   - Clay command system integration
   - Backward compatibility during migration

## Development Setup

1. **Environment**
   - Go 1.23 or later
   - Clay repository
   - Glazed package

2. **Build Process**
   ```shell
   go build ./...
   go test ./...
   ```

3. **Testing**
   - Unit tests for all components
   - Integration tests for search
   - Performance benchmarks

## Technical Decisions

1. **Bleve Backend**
   - Mature search library
   - Rich query capabilities
   - Good performance
   - Active maintenance

2. **Builder Pattern**
   - Fluent interface for better DX
   - Method chaining for readability
   - Easy to extend
   - Type-safe operations

3. **In-Memory Index**
   - Fast operations
   - No persistence needed
   - Simple implementation
   - Lower resource usage

## Core Technologies

### Bleve Search Engine
1. Index Configuration
   - In-memory index for command filtering
   - Custom field mappings for specialized text analysis
   - Keyword analyzer for path fields
   - Support for boolean query combinations

2. Query Types
   - Term queries for exact matching
   - Prefix queries for path prefixes
   - Wildcard queries for glob patterns
   - Boolean queries for combinations

3. Text Analysis
   - Keyword analyzer: No tokenization, exact matching
   - Default analyzer: Standard tokenization
   - Custom analyzers available when needed

### Go Libraries
1. Core Dependencies
   - `github.com/blevesearch/bleve/v2`: Search engine
   - `github.com/go-go-golems/glazed`: Command framework
   - Standard library: `path/filepath`, `strings`

2. Testing Framework
   - `github.com/stretchr/testify/assert`
   - `github.com/stretchr/testify/require`
   - Integration testing support

## Development Setup

### Project Structure
```
clay/
  pkg/
    filters/
      command/
        builder/     # Query builder
        index.go     # Search index
        document.go  # Document schema
        filter.go    # Filter types
```

### Build Requirements
- Go 1.21 or later
- Bleve v2.x
- Testify for testing

### Development Tools
1. Testing
   - `go test` for unit tests
   - Integration tests for query verification
   - Debug logging for query inspection

2. Code Organization
   - Builder pattern for queries
   - Clear separation of concerns
   - Type-safe interfaces

## Technical Constraints

### Search Limitations
1. Query Performance
   - Wildcard queries can be expensive
   - No built-in query caching
   - Full result set loading

2. Path Handling
   - Platform-specific separators
   - Special character limitations
   - Deep hierarchy performance

3. Memory Usage
   - In-memory index size
   - Result set memory requirements
   - Document storage overhead

### Best Practices
1. Query Construction
   - Use specific field queries when possible
   - Combine queries with AND/OR operations
   - Validate input patterns

2. Index Management
   - Configure proper analyzers
   - Use appropriate field mappings
   - Monitor index size

3. Error Handling
   - Validate queries before execution
   - Handle search errors gracefully
   - Provide meaningful error messages

## Dependencies

### Required Packages
```go
import (
    "github.com/blevesearch/bleve/v2"
    "github.com/blevesearch/bleve/v2/search/query"
    "github.com/go-go-golems/glazed/pkg/cmds"
)
```

### Optional Tools
- Debug logging utilities
- Performance monitoring
- Query analysis tools 