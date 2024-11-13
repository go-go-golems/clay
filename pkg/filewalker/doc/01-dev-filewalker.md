---
Title: Using the FileWalker Package
Slug: filewalker
Short: Learn how to use the filewalker package to traverse and manage file system structures in Go applications
Topics:
- filewalker
- filesystem
- traversal
Commands:
- NewWalker
- Walk
- GetNodeByPath
- GetNodeByRelativePath
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The filewalker package provides a powerful and flexible way to traverse file systems while building an Abstract Syntax Tree (AST)-like structure. It's designed to support both standard filesystems and custom fs.FS implementations, making it ideal for file processing, analysis, and manipulation tasks.

## Core Features

- AST-like file system representation
- Support for standard and custom fs.FS implementations
- Pre-visit and post-visit callbacks
- Symlink handling
- Path filtering
- Node retrieval by absolute or relative paths
- Virtual file tree creation from path lists

## Basic Usage

### Creating a Basic Walker

```go
import "github.com/go-go-golems/clay/pkg/filewalker"

// Create with default settings
walker, err := filewalker.NewWalker()
if err != nil {
    log.Fatal(err)
}

// Walk the file system
err = walker.Walk([]string{"."}, func(w *filewalker.Walker, node *filewalker.Node) error {
    fmt.Printf("Visiting: %s\n", node.GetPath())
    return nil
}, nil)
```

### Customizing with Options

```go
walker, err := filewalker.NewWalker(
    filewalker.WithFS(customFS),           // Use custom filesystem
    filewalker.WithFollowSymlinks(true),   // Follow symbolic links
    filewalker.WithFilter(filterFunc),     // Add custom filtering
)
```

## File System Traversal

### Basic Directory Walking

```go
// Define pre-visit function
preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    if node.GetType() == filewalker.FileNode {
        fmt.Printf("Found file: %s\n", node.GetPath())
    }
    return nil
}

// Define post-visit function
postVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    if node.GetType() == filewalker.DirectoryNode {
        fmt.Printf("Finished directory: %s\n", node.GetPath())
    }
    return nil
}

// Walk the filesystem
err = walker.Walk([]string{"path/to/dir"}, preVisit, postVisit)
```

### Working with Nodes

```go
// Access node information
preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    fmt.Printf("Type: %v\n", node.GetType())
    fmt.Printf("Path: %s\n", node.GetPath())
    
    if parent := node.GetParent(); parent != nil {
        fmt.Printf("Parent: %s\n", parent.GetPath())
    }
    
    for _, child := range node.ImmediateChildren() {
        fmt.Printf("Child: %s\n", child.GetPath())
    }
    
    return nil
}
```

## Path Filtering

### Using Custom Filters

```go
// Create a filter that only includes Go files and directories
filter := func(node *filewalker.Node) bool {
    if node.GetType() == filewalker.DirectoryNode {
        return true
    }
    return strings.HasSuffix(node.GetPath(), ".go")
}

walker, err := filewalker.NewWalker(
    filewalker.WithFilter(filter),
)
```

## Node Retrieval

### Getting Nodes by Path

```go
// Get node by absolute path
node, err := walker.GetNodeByPath("/path/to/file")

// Get node by relative path from a base node
baseNode, _ := walker.GetNodeByPath("/base/path")
relNode, err := walker.GetNodeByRelativePath(baseNode, "subdir/file.txt")
```

## Working with Virtual File Trees

```go
paths := []string{
    "/path/to/file1.txt",
    "/path/to/file2.txt",
    "/another/path/file3.txt",
}

walker, err := filewalker.NewWalker(
    filewalker.WithPaths(paths),
)

err = walker.Walk(paths, func(w *filewalker.Walker, node *filewalker.Node) error {
    fmt.Printf("Virtual node: %s\n", node.GetPath())
    return nil
}, nil)
```

## Error Handling

The Walker provides comprehensive error handling for common scenarios:

- Non-existent paths
- Permission denied errors
- I/O errors during file reading
- Invalid symlinks
- Custom errors from visit functions

```go
err = walker.Walk(paths, func(w *filewalker.Walker, node *filewalker.Node) error {
    if someCondition {
        return fmt.Errorf("custom error for %s", node.GetPath())
    }
    return nil
}, nil)

if err != nil {
    switch {
    case errors.Is(err, fs.ErrNotExist):
        fmt.Println("Path does not exist")
    case errors.Is(err, fs.ErrPermission):
        fmt.Println("Permission denied")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Best Practices

1. Use appropriate visit functions for your use case
2. Implement proper error handling
3. Use filters to reduce unnecessary processing
4. Be careful with symlink following to avoid loops
5. Consider memory usage for large directory structures
6. Use relative path resolution for portable code

## Common Use Cases

- Building file indexers
- Implementing search functionality
- Creating backup systems
- Processing file hierarchies
- Analyzing directory structures
- Implementing custom file processors

## Additional Examples

### Computing Total File Size

```go
var totalSize int64

preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    if node.GetType() == filewalker.FileNode {
        totalSize += node.Data.Size
    }
    return nil
}

walker, err := filewalker.NewWalker()
if err != nil {
    log.Fatal(err)
}

if err := walker.Walk([]string{"./data"}, preVisit, nil); err != nil {
    log.Fatal(err)
}

fmt.Printf("Total size of files: %d bytes\n", totalSize)
```

### Accessing File Content

```go
preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    if node.GetType() == filewalker.FileNode && 
       filepath.Base(node.GetPath()) == "config.yaml" {
        content, err := os.ReadFile(node.GetPath())
        if err != nil {
            return err
        }
        fmt.Printf("Config file found at %s:\n%s\n", node.GetPath(), string(content))
    }
    return nil
}
```
