package repositories

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
	"os"
)

func LoadCommandsFromInputs(
	commandLoader loaders.CommandLoader,
	inputs []string,
) ([]cmds.Command, error) {
	files := []string{}
	directories := []Directory{}
	for _, input := range inputs {
		// check if is directory
		s, err := os.Stat(input)
		if err != nil {
			return nil, err
		}
		if s.IsDir() {
			directories = append(directories, Directory{
				FS:               os.DirFS(input),
				RootDirectory:    ".",
				RootDocDirectory: "doc",
				Name:             input,
				WatchDirectory:   input,
				SourcePrefix:     "file",
			})
		} else {
			files = append(files, input)
		}
	}

	repository := NewRepository(
		WithCommandLoader(commandLoader),
		WithDirectories(directories...),
	)

	helpSystem := help.NewHelpSystem()
	// TODO(manuel, 2024-01-20) We actually need to load the files here as well to properly do the alias resolution
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

func LoadRepositories(
	helpSystem *help.HelpSystem,
	rootCmd *cobra.Command,
	repositories_ []*Repository,
	options ...cli.CobraParserOption,
) ([]cmds.Command, error) {

	allCommands := []cmds.Command{}

	for _, repository := range repositories_ {
		err := repository.LoadCommands(helpSystem)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error initializing commands: %s\n", err)
			return nil, err
		}

		aliases := []*alias.CommandAlias{}
		commands := []cmds.Command{}
		commands_ := repository.CollectCommands([]string{}, true)

		for _, command := range commands_ {
			switch v := command.(type) {
			case *alias.CommandAlias:
				aliases = append(aliases, v)
			case cmds.Command:
				commands = append(commands, v)
			}
		}

		allCommands = append(allCommands, commands_...)

		err = cli.AddCommandsToRootCommand(
			rootCmd, commands, aliases,
			options...,
		)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error initializing commands: %s\n", err)
			return nil, err
		}

	}
	return allCommands, nil
}
