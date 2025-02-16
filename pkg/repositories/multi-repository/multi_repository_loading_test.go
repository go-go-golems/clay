package multi_repository

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/stretchr/testify/assert"
)

func TestLoadCommands(t *testing.T) {
	tests := []struct {
		name  string
		repos map[string]struct {
			commands []cmds.Command
			err      error
		}
		wantErr bool
	}{
		{
			name: "single repo success",
			repos: map[string]struct {
				commands []cmds.Command
				err      error
			}{
				"/": {
					commands: []cmds.Command{
						createTestCommand("test", nil),
					},
					err: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "multiple repos success",
			repos: map[string]struct {
				commands []cmds.Command
				err      error
			}{
				"/test1": {
					commands: []cmds.Command{
						createTestCommand("test1", nil),
					},
					err: nil,
				},
				"/test2": {
					commands: []cmds.Command{
						createTestCommand("test2", nil),
					},
					err: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "repo with error",
			repos: map[string]struct {
				commands []cmds.Command
				err      error
			}{
				"/": {
					commands: nil,
					err:      assert.AnError,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()
			helpSystem := help.NewHelpSystem()

			for path, repo := range tt.repos {
				mockRepo := NewMockRepository(repo.commands)
				mockRepo.loadError = repo.err
				mr.Mount(path, mockRepo)
			}

			err := mr.LoadCommands(helpSystem)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify each mock repository was called with the help system
				for _, mounted := range mr.repositories {
					mockRepo := mounted.Repository.(*MockRepository)
					assert.Equal(t, helpSystem, mockRepo.helpSystem)
				}
			}
		})
	}
}
