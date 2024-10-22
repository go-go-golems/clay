package filewalker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWalker_EmptyDirectory(t *testing.T) {
	testFS := fstest.MapFS{
		"empty_dir/": &fstest.MapFile{Mode: os.ModeDir},
	}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	var visitedNodes int
	err = w.Walk([]string{"empty_dir"}, func(w *Walker, node *Node) error {
		visitedNodes++
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if visitedNodes != 1 {
		t.Errorf("Expected to visit 1 node (empty directory), but visited %d", visitedNodes)
	}
}

func TestWalker_OnlyFiles(t *testing.T) {
	testFS := fstest.MapFS{
		"file1.txt": &fstest.MapFile{},
		"file2.txt": &fstest.MapFile{},
		"file3.txt": &fstest.MapFile{},
	}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	var visitedFiles int
	err = w.Walk([]string{"."}, func(w *Walker, node *Node) error {
		if node.Type == FileNode {
			visitedFiles++
		}
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if visitedFiles != 3 {
		t.Errorf("Expected to visit 3 files, but visited %d", visitedFiles)
	}
}

func TestWalker_OnlyDirectories(t *testing.T) {
	testFS := fstest.MapFS{
		"dir1/":     &fstest.MapFile{Mode: os.ModeDir},
		"dir2/":     &fstest.MapFile{Mode: os.ModeDir},
		"dir3/":     &fstest.MapFile{Mode: os.ModeDir},
		"dir1/dir4": &fstest.MapFile{Mode: os.ModeDir},
	}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	var visitedDirs int
	err = w.Walk([]string{"."}, func(w *Walker, node *Node) error {
		if node.Type == DirectoryNode {
			visitedDirs++
		}
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if visitedDirs != 5 { // 4 directories + root
		t.Errorf("Expected to visit 5 directories, but visited %d", visitedDirs)
	}
}

func TestWalker_DeeplyNested(t *testing.T) {
	testFS := fstest.MapFS{}
	depth := 100
	path := ""
	for i := 0; i < depth; i++ {
		path = filepath.Join(path, fmt.Sprintf("dir%d", i))
		testFS[path] = &fstest.MapFile{Mode: os.ModeDir}
	}
	testFS[filepath.Join(path, "file.txt")] = &fstest.MapFile{}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	allPaths := make([]string, 0)
	var maxDepth int
	err = w.Walk([]string{"."}, func(w *Walker, node *Node) error {
		currentDepth := len(strings.Split(node.Path, string(os.PathSeparator)))
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}
		allPaths = append(allPaths, node.Path)
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// one for the root, one for file.txt
	if maxDepth != depth+2 { // +1 for the file at the deepest level
		t.Errorf("Expected max depth of %d, but got %d", depth+1, maxDepth)
	}
}

func TestWalker_SpecialCharacters(t *testing.T) {
	testFS := fstest.MapFS{
		"file with spaces.txt":     &fstest.MapFile{},
		"file_with_underscore.txt": &fstest.MapFile{},
		"file-with-dashes.txt":     &fstest.MapFile{},
		"file_with_!@#$%^&*().txt": &fstest.MapFile{},
		"dir with spaces/":         &fstest.MapFile{Mode: os.ModeDir},
		"dir_with_!@#$%^&*()/":     &fstest.MapFile{Mode: os.ModeDir},
	}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	var visitedNodes int
	err = w.Walk([]string{"."}, func(w *Walker, node *Node) error {
		visitedNodes++
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	expectedNodes := len(testFS) + 1 // +1 for root directory
	if visitedNodes != expectedNodes {
		t.Errorf("Expected to visit %d nodes, but visited %d", expectedNodes, visitedNodes)
	}
}
