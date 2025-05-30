package multi_repository

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/assert"
)

func TestCollectCommands(t *testing.T) {
	tests := []struct {
		name     string
		repos    map[string][]cmds.Command
		prefix   []string
		recurse  bool
		expected []string // command full paths
	}{
		{
			name: "root collection - no prefix",
			repos: map[string][]cmds.Command{
				"/": {
					createTestCommand("test1", nil),
					createTestCommand("test2", nil),
				},
			},
			prefix:   []string{},
			recurse:  true,
			expected: []string{"test1", "test2"}, // No leading slash for root mount
		},
		{
			name: "mounted repo collection - with prefix",
			repos: map[string][]cmds.Command{
				"/test": {
					createTestCommand("cmd1", nil),
					createTestCommand("cmd2", nil),
				},
			},
			prefix:   []string{"test"},
			recurse:  true,
			expected: []string{"test/cmd1", "test/cmd2"}, // Keep prefix for non-root mounts
		},
		{
			name: "multiple repos",
			repos: map[string][]cmds.Command{
				"/test1": {
					createTestCommand("cmd1", nil),
				},
				"/test2": {
					createTestCommand("cmd2", nil),
				},
			},
			prefix:   []string{},
			recurse:  true,
			expected: []string{"test1/cmd1", "test2/cmd2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()

			for path, commands := range tt.repos {
				mockRepo := NewMockRepository(commands)
				mr.Mount(path, mockRepo)
			}

			commands := mr.CollectCommands(tt.prefix, tt.recurse)
			var paths []string
			for _, cmd := range commands {
				paths = append(paths, cmd.Description().FullPath())
			}
			assert.ElementsMatch(t, tt.expected, paths)
		})
	}
}

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name          string
		repos         map[string][]cmds.Command
		commandPath   string
		shouldFind    bool
		expectedName  string
		expectedMount string
	}{
		{
			name: "root command",
			repos: map[string][]cmds.Command{
				"/": {
					createTestCommand("test", nil),
				},
			},
			commandPath:   "test",
			shouldFind:    true,
			expectedName:  "test",
			expectedMount: "/",
		},
		{
			name: "mounted command",
			repos: map[string][]cmds.Command{
				"/test": {
					createTestCommand("cmd", nil),
				},
			},
			commandPath:   "/test/cmd",
			shouldFind:    true,
			expectedName:  "cmd",
			expectedMount: "/test",
		},
		{
			name: "command not found",
			repos: map[string][]cmds.Command{
				"/": {
					createTestCommand("test", nil),
				},
			},
			commandPath: "nonexistent",
			shouldFind:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()

			for path, commands := range tt.repos {
				mockRepo := NewMockRepository(commands)
				mr.Mount(path, mockRepo)
			}

			cmd, found := mr.GetCommand(tt.commandPath)
			assert.Equal(t, tt.shouldFind, found)

			if tt.shouldFind {
				assert.Equal(t, tt.expectedName, cmd.Description().Name)
			}
		})
	}
}
