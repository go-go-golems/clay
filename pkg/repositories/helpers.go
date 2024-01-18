package repositories

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/help"
	"os"
)

func LoadCommandsFromInputs(
	commandLoader loaders.CommandLoader,
	inputs []string,
) ([]cmds.Command, error) {
	files := []string{}
	directories := []string{}
	for _, input := range inputs {
		// check if is directory
		s, err := os.Stat(input)
		if err != nil {
			return nil, err
		}
		if s.IsDir() {
			directories = append(directories, input)
		} else {
			files = append(files, input)
		}
	}

	repository := NewRepository(
		WithCommandLoader(commandLoader),
		WithDirectories(directories...),
	)

	helpSystem := help.NewHelpSystem()
	err := repository.LoadCommands(helpSystem)
	if err != nil {
		return nil, err
	}

	commands := repository.CollectCommands([]string{}, true)
	for _, file := range files {
		f, file_, err := loaders.FileNameToFsFilePath(file)
		if err != nil {
			return nil, err
		}

		cmds_, err := commandLoader.LoadCommands(f, file_, []cmds.CommandDescriptionOption{}, []alias.Option{})
		if err != nil {
			return nil, err
		}

		commands = append(commands, cmds_...)
	}

	return commands, nil
}
