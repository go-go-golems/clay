# Refactoring Command Management with `commandmeta`

This document explains how to replace the separate `ls-commands`, `edit-command`, and `filter-command` setups in your Cobra application with the unified `commandmeta` package provided by `clay`.

**Motivation:**

The goal is to consolidate command listing, filtering, and editing functionalities into a single, consistent, and reusable command group, typically accessed via `your-app commands list` and `your-app commands edit`. This simplifies the main application setup code and leverages the improved filtering capabilities of the underlying `clay/pkg/filters/command` package.

## Before Refactoring

Previously, you likely had code similar to this in your main application setup (`main.go` or `cmd/root.go`), where you loaded all commands and then individually created and added the `ls`, `edit`, and `filter` commands:

```go
// Example from sqleton/cmd/sqleton/main.go (Simplified)
package main

import (
	// ... other imports
	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cli"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"

	// Old command management imports
	edit_command "github.com/go-go-golems/clay/pkg/cmds/edit-command"
	ls_commands "github.com/go-go-golems/clay/pkg/cmds/ls-commands"
	filter_command "github.com/go-go-golems/clay/pkg/filters/command/cmds"

	// Application specific command types (e.g., sqleton)
	sqleton_cmds "github.com/go-go-golems/sqleton/pkg/cmds"
)

func initAllCommands(helpSystem *help.HelpSystem) error {
	// ... load repositories and allCommands ...
	allCommands, err := repositories.LoadRepositories(
		helpSystem,
		rootCmd, /* ... */
	)
	if err != nil {
		return err
	}

	// *** Old Setup Start ***

	// Create a separate group (optional)
	commandsGroup := &cobra.Command{
		Use:   "commands",
		Short: "Commands for managing and filtering sqleton commands",
	}
	rootCmd.AddCommand(commandsGroup)

	// Add List command
	queriesCommand, err := ls_commands.NewListCommandsCommand(allCommands,
		ls_commands.WithCommandDescriptionOptions(
			glazed_cmds.WithShort("Commands related to sqleton queries"),
		),
		ls_commands.WithAddCommandToRowFunc(func(
			command glazed_cmds.Command,
			row types.Row,
			parsedLayers *glazed_layers.ParsedLayers,
		) ([]types.Row, error) {
			ret := []types.Row{row}
			switch c := command.(type) {
			case *sqleton_cmds.SqlCommand:
				row.Set("query", c.Query)
				row.Set("type", "sql")
			default:
			}
			return ret, nil
		}),
	)
	if err != nil {
		return err
	}
	cobraQueriesCommand, err := cli.BuildCobraCommandFromGlazeCommand(queriesCommand) // Assuming Glaze command
	if err != nil {
		return err
	}
	commandsGroup.AddCommand(cobraQueriesCommand) // Added to separate group
	// Note: Sometimes ls-commands might be added directly to rootCmd

	// Add Edit command
	editCommandCommand, err := edit_command.NewEditCommand(allCommands)
	if err != nil {
		return err
	}
	cobraEditCommandCommand, err := cli.BuildCobraCommandFromBareCommand(editCommandCommand) // Assuming Bare command
	if err != nil {
		return err
	}
	commandsGroup.AddCommand(cobraEditCommandCommand) // Added to separate group

	// Add Filter command
	filterCommand, err := filter_command.NewFilterCommand(convertCommandsToDescriptions(allCommands))
	if err != nil {
		return err
	}
	cobraFilterCommand, err := cli.BuildCobraCommandFromGlazeCommand(filterCommand)
	if err != nil {
		return err
	}
	commandsGroup.AddCommand(cobraFilterCommand)

	// *** Old Setup End ***

	// ... other command initializations ...

	return nil
}

// Helper function (now redundant)
func convertCommandsToDescriptions(commands []glazed_cmds.Command) []*glazed_cmds.CommandDescription {
	descriptions := make([]*glazed_cmds.CommandDescription, 0, len(commands))
	for _, cmd := range commands {
		if _, ok := cmd.(*alias.CommandAlias); ok {
			continue
		}
		descriptions = append(descriptions, cmd.Description())
	}
	return descriptions
}

```

## After Refactoring

Replace the old setup block with a single call to `clay_commandmeta.NewCommandManagementCommandGroup`:

```go
package main

import (
	// ... other imports
	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cli"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"

	// New command management import
	clay_commandmeta "github.com/go-go-golems/clay/pkg/cmds/commandmeta"

	// Application specific command types (e.g., sqleton)
	sqleton_cmds "github.com/go-go-golems/sqleton/pkg/cmds"
)

func initAllCommands(helpSystem *help.HelpSystem) error {
	// ... load repositories and allCommands ...
	allCommands, err := repositories.LoadRepositories(
		helpSystem,
		rootCmd, /* ... */
	)
	if err != nil {
		return err
	}

	// *** New Setup Start ***

	// Create and add the unified command management group
	commandManagementCmd, err := clay_commandmeta.NewCommandManagementCommandGroup(
		allCommands, // Pass the loaded commands
		// Pass the existing AddCommandToRowFunc logic (if any) as an option
		clay_commandmeta.WithListAddCommandToRowFunc(func(
			command glazed_cmds.Command,
			row types.Row,
			parsedLayers *glazed_layers.ParsedLayers,
		) ([]types.Row, error) {
			// Example: Set 'type' and 'query' based on command type
			switch c := command.(type) {
			case *sqleton_cmds.SqlCommand:
				row.Set("query", c.Query)
				row.Set("type", "sql")
			case *alias.CommandAlias: // Handle aliases if needed
				row.Set("type", "alias")
				row.Set("aliasFor", c.AliasFor)
			default:
				// Default type handling if needed
				if _, ok := row.Get("type"); !ok {
					row.Set("type", "unknown")
				}
			}
			return []types.Row{row}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize command management commands: %w", err)
	}
	rootCmd.AddCommand(commandManagementCmd) // Add the group directly to root

	// *** New Setup End ***

	// ... other command initializations ...

	return nil
}

// The convertCommandsToDescriptions function is no longer needed.

```

## Key Changes Summarized

1.  **Remove Imports:** Delete imports for `github.com/go-go-golems/clay/pkg/cmds/edit-command`, `.../ls-commands`, and `.../filters/command/cmds`.
2.  **Add Import:** Add `clay_commandmeta "github.com/go-go-golems/clay/pkg/cmds/commandmeta"`.
3.  **Replace Code Block:** Remove the entire code section that created and added the individual `ls`, `edit`, and `filter` commands (and potentially the separate `commandsGroup`).
4.  **Add `NewCommandManagementCommandGroup` Call:** Insert the call to `clay_commandmeta.NewCommandManagementCommandGroup`, passing `allCommands`.
5.  **Pass Hook (Optional):** If you were using `ls_commands.WithAddCommandToRowFunc`, move that function literal into the `clay_commandmeta.WithListAddCommandToRowFunc` option.
6.  **Add to Root:** Add the returned `commandManagementCmd` directly to your `rootCmd`.
7.  **Remove Helper:** Delete the `convertCommandsToDescriptions` function if it exists; `commandmeta` handles this internally.

## Benefits

- **Simpler Code:** Reduces boilerplate in your main setup function.
- **Consistency:** Provides a standard `commands` group (`list`, `edit`) across applications using `clay`.
- **Improved Filtering:** The new `commands list` subcommand uses a more powerful and efficient filtering backend.
