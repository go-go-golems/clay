package commandmeta

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CommandManagementConfig holds configuration options for the command management group.
type CommandManagementConfig struct {
	ListAddCommandToRowFunc AddCommandToRowFunc
}

// Option defines a function signature for configuring CommandManagementConfig.
type Option func(*CommandManagementConfig)

// WithListAddCommandToRowFunc provides an option to set the AddCommandToRowFunc for the list command.
func WithListAddCommandToRowFunc(f AddCommandToRowFunc) Option {
	return func(cfg *CommandManagementConfig) {
		cfg.ListAddCommandToRowFunc = f
	}
}

// NewCommandManagementCommandGroup creates a new Cobra command group for managing commands.
// It includes subcommands for listing/filtering ('list') and editing ('edit').
func NewCommandManagementCommandGroup(
	allCommands []glazed_cmds.Command,
	options ...Option,
) (*cobra.Command, error) {
	cfg := &CommandManagementConfig{
		// Default config values here if needed
	}
	for _, opt := range options {
		opt(cfg)
	}

	// Create the root command for this group
	rootCmd := &cobra.Command{
		Use:   "commands",
		Short: "Manage and inspect available commands",
	}

	// Extract descriptions for indexing (used by list command)
	descriptions := make([]*glazed_cmds.CommandDescription, len(allCommands))
	for i, cmd := range allCommands {
		descriptions[i] = cmd.Description()
	}

	// Create and add the 'list' subcommand
	listCmd, err := newListCommand(allCommands, descriptions, cfg.ListAddCommandToRowFunc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create list command")
	}
	listCobraCmd, err := cli.BuildCobraCommand(listCmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build list cobra command")
	}
	rootCmd.AddCommand(listCobraCmd)

	// Create and add the 'edit' subcommand
	editCmd, err := newEditCommand(allCommands)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create edit command")
	}
	editCobraCmd, err := cli.BuildCobraCommand(editCmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build edit cobra command")
	}
	rootCmd.AddCommand(editCobraCmd)

	return rootCmd, nil
}
