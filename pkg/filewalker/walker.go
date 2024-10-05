package filewalker

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"
)

// NodeType represents the type of the node: file or directory.
type NodeType int

const (
	FileNode NodeType = iota
	DirectoryNode
)

// Node represents a file or directory in the AST.
type Node struct {
	Type     NodeType
	Path     string
	Parent   *Node
	Children []*Node
}

// GetType returns the type of the node (file or directory).
func (n *Node) GetType() NodeType {
	return n.Type
}

// GetPath returns the path of the node.
func (n *Node) GetPath() string {
	return n.Path
}

// GetParent returns the parent node.
func (n *Node) GetParent() *Node {
	return n.Parent
}

// ImmediateChildren returns the immediate child nodes.
func (n *Node) ImmediateChildren() []*Node {
	return n.Children
}

// AllDescendants returns all descendant nodes recursively.
func (n *Node) AllDescendants() []*Node {
	var descendants []*Node
	for _, child := range n.Children {
		descendants = append(descendants, child)
		descendants = append(descendants, child.AllDescendants()...)
	}
	return descendants
}

// WalkerOption defines a function type for configuring the Walker.
type WalkerOption func(*Walker)

// Walker traverses the file system and builds the AST.
type Walker struct {
	FollowSymlinks bool
	nodeMap        map[string]*Node
	fs             fs.FS
	paths          []string
}

// NewWalker creates a new Walker with the provided options.
func NewWalker(opts ...WalkerOption) (*Walker, error) {
	w := &Walker{
		nodeMap: make(map[string]*Node),
	}
	for _, opt := range opts {
		opt(w)
	}

	if w.fs == nil && len(w.paths) == 0 {
		return nil, fmt.Errorf("either fs.FS must be set or paths must not be empty")
	}

	if w.fs == nil && len(w.paths) > 0 {
		mapFS := fstest.MapFS{}
		for _, path := range w.paths {
			mapFS[path] = &fstest.MapFile{}
		}
		w.fs = mapFS
	}

	return w, nil
}

// WithFS sets the file system for the Walker.
func WithFS(fsys fs.FS) WalkerOption {
	return func(w *Walker) {
		w.fs = fsys
	}
}

// WithPaths sets the paths for the Walker.
func WithPaths(paths []string) WalkerOption {
	return func(w *Walker) {
		w.paths = paths
	}
}

// WithFollowSymlinks sets whether the walker should follow symbolic links.
func WithFollowSymlinks(follow bool) WalkerOption {
	return func(w *Walker) {
		w.FollowSymlinks = follow
	}
}

// VisitFunc defines the function signature for pre- and post-visit callbacks.
type VisitFunc func(w *Walker, node *Node) error

// Walk traverses the file system or creates a virtual file tree from the given paths.
func (w *Walker) Walk(paths []string, preVisit VisitFunc, postVisit VisitFunc) error {
	if w.fs == nil && len(paths) == 0 {
		return fmt.Errorf("either fs.FS must be set or paths must not be empty")
	}

	if w.fs != nil {
		return w.walkFS(paths, preVisit, postVisit)
	}

	return w.createVirtualTree(paths, preVisit, postVisit)
}

func (w *Walker) walkFS(rootPaths []string, preVisit VisitFunc, postVisit VisitFunc) error {
	for _, rootPath := range rootPaths {
		node, err := w.buildFSNode(nil, ".", rootPath)
		if err != nil {
			return err
		}
		if err := w.walkNode(node, preVisit, postVisit); err != nil {
			return err
		}
	}
	return nil
}

func (w *Walker) createVirtualTree(paths []string, preVisit VisitFunc, postVisit VisitFunc) error {
	root := &Node{
		Type: DirectoryNode,
		Path: "/",
	}
	w.nodeMap["/"] = root

	for _, path := range paths {
		if err := w.addVirtualNode(root, path); err != nil {
			return err
		}
	}

	return w.walkNode(root, preVisit, postVisit)
}

func (w *Walker) addVirtualNode(root *Node, path string) error {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	current := root

	for i, part := range parts {
		if part == "" {
			continue
		}

		childPath := "/" + filepath.Join(parts[:i+1]...)
		child, exists := w.nodeMap[childPath]

		if !exists {
			isDir := i < len(parts)-1
			nodeType := DirectoryNode
			if !isDir {
				nodeType = FileNode
			}
			child = &Node{
				Type:   nodeType,
				Path:   childPath,
				Parent: current,
			}
			current.Children = append(current.Children, child)
			w.nodeMap[childPath] = child
		}

		current = child
	}

	return nil
}

func (w *Walker) buildFSNode(parent *Node, basePath string, path string) (*Node, error) {
	fileInfo, err := fs.Stat(w.fs, path)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Type:   determineNodeType(fileInfo.IsDir()),
		Path:   path,
		Parent: parent,
	}

	w.nodeMap[path] = node

	if fileInfo.IsDir() {
		entries, err := fs.ReadDir(w.fs, path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			info, err := entry.Info()
			if err != nil {
				return nil, err
			}
			if !w.FollowSymlinks && isSymlink(info) {
				continue
			}
			childNode, err := w.buildFSNode(node, basePath, childPath)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, childNode)
		}
	}
	return node, nil
}

func (w *Walker) walkNode(node *Node, preVisit VisitFunc, postVisit VisitFunc) error {
	if preVisit != nil {
		if err := preVisit(w, node); err != nil {
			return err
		}
	}

	for _, child := range node.Children {
		if err := w.walkNode(child, preVisit, postVisit); err != nil {
			return err
		}
	}

	if postVisit != nil {
		if err := postVisit(w, node); err != nil {
			return err
		}
	}
	return nil
}

func determineNodeType(isDir bool) NodeType {
	if isDir {
		return DirectoryNode
	}
	return FileNode
}

func isSymlink(fileInfo os.FileInfo) bool {
	return fileInfo.Mode()&os.ModeSymlink != 0
}

// GetNodeByPath retrieves a node by its path.
func (w *Walker) GetNodeByPath(path string) (*Node, error) {
	node, ok := w.nodeMap[path]
	if !ok {
		return nil, fmt.Errorf("node not found for path: %s", path)
	}
	return node, nil
}

// GetNodeByRelativePath retrieves a node by a path relative to a base node.
func (w *Walker) GetNodeByRelativePath(baseNode *Node, relativePath string) (*Node, error) {
	fullPath := filepath.Join(baseNode.Path, relativePath)
	return w.GetNodeByPath(fullPath)
}
