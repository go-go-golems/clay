package repositories

import (
	"embed"
	"fmt"
	"path/filepath"

	yaml_editor "github.com/go-go-golems/clay/pkg/yaml-editor"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRepositoriesGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repositories",
		Short: "Manage repositories in the configuration",
	}

	cmd.AddCommand(NewAddRepositoryCommand())
	cmd.AddCommand(NewRemoveRepositoryCommand())
	cmd.AddCommand(NewPrintRepositoriesCommand())

	return cmd
}

func NewAddRepositoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [directories...]",
		Short: "Add directories to the repository entry in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one directory must be provided")
			}

			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				return fmt.Errorf("no config file found")
			}

			// Create a new YAML editor from the config file
			editor, err := yaml_editor.NewYAMLEditorFromFile(configFile)
			if err != nil {
				return fmt.Errorf("error creating YAML editor: %w", err)
			}

			// Get or create the repositories sequence node
			repoNode, err := editor.GetNode("repositories")
			if err != nil {
				// Create a new sequence node if it doesn't exist
				repoNode = editor.CreateSequence()
				err = editor.SetNode(repoNode, "repositories")
				if err != nil {
					return fmt.Errorf("error creating repositories node: %w", err)
				}
			}

			added := false

			// Append new directories
			for _, dir := range args {
				absDir, err := filepath.Abs(dir)
				if err != nil {
					return fmt.Errorf("error getting absolute path for %s: %w", dir, err)
				}

				// Check if the repository already exists
				exists := false
				for _, node := range repoNode.Content {
					if node.Value == absDir {
						exists = true
						break
					}
				}

				if exists {
					fmt.Printf("Repository %s already exists in the list. Skipping.\n", absDir)
					continue
				}

				fmt.Printf("Adding %s to repository list.\n", absDir)
				err = editor.AppendToSequence(editor.CreateScalar(absDir), "repositories")
				if err != nil {
					return fmt.Errorf("error appending repository: %w", err)
				}
				added = true
			}

			// Save the updated config
			if err := editor.Save(configFile); err != nil {
				return fmt.Errorf("error saving config file: %w", err)
			}

			// Print out the total list of repositories
			if added {
				fmt.Println("\nCurrent repository list:")
				for _, node := range repoNode.Content {
					fmt.Printf("- %s\n", node.Value)
				}
			}

			return nil
		},
	}

	return cmd
}

func NewRemoveRepositoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [directories...]",
		Short: "Remove directories from the repository entry in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one directory must be provided")
			}

			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				return fmt.Errorf("no config file found")
			}

			// Create a new YAML editor from the config file
			editor, err := yaml_editor.NewYAMLEditorFromFile(configFile)
			if err != nil {
				return fmt.Errorf("error creating YAML editor: %w", err)
			}

			// Get the repositories sequence node
			repoNode, err := editor.GetNode("repositories")
			if err != nil {
				return fmt.Errorf("repositories node not found: %w", err)
			}

			removed := false
			for _, dir := range args {
				absDir, err := filepath.Abs(dir)
				if err != nil {
					return fmt.Errorf("error getting absolute path for %s: %w", dir, err)
				}

				// Find and remove the repository
				for i, node := range repoNode.Content {
					if node.Value == absDir {
						err = editor.RemoveFromSequence(i, "repositories")
						if err != nil {
							return fmt.Errorf("error removing repository: %w", err)
						}
						fmt.Printf("Removed %s from repository list.\n", absDir)
						removed = true
						break
					}
				}

				if !removed {
					fmt.Printf("Repository %s not found in the list. Skipping.\n", absDir)
				}
			}

			if removed {
				// Save the updated config
				if err := editor.Save(configFile); err != nil {
					return fmt.Errorf("error saving config file: %w", err)
				}

				fmt.Println("\nUpdated repository list:")
				repoNode, _ = editor.GetNode("repositories")
				for _, node := range repoNode.Content {
					fmt.Printf("- %s\n", node.Value)
				}
			}

			return nil
		},
	}
	return cmd
}

func NewPrintRepositoriesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Print the list of repositories in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				return fmt.Errorf("no config file found")
			}

			// Create a new YAML editor from the config file
			editor, err := yaml_editor.NewYAMLEditorFromFile(configFile)
			if err != nil {
				return fmt.Errorf("error creating YAML editor: %w", err)
			}

			// Get the repositories sequence node
			repoNode, err := editor.GetNode("repositories")
			if err != nil {
				return fmt.Errorf("repositories node not found: %w", err)
			}

			// Print repositories
			for _, node := range repoNode.Content {
				fmt.Printf("- %s\n", node.Value)
			}

			return nil
		},
	}
	return cmd
}

//go:embed docs/*
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
