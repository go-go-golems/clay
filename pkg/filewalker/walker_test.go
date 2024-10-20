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

func TestWalker_ResolveRelativePath(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	w := &Walker{currentDir: currentDir}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Absolute path",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "Relative path",
			input:    "relative/path",
			expected: filepath.Join(currentDir, "relative/path"),
		},
		{
			name:     "Current directory",
			input:    ".",
			expected: currentDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := w.resolveRelativePath(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Add more tests for other methods as needed...
