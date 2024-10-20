# AST Walker for Filesystem Traversal

## Introduction

The **AST Walker** is a Go package designed to traverse files and directories, providing a flexible and efficient way to perform operations over a set of files and directories. It constructs an Abstract Syntax Tree (AST)-like structure of the filesystem, allowing you to perform pre- and post-visit operations on each node (file or directory).

## Features

- **Flexible Traversal**: Start traversal from directories or a list of specific filenames.
- **Support for fs.FS**: Work with both standard file systems and custom file system implementations.
- **Virtual File Tree**: Create a virtual file tree from a list of paths when no fs.FS is provided.
- **Pre and Post Visit Callbacks**: Execute custom functions before and after visiting each node.
- **Cross-Tree Navigation**: Retrieve nodes by path or relative path during traversal.
- **Configurable Options**: Customize behavior such as following symbolic links.
- **Node Filtering**: Apply custom filters to include or exclude nodes based on their properties.
- **Relative Path Resolution**: Automatically resolve relative paths to absolute paths.

## Installation

To install the package, use `go get`:

```bash
go get github.com/go-go-golems/clay/pkg/filewalker
```

Replace `go-go-golems/clay/pkg` with the appropriate username or repository path.

## Usage

### Importing the Package

```go
import (
    "github.com/go-go-golems/clay/pkg/filewalker"
)
```

### Creating a Walker

```go
// For standard file system
walker, err := filewalker.NewWalker()
if err != nil {
    log.Fatal(err)
}

// For custom file system
customFS := myCustomFS{}
walker, err := filewalker.NewWalker(filewalker.WithFS(customFS))
if err != nil {
    log.Fatal(err)
}
```

### Walking the File System

```go
err := walker.Walk([]string{"path/to/walk"}, func(w *filewalker.Walker, node *filewalker.Node) error {
    fmt.Printf("Visiting: %s\n", node.Path)
    return nil
}, nil)
if err != nil {
    log.Fatal(err)
}
```

### Resolving Relative Paths

The Walker now automatically resolves relative paths to absolute paths:

```go
absPath := walker.resolveRelativePath("relative/path")
fmt.Println(absPath) // Prints the absolute path
```

### Filtering Nodes

You can use the `WithFilter` option to specify a custom filter function that determines which nodes should be included in the walk:

```go
filter := func(node *filewalker.Node) bool {
    return node.Type == filewalker.DirectoryNode || strings.HasSuffix(node.Path, ".go")
}

walker, err := filewalker.NewWalker(filewalker.WithFilter(filter))
if err != nil {
    log.Fatal(err)
}

// This walk will only include directories and .go files
err = walker.Walk([]string{"."}, preVisitFunc, postVisitFunc)
if err != nil {
    log.Fatal(err)
}
```

## API Reference

### Walker

```go
type Walker struct {
    FollowSymlinks bool
    // Other unexported fields...
}

func NewWalker(opts ...WalkerOption) (*Walker, error)
func (w *Walker) Walk(paths []string, preVisit VisitFunc, postVisit VisitFunc) error
func (w *Walker) GetNodeByPath(path string) (*Node, error)
func (w *Walker) GetNodeByRelativePath(baseNode *Node, relativePath string) (*Node, error)
func (w *Walker) resolveRelativePath(path string) string
```

### Node

```go
type Node struct {
    Type     NodeType
    Path     string
    Parent   *Node
    Children []*Node
}

func (n *Node) GetType() NodeType
func (n *Node) GetPath() string
func (n *Node) GetParent() *Node
func (n *Node) ImmediateChildren() []*Node
func (n *Node) AllDescendants() []*Node
```

## Best Practices

- Always check for errors when creating a new Walker or calling its methods.
- Use the `resolveRelativePath` method when working with potentially relative paths.
- Be mindful of memory usage when traversing large file systems.
- Implement proper error handling in your visit functions.

## Examples

### Example 1: Computing Total File Size

Calculate the total size of all files in a directory:

```go
package main

import (
    "fmt"
    "log"

    "github.com/go-go-golems/clay/pkg/filewalker"
)

func main() {
    var totalSize int64

    preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
        if node.Type == filewalker.FileNode {
            totalSize += node.Data.Size
        }
        return nil
    }

    walker := filewalker.NewWalker()
    rootPaths := []string{"./data"}
    if err := walker.Walk(rootPaths, preVisit, nil); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total size of files: %d bytes\n", totalSize)
}
```

### Example 2: Filtering and Processing Files

Print the paths of all `.txt` files:

```go
preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    if node.Type == filewalker.FileNode && filepath.Ext(node.Path) == ".txt" {
        fmt.Println("Text file:", node.Path)
    }
    return nil
}

walker := filewalker.NewWalker()
rootPaths := []string{"./documents"}
if err := walker.Walk(rootPaths, preVisit, nil); err != nil {
    log.Fatal(err)
}
```

### Example 3: Accessing File Content

Read and print the content of files named `config.yaml`:

```go
preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    if node.Type == filewalker.FileNode && filepath.Base(node.Path) == "config.yaml" {
        content, err := os.ReadFile(node.Path)
        if err != nil {
            return err
        }
        fmt.Printf("Config file found at %s:\n%s\n", node.Path, string(content))
    }
    return nil
}

walker := filewalker.NewWalker()
rootPaths := []string{"/etc", "/usr/local/etc"}
if err := walker.Walk(rootPaths, preVisit, nil); err != nil {
    log.Fatal(err)
}
```

## Limitations

- **Large Filesystems**: Storing all nodes in memory may not be feasible for extremely large filesystems.
- **Symbolic Links**: By default, the walker does not follow symbolic links to prevent infinite loops.
- **Thread Safety**: The walker is not safe for concurrent use without additional synchronization.
