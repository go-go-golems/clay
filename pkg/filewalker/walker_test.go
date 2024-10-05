package filewalker

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestWalker_Walk(t *testing.T) {
	// Create an in-memory file system for testing
	testFS := fstest.MapFS{
		"file1.txt":                 &fstest.MapFile{},
		"file2.txt":                 &fstest.MapFile{},
		"subdir1/file3.txt":         &fstest.MapFile{},
		"subdir1/subdir2/file4.txt": &fstest.MapFile{},
	}

	tests := []struct {
		name          string
		rootPath      string
		expectedFiles int
		expectedDirs  int
	}{
		{
			name:          "Walk entire directory",
			rootPath:      ".",
			expectedFiles: 4,
			expectedDirs:  3, // root, subdir1, subdir1/subdir2
		},
		{
			name:          "Walk subdirectory",
			rootPath:      "subdir1",
			expectedFiles: 2,
			expectedDirs:  2, // subdir1, subdir1/subdir2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWalker(WithFS(testFS))
			if err != nil {
				t.Fatalf("Failed to create Walker: %v", err)
			}
			var filesCount, dirsCount int

			err = w.Walk([]string{tt.rootPath}, func(w *Walker, node *Node) error {
				if node.Type == FileNode {
					filesCount++
				} else {
					dirsCount++
				}
				return nil
			}, nil)

			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}

			if filesCount != tt.expectedFiles {
				t.Errorf("Expected %d files, got %d", tt.expectedFiles, filesCount)
			}

			if dirsCount != tt.expectedDirs {
				t.Errorf("Expected %d directories, got %d", tt.expectedDirs, dirsCount)
			}
		})
	}
}

func TestWalker_GetNodeByPath(t *testing.T) {
	_, w, cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name          string
		path          string
		expectedType  NodeType
		expectedError bool
	}{
		{
			name:          "Get existing file",
			path:          "file1.txt",
			expectedType:  FileNode,
			expectedError: false,
		},
		{
			name:          "Get existing directory",
			path:          "subdir1",
			expectedType:  DirectoryNode,
			expectedError: false,
		},
		{
			name:          "Get non-existent file",
			path:          "nonexistent.txt",
			expectedType:  FileNode,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := w.GetNodeByPath(tt.path)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if node.Type != tt.expectedType {
					t.Errorf("Expected node type %v, but got %v", tt.expectedType, node.Type)
				}
				if node.Path != tt.path {
					t.Errorf("Expected path %s, but got %s", tt.path, node.Path)
				}
			}
		})
	}
}

func TestWalker_GetNodeByRelativePath(t *testing.T) {
	_, w, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Get the root node
	rootNode, err := w.GetNodeByPath(".")
	if err != nil {
		t.Fatalf("Failed to get root node: %v", err)
	}

	tests := []struct {
		name          string
		relativePath  string
		expectedType  NodeType
		expectedError bool
	}{
		{
			name:          "Get existing file",
			relativePath:  "file1.txt",
			expectedType:  FileNode,
			expectedError: false,
		},
		{
			name:          "Get existing directory",
			relativePath:  "subdir1",
			expectedType:  DirectoryNode,
			expectedError: false,
		},
		{
			name:          "Get file in subdirectory",
			relativePath:  filepath.Join("subdir1", "file3.txt"),
			expectedType:  FileNode,
			expectedError: false,
		},
		{
			name:          "Get non-existent file",
			relativePath:  "nonexistent.txt",
			expectedType:  FileNode,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := w.GetNodeByRelativePath(rootNode, tt.relativePath)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if node.Type != tt.expectedType {
					t.Errorf("Expected node type %v, but got %v", tt.expectedType, node.Type)
				}
				expectedPath := filepath.Join(".", tt.relativePath)
				if node.Path != expectedPath {
					t.Errorf("Expected path %s, but got %s", expectedPath, node.Path)
				}
			}
		})
	}
}

// Helper function to set up the test environment
func setupTestEnvironment(t *testing.T) (fstest.MapFS, *Walker, func()) {
	testFS := fstest.MapFS{
		"file1.txt":                 &fstest.MapFile{},
		"file2.txt":                 &fstest.MapFile{},
		"subdir1/file3.txt":         &fstest.MapFile{},
		"subdir1/subdir2/file4.txt": &fstest.MapFile{},
	}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}
	err = w.Walk([]string{"."}, nil, nil)
	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	cleanup := func() {
		// No cleanup needed for in-memory file system
	}

	return testFS, w, cleanup
}

func TestWalker_FollowSymlinks(t *testing.T) {
	tempDir, cleanup := setupSymlinkTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name           string
		followSymlinks bool
		expectedFiles  int
		expectedDirs   int
	}{
		{
			name:           "Follow symlinks",
			followSymlinks: true,
			expectedFiles:  7, // 4 regular files + 1 symlinked file + 1 symlinked dir
			expectedDirs:   5, // root, subdir1, subdir1/subdir2, symlinked_dir
		},
		{
			name:           "Don't follow symlinks",
			followSymlinks: false,
			expectedFiles:  4, // 4 regular files
			expectedDirs:   3, // root, subdir1, subdir1/subdir2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWalker(WithFS(os.DirFS(tempDir)), WithFollowSymlinks(tt.followSymlinks))
			if err != nil {
				t.Fatalf("Failed to create Walker: %v", err)
			}
			var filesCount, dirsCount int
			var filesList, dirsList []string

			err = w.Walk([]string{"."}, func(w *Walker, node *Node) error {
				if node.Type == FileNode {
					filesCount++
					filesList = append(filesList, node.Path)
				} else {
					dirsCount++
					dirsList = append(dirsList, node.Path)
				}
				return nil
			}, nil)

			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}

			if filesCount != tt.expectedFiles {
				t.Errorf("Expected %d files, got %d", tt.expectedFiles, filesCount)
			}

			if dirsCount != tt.expectedDirs {
				t.Errorf("Expected %d directories, got %d", tt.expectedDirs, dirsCount)
			}

			//t.Logf("Files found: %v", filesList)
			//t.Logf("Directories found: %v", dirsList)
		})
	}
}

// Helper function to set up a test environment with symlinks
func setupSymlinkTestEnvironment(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "filewalker_symlink_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	testFiles := map[string]string{
		"file1.txt":                 "content1",
		"file2.txt":                 "content2",
		"subdir1/file3.txt":         "content3",
		"subdir1/subdir2/file4.txt": "content4",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create a symlink to a file
	if err := os.Symlink(filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "symlink_file.txt")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create a symlink to a directory
	if err := os.Symlink(filepath.Join(tempDir, "subdir1"), filepath.Join(tempDir, "symlink_dir")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestWalker_MultipleRootPaths(t *testing.T) {
	paths := []string{
		"file1.txt",
		"file2.txt",
		"subdir1/file3.txt",
		"subdir1/subdir2/file4.txt",
	}

	w, err := NewWalker(WithPaths(paths))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	var visitedPaths []string

	err = w.Walk([]string{"."}, func(w *Walker, node *Node) error {
		if node.Type == FileNode {
			visitedPaths = append(visitedPaths, node.Path)
		}
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	expectedPaths := paths

	if len(visitedPaths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(visitedPaths))
	}

	for _, expected := range expectedPaths {
		found := false
		for _, visited := range visitedPaths {
			if visited == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected path %s not found in visited paths", expected)
		}
	}
}

func TestWalker_PrePostVisitFunctions(t *testing.T) {
	_, w, cleanup := setupTestEnvironment(t)
	defer cleanup()

	var preVisitOrder, postVisitOrder []string

	preVisit := func(w *Walker, node *Node) error {
		preVisitOrder = append(preVisitOrder, node.Path)
		return nil
	}

	postVisit := func(w *Walker, node *Node) error {
		postVisitOrder = append(postVisitOrder, node.Path)
		return nil
	}

	err := w.Walk([]string{"."}, preVisit, postVisit)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	expectedPreOrder := []string{
		".",
		"file1.txt",
		"file2.txt",
		"subdir1",
		"subdir1/file3.txt",
		"subdir1/subdir2",
		"subdir1/subdir2/file4.txt",
	}

	expectedPostOrder := []string{
		"file1.txt",
		"file2.txt",
		"subdir1/file3.txt",
		"subdir1/subdir2/file4.txt",
		"subdir1/subdir2",
		"subdir1",
		".",
	}

	if !stringSlicesEqual(preVisitOrder, expectedPreOrder) {
		t.Errorf("Pre-visit order incorrect.\nExpected: %v\nGot: %v", expectedPreOrder, preVisitOrder)
	}

	if !stringSlicesEqual(postVisitOrder, expectedPostOrder) {
		t.Errorf("Post-visit order incorrect.\nExpected: %v\nGot: %v", expectedPostOrder, postVisitOrder)
	}
}

// Helper function to compare string slices
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
