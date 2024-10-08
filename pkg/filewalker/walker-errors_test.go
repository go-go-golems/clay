package filewalker

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"
)

func TestWalker_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setupFS     func() fs.FS
		walkPath    string
		expectedErr string
	}{
		{
			name: "Non-existent path",
			setupFS: func() fs.FS {
				return fstest.MapFS{}
			},
			walkPath:    "non_existent",
			expectedErr: "file does not exist",
		},
		{
			name: "readdir restricted: not implemented",
			setupFS: func() fs.FS {
				return &mockFS{
					files: map[string]*mockFile{
						"restricted": {
							isDir: true,
							mode:  0000,
						},
					},
				}
			},
			walkPath:    "restricted",
			expectedErr: "readdir restricted: not implemented",
		},
		{
			name: "I/O error during file reading",
			setupFS: func() fs.FS {
				return &mockFS{
					files: map[string]*mockFile{
						"error_file.txt": {
							isDir:    false,
							readErr:  errors.New("I/O error"),
							contents: []byte("test"),
						},
					},
				}
			},
			walkPath:    "error_file.txt",
			expectedErr: "I/O error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWalker(WithFS(tt.setupFS()))
			if err != nil {
				t.Fatalf("Failed to create Walker: %v", err)
			}

			err = w.Walk([]string{tt.walkPath}, nil, nil)
			if err == nil {
				t.Fatalf("Expected error, but got nil")
			}
			if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, fs.ErrPermission) && err.Error() != tt.expectedErr {
				t.Errorf("Expected error containing %q, but got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestWalker_NonExistentNodeRetrieval(t *testing.T) {
	w, err := NewWalker(WithFS(fstest.MapFS{}))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	_, err = w.GetNodeByPath("non_existent")
	if err == nil {
		t.Fatalf("Expected error when retrieving non-existent node, but got nil")
	}
	expectedErr := "node not found for path: non_existent"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, but got %q", expectedErr, err.Error())
	}
}

func TestWalker_PreVisitError(t *testing.T) {
	testFS := fstest.MapFS{
		"file.txt": &fstest.MapFile{},
	}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	preVisitErr := errors.New("pre-visit error")
	err = w.Walk([]string{"."}, func(w *Walker, node *Node) error {
		return preVisitErr
	}, nil)

	if err == nil {
		t.Fatalf("Expected error from pre-visit function, but got nil")
	}
	if err != preVisitErr {
		t.Errorf("Expected error %q, but got %q", preVisitErr, err)
	}
}

func TestWalker_PostVisitError(t *testing.T) {
	testFS := fstest.MapFS{
		"file.txt": &fstest.MapFile{},
	}

	w, err := NewWalker(WithFS(testFS))
	if err != nil {
		t.Fatalf("Failed to create Walker: %v", err)
	}

	postVisitErr := errors.New("post-visit error")
	err = w.Walk([]string{"."}, nil, func(w *Walker, node *Node) error {
		return postVisitErr
	})

	if err == nil {
		t.Fatalf("Expected error from post-visit function, but got nil")
	}
	if err != postVisitErr {
		t.Errorf("Expected error %q, but got %q", postVisitErr, err)
	}
}

// mockFS and mockFile implementations for custom error scenarios
type mockFS struct {
	files map[string]*mockFile
}

func (m *mockFS) Open(name string) (fs.File, error) {
	f, ok := m.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return f, nil
}

type mockFile struct {
	isDir    bool
	mode     fs.FileMode
	contents []byte
	readErr  error
}

func (m *mockFile) Stat() (fs.FileInfo, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return &mockFileInfo{m}, nil
}

func (m *mockFile) Read(b []byte) (int, error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	n := copy(b, m.contents)
	return n, nil
}

func (m *mockFile) Close() error {
	return nil
}

type mockFileInfo struct {
	file *mockFile
}

func (m *mockFileInfo) Name() string       { return "mock" }
func (m *mockFileInfo) Size() int64        { return int64(len(m.file.contents)) }
func (m *mockFileInfo) Mode() fs.FileMode  { return m.file.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.file.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
