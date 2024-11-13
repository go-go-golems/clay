---
Title: Using the FileFilter Package
Slug: filefilter
Short: Learn how to use the filefilter package to filter and process files in your Go applications
Topics: 
- filefilter
- files
- filtering
Commands:
- NewFileFilter
- FilterPath
- FilterNode
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The filefilter package provides a flexible and powerful way to filter files and directories in Go applications. It's particularly useful when you need to process specific files while excluding others based on various criteria such as file size, extensions, patterns, and Git ignore rules.

## Core Features

- File size filtering
- Extension-based filtering (include/exclude)
- Pattern matching for filenames and paths
- Directory exclusion
- GitIgnore integration
- Binary file detection
- Profile-based configurations
- YAML configuration support

## Basic Usage

### Creating a Basic FileFilter

The simplest way to create a FileFilter is using the `NewFileFilter` constructor:

```go
import "github.com/your-repo/filefilter"

// Create with default settings
ff := filefilter.NewFileFilter()

// Check if a file should be processed
shouldProcess := ff.FilterPath("path/to/your/file.txt")
```

### Customizing with Options

You can customize the FileFilter using various options:

```go
ff := filefilter.NewFileFilter(
    filefilter.WithMaxFileSize(5 * 1024 * 1024), // 5MB
    filefilter.WithIncludeExts([]string{".go", ".js", ".ts"}),
    filefilter.WithExcludeDirs([]string{"node_modules", "vendor"}),
    filefilter.WithFilterBinaryFiles(true),
)
```

## File Extension Filtering

### Including Specific Extensions

```go
ff := filefilter.NewFileFilter(
    filefilter.WithIncludeExts([]string{".go", ".md", ".txt"}),
)
```

### Excluding Extensions

```go
ff := filefilter.NewFileFilter(
    filefilter.WithExcludeExts([]string{".exe", ".dll", ".so"}),
)
```

## Pattern Matching

### Filename Pattern Matching

```go
ff := filefilter.NewFileFilter(
    filefilter.WithMatchFilenames([]string{
        `^test.*\.go$`,     // Files starting with "test"
        `_test\.go$`,       // Go test files
    }),
)
```

### Path Pattern Matching

```go
ff := filefilter.NewFileFilter(
    filefilter.WithMatchPaths([]string{
        `^src/.*\.go$`,     // Go files in src directory
        `^tests/.*_test\.go$`, // Test files in tests directory
    }),
)
```

## Directory Exclusion

```go
ff := filefilter.NewFileFilter(
    filefilter.WithExcludeDirs([]string{
        "node_modules",
        "vendor",
        ".git",
        "build",
    }),
)
```

## GitIgnore Integration

By default, FileFilter respects .gitignore rules. You can disable this feature:

```go
ff := filefilter.NewFileFilter(
    filefilter.WithDisableGitIgnore(true),
)
```

## Integration with FileWalker

The FileFilter package is designed to work seamlessly with the FileWalker package, providing powerful file traversal and filtering capabilities:

```go
import (
    "github.com/your-repo/filefilter"
    "github.com/go-go-golems/clay/pkg/filewalker"
)

// Create file filter
ff := filefilter.NewFileFilter(
    filefilter.WithIncludeExts([]string{".go", ".md"}),
    filefilter.WithExcludeDirs([]string{"vendor"}),
)

// Create walker with the file filter
walker, err := filewalker.NewWalker(
    filewalker.WithFilter(ff.FilterNode),
)
if err != nil {
    log.Fatal(err)
}

// Walk and process only filtered files
err = walker.Walk([]string{"."}, func(w *filewalker.Walker, node *filewalker.Node) error {
    fmt.Printf("Processing filtered file: %s\n", node.GetPath())
    return nil
}, nil)
```

### Advanced Integration Example

Here's a more complex example showing how to combine FileFilter and FileWalker features:

```go
// Create a file filter with multiple conditions
ff := filefilter.NewFileFilter(
    filefilter.WithMaxFileSize(1024 * 1024), // 1MB
    filefilter.WithIncludeExts([]string{".go"}),
    filefilter.WithMatchFilenames([]string{
        `^test.*\.go$`,
        `_test\.go$`,
    }),
    filefilter.WithFilterBinaryFiles(true),
)

// Create a walker with the filter
walker, err := filewalker.NewWalker(
    filewalker.WithFilter(ff.FilterNode),
    filewalker.WithFollowSymlinks(false),
)
if err != nil {
    log.Fatal(err)
}

// Walk with both pre and post visit functions
err = walker.Walk([]string{"."}, 
    // Pre-visit function
    func(w *filewalker.Walker, node *filewalker.Node) error {
        if node.GetType() == filewalker.FileNode {
            fmt.Printf("Found test file: %s\n", node.GetPath())
        }
        return nil
    },
    // Post-visit function
    func(w *filewalker.Walker, node *filewalker.Node) error {
        if node.GetType() == filewalker.DirectoryNode {
            children := node.ImmediateChildren()
            fmt.Printf("Directory %s contains %d filtered files\n", 
                node.GetPath(), len(children))
        }
        return nil
    },
)
```

## YAML Configuration

### Saving Configuration

```go
ff := filefilter.NewFileFilter(
    filefilter.WithIncludeExts([]string{".go", ".js"}),
    filefilter.WithExcludeDirs([]string{"node_modules"}),
)

err := ff.SaveToFile("filefilter-config.yaml")
if err != nil {
    // Handle error
}
```

### Loading Configuration

```go
ff, err := filefilter.LoadFromFile("filefilter-config.yaml", "")
if err != nil {
    // Handle error
}
```

### Using Profiles

You can define multiple profiles in your YAML configuration:

```yaml
max-file-size: 1048576
include-exts:
  - .go
  - .js
profiles:
  docs:
    include-exts:
      - .md
      - .txt
  code:
    include-exts:
      - .go
      - .rs
```

Load a specific profile:

```go
ff, err := filefilter.LoadFromFile("filefilter-config.yaml", "docs")
if err != nil {
    // Handle error
}
```

## Command Line Integration

The package provides integration with Glazed for command-line applications:

```go
layer, err := filefilter.NewFileFilterParameterLayer()
if err != nil {
    // Handle error
}

// Use the layer in your Glazed command
ff, err := filefilter.CreateFileFilterFromSettings(parsedLayer)
if err != nil {
    // Handle error
}
```

## Debug and Verbose Mode

Enable verbose mode to debug which files are being filtered:

```go
ff := filefilter.NewFileFilter(
    filefilter.WithVerbose(true),
)
```

## Default Filters

The package comes with sensible defaults for binary files, common build artifacts, and system files. You can disable these defaults:

```go
ff := filefilter.NewFileFilter(
    filefilter.WithDisableDefaultFilters(true),
)
```

## Best Practices

1. Start with default settings and customize as needed
2. Use profiles for different file processing scenarios
3. Enable verbose mode during development for debugging
4. Consider storing common configurations in YAML files
5. Use pattern matching carefully to avoid performance impacts
6. Test your filter configuration with a small dataset first

## Common Use Cases

- Processing source code files while ignoring build artifacts
- Analyzing documentation while excluding binary files
- Implementing custom build systems
- Creating file processing pipelines
- Implementing search functionality
- Managing document processing workflows
