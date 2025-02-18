# Command Filter Technical Context

## Technology Stack

1. **Core Technologies**
   - Go 1.23
   - Bleve v2 (search backend)
   - Clay command system
   - Glazed package
   - Zerolog (structured logging)

2. **Dependencies**
   ```go
   require (
       "github.com/blevesearch/bleve/v2"
       "github.com/go-go-golems/glazed/pkg/cmds"
       "github.com/rs/zerolog/log"
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
   - `zerolog`: Structured logging throughout components

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
   - Monitor logging memory impact

2. **Performance**
   - Fast search response times
   - Efficient boolean operations
   - Concurrent search support
   - Logging overhead consideration

3. **Compatibility**
   - Go 1.23 compatibility
   - Clay command system integration
   - Backward compatibility during migration
   - Zerolog integration requirements

## Development Setup

1. **Environment**
   - Go 1.23 or later
   - Clay repository
   - Glazed package
   - Zerolog for logging

2. **Build Process**
   ```shell
   go build ./...
   go test ./...
   ```

3. **Testing**
   - Unit tests for all components
   - Integration tests for search
   - Performance benchmarks
   - Logging verification

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

4. **Structured Logging**
   - Zerolog for performance
   - Debug-level operation tracing
   - Structured fields for filtering
   - Consistent logging patterns

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

# Technical Context

## Core Technologies

1. Bleve Search Engine
   - Version: v2.x
   - In-memory index
   - Custom field mappings
   - Flexible query types
   - Analyzer support

2. Go Language
   - Version: 1.x
   - Standard library
   - Context support
   - Error handling
   - Testing framework

## Dependencies

1. Required Packages
   ```go
   github.com/blevesearch/bleve/v2
   github.com/go-go-golems/glazed/pkg/cmds
   ```

2. Development Tools
   - Go toolchain
   - Testing utilities
   - Benchmarking tools
   - Documentation tools

## Technical Constraints

1. Memory Usage
   - In-memory index
   - Resource limits
   - Garbage collection
   - Memory management

2. Performance
   - Query response time
   - Index update speed
   - Search accuracy
   - Resource efficiency

3. Compatibility
   - Go version support
   - Bleve version support
   - Platform compatibility
   - API stability

## Development Setup

1. Environment
   - Go 1.x or higher
   - Git for version control
   - IDE with Go support
   - Testing tools

2. Build Process
   - Standard Go build
   - Unit tests
   - Integration tests
   - Benchmarks

3. Testing
   - Unit test suite
   - Integration tests
   - Performance tests
   - Coverage reports

## Technical Decisions

1. Search Engine
   - Using Bleve for flexibility
   - In-memory index for simplicity
   - Custom field mappings
   - Query builder pattern

2. Field Analysis
   - Keyword analyzer for exact matches
   - Standard analyzer for text
   - Dynamic mapping for metadata
   - Field-specific settings

3. Query Building
   - Fluent builder API
   - Type-safe methods
   - Boolean operations
   - Helper functions

4. Error Handling
   - Explicit error types
   - Context support
   - Resource cleanup
   - Validation checks

## Performance Considerations

1. Index Configuration
   - Field-specific analyzers
   - Selective field storage
   - Memory optimization
   - Batch operations

2. Query Optimization
   - Query planning
   - Result caching
   - Connection pooling
   - Resource limits

3. Memory Management
   - Index size limits
   - Resource cleanup
   - Memory monitoring
   - Garbage collection

## Development Guidelines

1. Code Structure
   - Package organization
   - Interface design
   - Error handling
   - Documentation

2. Testing
   - Unit test coverage
   - Integration testing
   - Performance testing
   - Documentation testing

3. Documentation
   - Package documentation
   - API documentation
   - Examples
   - Usage guides

4. Performance
   - Benchmarking
   - Profiling
   - Optimization
   - Monitoring 