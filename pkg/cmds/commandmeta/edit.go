package commandmeta

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

// EditCommand implements the command to edit the source file of another command.
type EditCommand struct {
	*glazed_cmds.CommandDescription
	commands []glazed_cmds.Command
}

var _ glazed_cmds.BareCommand = (*EditCommand)(nil)

// EditCommandSettings holds the arguments for the edit command.
type EditCommandSettings struct {
	CommandPath string `glazed.parameter:"command-path"`
}

// newEditCommand creates a new EditCommand.
func newEditCommand(allCommands []glazed_cmds.Command) (*EditCommand, error) {
	return &EditCommand{
		commands: allCommands,
		CommandDescription: glazed_cmds.NewCommandDescription(
			"edit",
			glazed_cmds.WithShort("Edit the source file of a command"),
			glazed_cmds.WithArguments(
				parameters.NewParameterDefinition(
					"command-path",
					parameters.ParameterTypeString,
					parameters.WithHelp("Full path of the command to edit (e.g., 'query es')"),
					parameters.WithRequired(true),
				),
			),
		),
	}, nil
}

// Run executes the edit command.
func (c *EditCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &EditCommandSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	var matchedCommand glazed_cmds.Command
	for _, cmd := range c.commands {
		// Match using FullPath() for clarity
		if cmd.Description().FullPath() == s.CommandPath {
			matchedCommand = cmd
			break
		}
	}

	if matchedCommand == nil {
		// Suggest similar commands? Maybe too complex for now.
		return fmt.Errorf("command not found: %s", s.CommandPath)
	}

	source := matchedCommand.Description().Source
	// Currently only support editing commands loaded from files.
	if !strings.HasPrefix(source, "file:") {
		return fmt.Errorf("cannot edit command '%s': source is not a local file ('%s')", s.CommandPath, source)
	}

	filePath := strings.TrimPrefix(source, "file:")
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to get absolute path for '%s'", filePath)
	}

	// Check if file exists before trying to edit
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		return fmt.Errorf("cannot edit command '%s': source file not found ('%s')", s.CommandPath, absFilePath)
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat source file '%s'", absFilePath)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fallback to common editors
		if _, err := exec.LookPath("vim"); err == nil {
			editor = "vim"
		} else if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else {
			return errors.New("cannot edit command: EDITOR environment variable not set and 'vim' or 'nano' not found")
		}
	}

	// #nosec G204 -- User intends to run their configured editor on a path derived from command metadata
	cmd := exec.CommandContext(ctx, editor, absFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Opening %s in %s...", absFilePath, editor)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to open file '%s' in editor '%s'", absFilePath, editor)
	}

	fmt.Printf("Editor closed for %s.\n", absFilePath)
	return nil
}
