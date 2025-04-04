package edit_command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
)

type EditCommand struct {
	*glazed_cmds.CommandDescription
	commands []glazed_cmds.Command
}
type EditCommandOption func(*EditCommand) error

func WithCommandDescriptionOptions(options ...glazed_cmds.CommandDescriptionOption) EditCommandOption {
	return func(q *EditCommand) error {
		description := q.Description()
		for _, option := range options {
			option(description)
		}
		return nil
	}
}

func NewEditCommand(allCommands []glazed_cmds.Command, options ...EditCommandOption) (*EditCommand, error) {
	ret := &EditCommand{
		commands: allCommands,
		CommandDescription: glazed_cmds.NewCommandDescription(
			"edit-command",
			glazed_cmds.WithShort("Edit a command"),
			glazed_cmds.WithArguments(
				parameters.NewParameterDefinition(
					"command",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the command to edit"),
				),
			),
		),
	}

	for _, option := range options {
		if err := option(ret); err != nil {
			return nil, errors.Wrap(err, "failed to apply option")
		}
	}

	return ret, nil
}

type EditCommandsSettings struct {
	Command string `glazed.parameter:"command"`
}

func (c *EditCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &EditCommandsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to unmarshal settings")

	}

	var matchedCommand glazed_cmds.Command
	for _, cmd := range c.commands {
		if cmd.Description().FullPath() == s.Command {
			matchedCommand = cmd
			break
		}
	}

	if matchedCommand == nil {
		return fmt.Errorf("command not found: %s", s.Command)
	}

	source := matchedCommand.Description().Source
	if !strings.HasPrefix(source, "file:") {
		return fmt.Errorf("unsupported command source: %s", source)
	}

	filePath := strings.TrimPrefix(source, "file:")
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to get absolute file path")
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, absFilePath) // #nosec G204
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to open file in editor")
	}

	return nil
}

var _ glazed_cmds.BareCommand = (*EditCommand)(nil)
