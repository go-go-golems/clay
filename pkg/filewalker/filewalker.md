# AST Walker for Filesystem Traversal

## Introduction

The **AST Walker** is a Go package designed to traverse files and directories, providing a flexible and efficient way to perform operations over a set of files and directories. It constructs an Abstract Syntax Tree (AST)-like structure of the filesystem, allowing you to perform pre- and post-visit operations on each node (file or directory).

This package is particularly useful when you need to:

- Process files and directories recursively.
- Apply filters or transformations to a set of files.
- Navigate and manipulate the filesystem tree programmatically.

## Features

- **Flexible Traversal**: Start traversal from directories or a list of specific filenames.
- **Support for fs.FS**: Work with both standard file systems and custom file system implementations.
- **Virtual File Tree**: Create a virtual file tree from a list of paths when no fs.FS is provided.
- **Pre and Post Visit Callbacks**: Execute custom functions before and after visiting each node.
- **Cross-Tree Navigation**: Retrieve nodes by path or relative path during traversal.
- **Configurable Options**: Customize behavior such as following symbolic links.

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

Instantiate a new walker with the desired options:

```go
// For standard file system
standardFS := os.DirFS("/path/to/root")
walker := filewalker.NewWalker(
    filewalker.WithFS(standardFS),
    filewalker.WithFollowSymlinks(false), // Don't follow symbolic links
)

// For virtual file tree
virtualWalker := filewalker.NewWalker()
```

### Defining Visit Functions

Define functions to execute before and after visiting each node:

```go
preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    fmt.Printf("Visiting: %s\n", node.Path)
    // Custom logic here
    return nil
}

postVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    fmt.Printf("Finished: %s\n", node.Path)
    // Custom logic here
    return nil
}
```

### Starting the Walk

Begin traversal from specified root paths:

```go
// With fs.FS
rootPaths := []string{".", "subdir"}
if err := walker.Walk(rootPaths, preVisit, postVisit); err != nil {
    log.Fatal(err)
}

// With virtual file tree
paths := []string{
    "/file1.txt",
    "/subdir1/file2.txt",
    "/subdir1/subdir2/file3.txt",
}
if err := virtualWalker.Walk(paths, preVisit, postVisit); err != nil {
    log.Fatal(err)
}
```

### Retrieving Nodes by Path

You can retrieve any node by its absolute path during traversal:

```go
node, err := w.GetNodeByPath("/absolute/path/to/file.txt")
if err != nil {
    fmt.Println("Node not found:", err)
} else {
    fmt.Println("Found node:", node.Path)
}
```

Or retrieve a node relative to another node:

```go
relativeNode, err := w.GetNodeByRelativePath(baseNode, "relative/path.txt")
if err != nil {
    fmt.Println("Relative node not found:", err)
} else {
    fmt.Println("Found relative node:", relativeNode.Path)
}
```

## API Reference

### Types

#### Walker

Represents the file system walker.

```go
type Walker struct {
    FollowSymlinks bool
    fs             fs.FS // Optional
    // Other fields...
}
```

- **Methods**:
  - `NewWalker(opts ...WalkerOption) *Walker`
  - `Walk(paths []string, preVisit VisitFunc, postVisit VisitFunc) error`
  - `GetNodeByPath(path string) (*Node, error)`
  - `GetNodeByRelativePath(baseNode *Node, relativePath string) (*Node, error)`

#### Node

Represents a file or directory node in the AST.

```go
type Node struct {
    Type     NodeType
    Path     string
    Parent   *Node
    Children []*Node
}
```

- **Methods**:
  - `GetType() NodeType`
  - `GetPath() string`
  - `GetParent() *Node`
  - `ImmediateChildren() []*Node`
  - `AllDescendants() []*Node`

#### NodeType

Enumeration of node types.

```go
type NodeType int

const (
    FileNode NodeType = iota
    DirectoryNode
)
```

#### VisitFunc

Function signature for visit callbacks.

```go
type VisitFunc func(w *Walker, node *Node) error
```

### Functions

#### NewWalker

Creates a new `Walker` with optional configurations.

```go
func NewWalker(opts ...WalkerOption) *Walker
```

#### WithFS

Option to set the file system for the Walker.

```go
func WithFS(fsys fs.FS) WalkerOption
```

#### WithFollowSymlinks

Option to set whether the walker follows symbolic links.

```go
func WithFollowSymlinks(follow bool) WalkerOption
```

#### WalkerOption

Function type for configuring the walker.

```go
type WalkerOption func(*Walker)
```

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

## Best Practices

- **Error Handling**: Always check for errors when calling walker methods.
- **Resource Management**: Be mindful of memory usage, especially when traversing large filesystems.
- **Avoid Side Effects**: Be cautious when modifying the filesystem during traversal, as it may affect the walker's behavior.
- **Concurrency**: The current implementation is not thread-safe. If you need concurrent access, consider implementing synchronization.

## Advanced Usage

### Customizing File Parsing

You can implement custom parsing logic within your visit functions:

```go
preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
    if node.Type == filewalker.FileNode && filepath.Ext(node.Path) == ".json" {
        content, err := os.ReadFile(node.Path)
        if err != nil {
            return err
        }
        var parsedData map[string]interface{}
        err = json.Unmarshal(content, &parsedData)
        if err != nil {
            return err
        }
        // Process parsedData here
    }
    return nil
}
```

### Modifying the Walker Behavior

Add custom options or modify the walker by extending the `WalkerOption`:

```go
func WithCustomOption(value string) filewalker.WalkerOption {
    return func(w *filewalker.Walker) {
        w.CustomField = value
    }
}

walker := filewalker.NewWalker(
    WithCustomOption("myValue"),
)
```

## Limitations

- **Large Filesystems**: Storing all nodes in memory may not be feasible for extremely large filesystems.
- **Symbolic Links**: By default, the walker does not follow symbolic links to prevent infinite loops.
- **Thread Safety**: The walker is not safe for concurrent use without additional synchronization.
