---
Title: Command Filter System
Slug: command-filter
Short: Learn how to use Clay's powerful command filtering system to search and organize commands using various criteria
Topics:
- filtering
- search
- commands
- organization
Commands:
- filter
- NewFilterCommand
- NewCommandIndex
Flags:
- type
- types
- tag
- tags
- all-tags
- any-tags
- path
- path-glob
- path-prefix
- name
- name-pattern
- metadata-key
- metadata-value
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The Clay command filter system provides a powerful way to search and filter commands based on various criteria such as type, tags, path, name, and metadata. It uses Bleve as the search backend and offers a modern, type-safe API for building complex queries.

## Overview

The filter system allows you to:
- Search commands by type, tags, name, and path
- Use pattern matching and glob patterns
- Filter based on metadata fields
- Combine filters with boolean operations
- Get structured output in various formats

## Using the Filter Command

The `filter` command provides a CLI interface to the filtering system. It can easily be added to any tool using glazed commands.

```shell
# Filter by type
XXX commands filter --type http

# Filter by multiple types
XXX commands filter --types http,grpc

# Filter by tag
XXX commands filter --tag api

# Filter by multiple tags (any match)
XXX commands filter --tags api,stable

# Filter by multiple tags (all must match)
XXX commands filter --all-tags api,stable

# Filter by path
XXX commands filter --path "services/api"

# Filter by path glob pattern
XXX commands filter --path-glob "services/*/api"

# Filter by path prefix
XXX commands filter --path-prefix "services/"

# Filter by name
XXX commands filter --name "http-server"

# Filter by name pattern
XXX commands filter --name-pattern "http-*"

# Filter by metadata
XXX commands filter --metadata-key version --metadata-value "1.0.0"
```

## Filter API

The filter system provides a fluent builder API for constructing queries programmatically:

```go
import (
    "github.com/go-go-golems/clay/pkg/filters/command/builder"
)

// Create a new builder
b := builder.New()

// Build simple filters
typeFilter := b.Type("http")
tagFilter := b.Tag("api")
pathFilter := b.Path("services/api")

// Build pattern filters
namePattern := b.NamePattern("http-*")
pathGlob := b.PathGlob("services/*/api")

// Build complex filters
filter := b.Type("http").
    And(b.AnyTags("api", "stable")).
    And(b.PathPrefix("services/"))
```

### Available Filter Methods

1. Type Filters:
   ```go
   b.Type("http")              // Single type
   b.Types("http", "grpc")     // Multiple types (OR)
   ```

2. Tag Filters:
   ```go
   b.Tag("api")                // Single tag
   b.Tags("api", "stable")     // Multiple tags (OR)
   b.AllTags("api", "stable")  // Multiple tags (AND)
   b.AnyTags("api", "stable")  // Alias for Tags
   ```

3. Path Filters:
   ```go
   b.Path("services/api")      // Exact path
   b.PathGlob("services/*/api") // Glob pattern
   b.PathPrefix("services/")    // Path prefix
   ```

4. Name Filters:
   ```go
   b.Name("http-server")       // Exact name
   b.NamePattern("http-*")     // Name pattern
   ```

5. Metadata Filters:
   ```go
   b.Metadata("version", "1.0.0")  // Single field
   b.MetadataMatch(map[string]interface{}{  // Multiple fields
       "version": "1.0.0",
       "stage": "prod",
   })
   ```

### Boolean Operations

Filters can be combined using boolean operations:

```go
// AND combination
filter := b.Type("http").And(b.Tag("api"))

// OR combination
filter := b.Type("http").Or(b.Type("grpc"))

// Complex combinations
filter := b.Type("http").
    And(
        b.AnyTags("api", "stable").
        Or(b.Tag("internal"))
    ).
    And(b.PathPrefix("services/"))
```

## Implementation Details

### Field Mappings

The filter system uses specialized field mappings for different types of fields:

1. Keyword Fields (exact matching):
   - `name`
   - `type`
   - `tags`
   - `full_path`

2. Text Fields (analyzed):
   - `name` (additional mapping for pattern matching)

3. Dynamic Fields:
   - `metadata.*` (supports various value types)

### Search Process

1. Query Building:
   ```
   Filter Methods -> Query Builder -> Bleve Query
   ```

2. Execution:
   ```
   Query -> Search Request -> Results -> Command List
   ```

3. Result Processing:
   ```
   Command List -> Structured Output -> Formatted Display
   ```

## Best Practices

1. Query Construction:
   - Use specific field queries when possible
   - Combine queries with AND/OR operations
   - Validate input patterns
   - Consider performance implications

2. Path Handling:
   - Use consistent path separators
   - Handle empty paths appropriately
   - Consider platform-specific issues
   - Test with deep hierarchies

3. Performance:
   - Avoid expensive wildcard queries
   - Use prefix queries when possible
   - Consider result set size
   - Monitor query performance

## Examples

### Finding HTTP API Commands

```go
filter := b.Type("http").
    And(b.Tag("api")).
    And(b.PathPrefix("services/"))
```

### Finding Stable or Beta Commands

```go
filter := b.AnyTags("stable", "beta").
    And(b.PathGlob("*/v[0-9]*"))
```

### Finding Commands by Version

```go
filter := b.MetadataMatch(map[string]interface{}{
    "version": "1.0.0",
    "stage": "prod",
})
```

## Debugging

The filter system includes comprehensive debug logging (when enabled):

```go
// Enable debug logging
log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)

// Logs will show:
// - Query construction details
// - Field mapping information
// - Search execution details
// - Result processing
```

## Error Handling

The system provides clear error messages for common issues:

1. Invalid Patterns:
   ```
   "invalid glob pattern: [invalid-glob"
   ```

2. Missing Fields:
   ```
   "command document must have a name"
   "command document must have a type"
   ```

3. Search Errors:
   ```
   "could not search commands: [error details]"
   ```
