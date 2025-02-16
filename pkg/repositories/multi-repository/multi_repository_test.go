package multi_repository

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/assert"
)

func createTestCommand(name string, parents []string) cmds.Command {
	return &cmds.CommandDescription{
		Name:    name,
		Parents: parents,
	}
}

func TestNewMultiRepository(t *testing.T) {
	mr := NewMultiRepository()
	assert.NotNil(t, mr)
	assert.Empty(t, mr.repositories)
}

func TestMount(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "root mount",
			path:     "/",
			expected: "/",
		},
		{
			name:     "clean path",
			path:     "test//path/",
			expected: "/test/path",
		},
		{
			name:     "relative path",
			path:     "test",
			expected: "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()
			repo := NewMockRepository(nil)
			mr.Mount(tt.path, repo)

			assert.Len(t, mr.repositories, 1)
			assert.Equal(t, tt.expected, mr.repositories[0].Path)
			assert.Equal(t, repo, mr.repositories[0].Repository)
		})
	}
}

func TestUnmount(t *testing.T) {
	tests := []struct {
		name          string
		mountPath     string
		unmountPath   string
		shouldUnmount bool
	}{
		{
			name:          "exact match",
			mountPath:     "/test",
			unmountPath:   "/test",
			shouldUnmount: true,
		},
		{
			name:          "clean paths",
			mountPath:     "test//",
			unmountPath:   "/test",
			shouldUnmount: true,
		},
		{
			name:          "no match",
			mountPath:     "/test",
			unmountPath:   "/other",
			shouldUnmount: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()
			repo := NewMockRepository(nil)
			mr.Mount(tt.mountPath, repo)

			initialLen := len(mr.repositories)
			mr.Unmount(tt.unmountPath)

			if tt.shouldUnmount {
				assert.Len(t, mr.repositories, initialLen-1)
			} else {
				assert.Len(t, mr.repositories, initialLen)
			}
		})
	}
}
