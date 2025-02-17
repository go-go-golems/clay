# Command Filter Implementation Plan

## Overview
Implement a flexible command filter system using Bleve as the search backend, allowing searching through CommandDescription objects using various criteria and boolean combinations.

## Core Components

### 1. Package Structure
- [x] Create `pkg/filters/command` package
- [x] Add `index.go` for CommandIndex implementation
- [x] Add `filter.go` for filter types
- [x] Add `builder.go` for query builder
- [x] Add `document.go` for document structure
- [x] Add `pkg/filters/command/builder` package for new API
  - [x] Add `builder.go` for main interface
  - [x] Add `filter.go` for filter builder
  - [x] Add `options.go` for builder options

### 2. Command Index Implementation
- [x] Create CommandIndex struct
  ```go
  type CommandIndex struct {
      index bleve.Index
  }
  ```
- [x] Implement NewCommandIndex constructor
- [x] Add Close method
- [x] Add Search method with context support
- [x] Add error handling and logging
- [x] Update Search method to use new FilterBuilder

### 3. Document Structure
- [x] Define commandDocument struct
  ```go
  type commandDocument struct {
      Name        string
      NamePattern string
      FullPath    string
      Parents     []string
      Type        string
      Tags        []string
      Metadata    map[string]interface{}
  }
  ```
- [x] Add conversion methods from/to CommandDescription
- [x] Add validation for document fields

### 4. New Query Builder Implementation
- [x] Create QueryBuilder interface
  ```go
  type QueryBuilder interface {
      Type(type_ string) *FilterBuilder
      Types(types ...string) *FilterBuilder
      Tag(tag string) *FilterBuilder
      Tags(tags ...string) *FilterBuilder
      // ... other methods
  }
  ```
- [x] Implement FilterBuilder
  ```go
  type FilterBuilder struct {
      query query.Query
  }
  ```
- [x] Add builder methods:
  - [x] Type filters (Type, Types)
  - [x] Tag filters (Tag, Tags, AllTags, AnyTags)
  - [x] Path filters (Path, PathGlob, PathPrefix)
  - [x] Name filters (Name, NamePattern)
  - [x] Metadata filters (Metadata, MetadataMatch)
- [x] Add combination methods:
  - [x] And
  - [x] Or
  - [x] Not
- [x] Add helper functions:
  - [x] Must
  - [x] NewFilter
  - [x] WithOptions

## Testing

### 1. Unit Tests
- [x] Test CommandIndex
  - [x] Creation and closing
  - [x] Document indexing
  - [x] Search functionality
  - [x] Error cases

- [x] Test New Query Builder
  - [x] Individual filter methods
  - [x] Filter combinations
  - [x] Helper functions
  - [x] Options handling

- [x] Test Document Conversion
  - [x] CommandDescription to document
  - [x] Document to CommandDescription
  - [x] Field validation

### 2. Integration Tests
- [ ] Test complex query combinations:
  - [x] Type AND Tag combinations:
    ```go
    builder.Type("http").And(builder.Tag("api"))
    builder.Types("http", "grpc").And(builder.AllTags("api", "stable"))
    ```
  - [x] Path-based combinations:
    ```go
    builder.PathPrefix("service/").And(builder.Type("http"))
    builder.PathGlob("*/api/*").And(builder.Tag("stable"))
    ```
  - [x] Metadata combinations:
    ```go
    builder.Metadata("version", "2.0.0").And(builder.Tag("stable"))
    builder.MetadataMatch(map[string]interface{}{
        "version": "2.0.0",
        "stage": "prod",
    }).And(builder.Type("http"))
    ```
  - [x] Name pattern combinations:
    ```go
    builder.NamePattern("serve*").And(builder.Type("http"))
    builder.Name("api-server").Or(builder.Name("web-server"))
    ```
  - [x] Complex nested combinations:
    ```go
    builder.Type("http").And(
        builder.Or(
            builder.Tag("api"),
            builder.Tag("web"),
        ),
    ).And(
        builder.MetadataMatch(map[string]interface{}{
            "version": "2.0.0",
            "stage": "prod",
        }),
    )
    ```
  - [x] NOT combinations:
    ```go
    builder.Type("http").And(
        builder.Not(builder.Tag("deprecated")),
    )
    builder.PathPrefix("service/").And(
        builder.Not(builder.Or(
            builder.Type("test"),
            builder.Tag("experimental"),
        )),
    )
    ```
  - [x] Multi-level combinations:
    ```go
    builder.Or(
        builder.And(
            builder.Type("http"),
            builder.Tag("api"),
            builder.Metadata("version", "2.0.0"),
        ),
        builder.And(
            builder.Type("grpc"),
            builder.Tag("internal"),
            builder.PathPrefix("service/"),
        ),
    )
    ```

- [ ] Test with large command sets
- [ ] Test migration scenarios

## Documentation

### 1. Package Documentation
- [x] Add package overview
- [x] Document types and interfaces
- [x] Add usage examples
- [x] Document error handling
- [ ] Add migration guide
- [x] Document builder options

### 2. Examples
- [x] Basic usage examples
- [x] Complex query examples
- [x] Common use case examples
- [x] Error handling examples
- [ ] Migration examples

## CLI Integration

### 1. Command Line Interface
- [ ] Add filter subcommand
- [ ] Add filter flags for each type
- [ ] Add output formatting
- [ ] Add error reporting
- [ ] Update to use new builder API

### 2. Usage Examples
```go
// Example command usage:
clay filter --type http --tag api --parent-glob "service/*/api"
clay filter --name-pattern "serve*" --has-all-tags stable,v2
clay filter --metadata version=2.0.0 --type grpc
```

## Implementation Order

1. Core Structure
   - [x] Set up package structure
   - [x] Implement CommandIndex
   - [x] Add basic document structure

2. Basic Functionality
   - [x] Implement simple filters
   - [x] Add basic search functionality
   - [x] Create initial tests

3. Advanced Features
   - [x] Add all filter types
   - [x] Implement boolean combinations
   - [x] Add metadata handling

4. New Builder API
   - [x] Create builder package
   - [x] Implement core interfaces
   - [x] Add helper functions
   - [x] Create migration tools

5. Integration
   - [ ] Update CLI to use new API
   - [x] Create documentation
   - [x] Add examples
   - [ ] Add migration guide

## Notes
- Use in-memory Bleve index for simplicity
- Focus on clean API design
- Ensure proper error handling
- Add context support for cancellation
- Keep memory usage in check
- Provide smooth migration path
- Maintain backward compatibility during transition
