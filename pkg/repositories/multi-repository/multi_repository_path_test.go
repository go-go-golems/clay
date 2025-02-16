package multi_repository

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/assert"
)

func TestMountPathNormalization(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "clean root path",
			path:     "////",
			expected: "/",
		},
		{
			name:     "clean relative path",
			path:     "test/path///",
			expected: "/test/path",
		},
		{
			name:     "clean absolute path",
			path:     "/test//path/",
			expected: "/test/path",
		},
		{
			name:     "dot segments",
			path:     "/test/./path/../path",
			expected: "/test/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()
			repo := NewMockRepository(nil)
			mr.Mount(tt.path, repo)

			assert.Len(t, mr.repositories, 1)
			assert.Equal(t, tt.expected, mr.repositories[0].Path)
		})
	}
}

func TestNestedMountPaths(t *testing.T) {
	tests := []struct {
		name          string
		mounts        []string
		findPath      string
		shouldFind    bool
		expectedMount string
	}{
		{
			name:          "exact nested match",
			mounts:        []string{"/test", "/test/nested"},
			findPath:      "/test/nested/cmd",
			shouldFind:    true,
			expectedMount: "/test/nested",
		},
		{
			name:          "parent match",
			mounts:        []string{"/test", "/test/nested"},
			findPath:      "/test/cmd",
			shouldFind:    true,
			expectedMount: "/test",
		},
		{
			name:          "root vs nested",
			mounts:        []string{"/", "/test"},
			findPath:      "/test/cmd",
			shouldFind:    true,
			expectedMount: "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()

			// Mount repositories in reverse order to test proper path matching
			for i := len(tt.mounts) - 1; i >= 0; i-- {
				mockRepo := NewMockRepository([]cmds.Command{
					createTestCommand("cmd", nil),
				})
				mr.Mount(tt.mounts[i], mockRepo)
			}

			cmd, found := mr.GetCommand(tt.findPath)
			assert.Equal(t, tt.shouldFind, found)

			if tt.shouldFind {
				// Verify the command came from the correct repository
				for _, mounted := range mr.repositories {
					if mounted.Path == tt.expectedMount {
						mockRepo := mounted.Repository.(*MockRepository)
						assert.Contains(t, mockRepo.commands, cmd)
						break
					}
				}
			}
		})
	}
}
