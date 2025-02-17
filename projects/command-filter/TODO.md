# Command Filter Implementation Plan

## Overview
Implement a flexible command filter system using Bleve as the search backend, allowing searching through CommandDescription objects using various criteria and boolean combinations.

## Core Components

### 1. Package Structure
- [ ] Create `pkg/filters/command` package
- [ ] Add `index.go` for CommandIndex implementation
- [ ] Add `filter.go` for filter types
- [ ] Add `builder.go` for query builder
- [ ] Add `document.go` for document structure

### 2. Command Index Implementation
- [ ] Create CommandIndex struct
  ```go
  type CommandIndex struct {
      index bleve.Index
  }
  ```
- [ ] Implement NewCommandIndex constructor
- [ ] Add Close method
- [ ] Add Search method with context support
- [ ] Add error handling and logging

### 3. Document Structure
- [ ] Define commandDocument struct
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
- [ ] Add conversion methods from/to CommandDescription
- [ ] Add validation for document fields

### 4. Filter Implementation
- [ ] Create BleveFilter type
  ```go
  type BleveFilter struct {
      query bleve.Query
  }
  ```
- [ ] Implement filter methods:
  - [ ] ExactName
  - [ ] NamePattern
  - [ ] ParentsPrefix
  - [ ] ParentsGlob
  - [ ] Type
  - [ ] HasTag
  - [ ] HasAnyTag
  - [ ] HasAllTags
  - [ ] MetadataField
- [ ] Add boolean combinations:
  - [ ] And
  - [ ] Or

### 5. Query Builder
- [ ] Create QueryBuilder type
- [ ] Add methods for all filter types
- [ ] Add helper methods for common queries
- [ ] Add validation for query parameters

## Testing

### 1. Unit Tests
- [ ] Test CommandIndex
  - [ ] Creation and closing
  - [ ] Document indexing
  - [ ] Search functionality
  - [ ] Error cases

- [ ] Test Filters
  - [ ] Each filter type
  - [ ] Boolean combinations
  - [ ] Edge cases
  - [ ] Invalid inputs

- [ ] Test Document Conversion
  - [ ] CommandDescription to document
  - [ ] Document to CommandDescription
  - [ ] Field validation

### 2. Integration Tests
- [ ] Test with real CommandDescription objects
- [ ] Test complex query combinations
- [ ] Test concurrent searches
- [ ] Test with large command sets

## Documentation

### 1. Package Documentation
- [ ] Add package overview
- [ ] Document types and interfaces
- [ ] Add usage examples
- [ ] Document error handling

### 2. Examples
- [ ] Basic usage examples
- [ ] Complex query examples
- [ ] Common use case examples
- [ ] Error handling examples

## CLI Integration

### 1. Command Line Interface
- [ ] Add filter subcommand
- [ ] Add filter flags for each type
- [ ] Add output formatting
- [ ] Add error reporting

### 2. Usage Examples
```go
// Example command usage:
clay filter --type http --tag api --parent-glob "service/*/api"
clay filter --name-pattern "serve*" --has-all-tags stable,v2
clay filter --metadata version=2.0.0 --type grpc
```

## Implementation Order

1. Core Structure
   - [ ] Set up package structure
   - [ ] Implement CommandIndex
   - [ ] Add basic document structure

2. Basic Functionality
   - [ ] Implement simple filters
   - [ ] Add basic search functionality
   - [ ] Create initial tests

3. Advanced Features
   - [ ] Add all filter types
   - [ ] Implement boolean combinations
   - [ ] Add metadata handling

4. Integration
   - [ ] Add CLI integration
   - [ ] Create documentation
   - [ ] Add examples

## Notes
- Use in-memory Bleve index for simplicity
- Focus on clean API design
- Ensure proper error handling
- Add context support for cancellation
- Keep memory usage in check
