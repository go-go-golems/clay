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