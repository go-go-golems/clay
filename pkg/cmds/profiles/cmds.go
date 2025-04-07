package profiles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// InitialContentProvider is a function type that returns the default content for a new profiles file.
type InitialContentProvider func() string

// NewProfilesCommand creates the "profiles" command group for managing application profiles.
// It requires the application name (used for the config directory) and a function
// to provide the initial content for a new profiles file.
func NewProfilesCommand(appName string, initialContentProvider InitialContentProvider) (*cobra.Command, error) {
	cobraCmd := &cobra.Command{
		Use:   "profiles",
		Short: fmt.Sprintf("Manage %s profiles", appName),
	}

	// Helper function to get editor for this app
	getEditor := func() (*ProfilesEditor, error) {
		profilesPath, err := GetProfilesPathForApp(appName)
		if err != nil {
			return nil, fmt.Errorf("could not get profiles path for %s: %w", appName, err)
		}

		log.Debug().Str("profiles_path", profilesPath).Msg("using profiles file")
		editor, err := NewProfilesEditor(profilesPath)
		if err != nil {
			// If the error is because the file doesn't exist, that's okay for some commands (like init, list)
			// but not for others (get, set, delete, duplicate).
			// We will handle this check within each command's RunE where necessary.
			// NewProfilesEditor itself might need refinement to better signal "file not found".
			// For now, return the editor (which might be partially initialized or nil based on NewProfilesEditor impl)
			// and the error.
			return editor, fmt.Errorf("could not create profiles editor for %s: %w", profilesPath, err)
		}

		return editor, nil
	}

	cobraCmd.AddCommand(newListCommand(getEditor))
	cobraCmd.AddCommand(newGetCommand(getEditor))
	cobraCmd.AddCommand(newSetCommand(getEditor))
	cobraCmd.AddCommand(newDeleteCommand(getEditor))
	cobraCmd.AddCommand(newEditCommand(appName))
	cobraCmd.AddCommand(newInitCommand(appName, initialContentProvider))
	cobraCmd.AddCommand(newDuplicateCommand(getEditor))

	return cobraCmd, nil
}

// --- Subcommand implementations --- //

func newListCommand(getEditor func() (*ProfilesEditor, error)) *cobra.Command {
	var concise bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := getEditor()
			// If the file doesn't exist, ListProfiles should handle it gracefully (return empty list).
			if err != nil && !os.IsNotExist(err) { // Allow "not found" errors
				// If NewProfilesEditor failed for a reason other than file not found
				return err
			}
			// If editor is nil (because file not found and NewProfilesEditor returned nil),
			// treat as empty list.
			if editor == nil {
				fmt.Println("No profiles defined.")
				return nil
			}

			profiles, contents, err := editor.ListProfiles()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			if len(profiles) == 0 {
				fmt.Println("No profiles defined.")
				return nil
			}

			if concise {
				for _, profile := range profiles {
					fmt.Println(profile)
				}
				return nil
			}

			// Show full profile contents
			for _, profile := range profiles {
				fmt.Printf("%s:\n", profile)
				layersMap := contents[profile]
				// TODO(manuel, 2024-07-17) Consider sorting layers and settings for consistent output
				// Requires getting ordered maps back from ListProfiles or sorting here.
				for layerName, settings := range layersMap {
					fmt.Printf("  %s:\n", layerName)
					for key, value := range settings {
						fmt.Printf("    %s: %s\n", key, value)
					}
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&concise, "concise", "c", false, "Only show profile names (default: show full content)")
	return cmd
}

func newGetCommand(getEditor func() (*ProfilesEditor, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "get <profile> [layer] [key]",
		Short: "Get profile settings",
		Args:  cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := getEditor()
			if err != nil {
				// Cannot get settings if editor failed to load (e.g., file not found)
				return err
			}

			profile := args[0]

			if len(args) == 1 {
				// Show all layers for the profile
				layers, err := editor.GetProfileLayers(profile)
				if err != nil {
					return err // Error includes profile name context from GetProfileLayers
				}

				if layers.Len() == 0 {
					fmt.Printf("Profile '%s' exists but has no layers defined.\n", profile)
					return nil
				}

				for pair := layers.Oldest(); pair != nil; pair = pair.Next() {
					fmt.Printf("%s:\n", pair.Key)
					settings := pair.Value
					for settingPair := settings.Oldest(); settingPair != nil; settingPair = settingPair.Next() {
						fmt.Printf("  %s: %s\n", settingPair.Key, settingPair.Value)
					}
				}
				return nil
			}

			layer := args[1]
			if len(args) == 2 {
				// Show all settings for a specific layer
				layers, err := editor.GetProfileLayers(profile)
				if err != nil {
					return err
				}

				settings, ok := layers.Get(layer)
				if !ok {
					return fmt.Errorf("layer '%s' not found in profile '%s'", layer, profile)
				}

				if settings.Len() == 0 {
					fmt.Printf("Layer '%s' exists in profile '%s' but has no settings.\n", layer, profile)
					return nil
				}

				for pair := settings.Oldest(); pair != nil; pair = pair.Next() {
					fmt.Printf("%s: %s\n", pair.Key, pair.Value)
				}
				return nil
			}

			// Get a specific value
			key := args[2]
			value, err := editor.GetLayerValue(profile, layer, key)
			if err != nil {
				return err // Error includes context from GetLayerValue
			}

			fmt.Println(value)
			return nil
		},
	}
}

func newSetCommand(getEditor func() (*ProfilesEditor, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "set <profile> <layer> <key> <value>",
		Short: "Set a profile setting (creates profile/layer if needed)",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := getEditor()
			if err != nil && !os.IsNotExist(err) {
				// Allow file not found, as set will create it.
				// If NewProfilesEditor failed for other reasons, return error.
				return err
			}
			// If editor is nil (e.g., file didn't exist), we need to handle this.
			// Let's refine `getEditor` or `NewProfilesEditor` to handle this better.
			// Assuming for now that `NewProfilesEditor` returns a valid editor even if the file didn't exist.
			if editor == nil { // This check might be redundant if NewProfilesEditor guarantees non-nil on non-error
				return fmt.Errorf("internal error: profile editor is nil")
			}

			profile := args[0]
			layer := args[1]
			key := args[2]
			value := args[3]

			if err := editor.SetLayerValue(profile, layer, key, value); err != nil {
				return fmt.Errorf("failed to set value: %w", err)
			}

			if err := editor.Save(); err != nil {
				return fmt.Errorf("failed to save profiles: %w", err)
			}
			fmt.Printf("Set %s.%s.%s = %s\n", profile, layer, key, value)
			return nil
		},
	}
}

func newDeleteCommand(getEditor func() (*ProfilesEditor, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <profile> [layer] [key]",
		Short: "Delete a profile, layer, or setting",
		Args:  cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := getEditor()
			if err != nil {
				// Cannot delete if editor failed to load
				return err
			}

			profile := args[0]
			var deleteErr error
			var deletedItem string

			if len(args) == 1 {
				// Delete entire profile
				deleteErr = editor.DeleteProfile(profile)
				deletedItem = fmt.Sprintf("profile '%s'", profile)
			} else if len(args) == 2 {
				// Delete a layer
				layer := args[1]
				deleteErr = editor.DeleteLayer(profile, layer)
				deletedItem = fmt.Sprintf("layer '%s' in profile '%s'", layer, profile)
			} else if len(args) == 3 {
				// Delete specific setting
				layer := args[1]
				key := args[2]
				deleteErr = editor.DeleteLayerValue(profile, layer, key)
				deletedItem = fmt.Sprintf("setting '%s.%s' in profile '%s'", layer, key, profile)
			} else {
				// Should be caught by Args validation, but good practice to handle
				return fmt.Errorf("unexpected number of arguments: %d", len(args))
			}

			if deleteErr != nil {
				// Check if the error is because the item didn't exist
				// Need a way to check this from yaml-editor. Assuming generic error for now.
				// TODO(manuel, 2024-07-17) Improve error checking for "not found" on delete
				log.Warn().Err(deleteErr).Msgf("Could not delete %s (it might not exist)", deletedItem)
				// Don't return error if item not found, just warn? Or return specific error?
				// Let's return the error for now.
				return fmt.Errorf("failed to delete %s: %w", deletedItem, deleteErr)
			}

			if err := editor.Save(); err != nil {
				return fmt.Errorf("failed to save profiles after deletion: %w", err)
			}

			fmt.Printf("Deleted %s\n", deletedItem)
			return nil
		},
	}
}

func newEditCommand(appName string) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit the profiles file in your default editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			profilesPath, err := GetProfilesPathForApp(appName)
			if err != nil {
				return err
			}

			// Ensure the directory exists before trying to edit the file
			profilesDir := filepath.Dir(profilesPath)
			if err := os.MkdirAll(profilesDir, 0755); err != nil {
				return fmt.Errorf("could not create profiles directory %s: %w", profilesDir, err)
			}

			// If the file doesn't exist, create it empty so the editor opens something.

			if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
				if err := os.WriteFile(profilesPath, []byte{}, 0644); err != nil {
					return fmt.Errorf("could not create empty profiles file %s: %w", profilesPath, err)
				}
				log.Info().Str("path", profilesPath).Msg("Created empty profiles file for editing")
			}

			editorCmd := os.Getenv("EDITOR")
			if editorCmd == "" {
				editorCmd = "vim"
			}

			log.Debug().Str("editor", editorCmd).Str("path", profilesPath).Msg("Opening editor")

			editCmd := exec.Command(editorCmd, profilesPath)
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			if err := editCmd.Run(); err != nil {
				return fmt.Errorf("editor command ('%s') failed: %w", editorCmd, err)
			}
			return nil
		},
	}
}

func newInitCommand(appName string, initialContentProvider InitialContentProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new profiles file with default content",
		RunE: func(cmd *cobra.Command, args []string) error {
			profilesPath, err := GetProfilesPathForApp(appName)
			if err != nil {
				return err
			}

			// Check if file already exists
			if _, err := os.Stat(profilesPath); err == nil {
				return fmt.Errorf("profiles file already exists at %s", profilesPath)
			} else if !os.IsNotExist(err) {
				// Handle other errors during stat (e.g., permission issues)
				return fmt.Errorf("error checking profiles file %s: %w", profilesPath, err)
			}

			// Ensure the directory exists
			profilesDir := filepath.Dir(profilesPath)
			if err := os.MkdirAll(profilesDir, 0755); err != nil {
				return fmt.Errorf("could not create profiles directory %s: %w", profilesDir, err)
			}

			// Get initial content and write the file
			content := initialContentProvider()
			if err := os.WriteFile(profilesPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("could not write profiles file %s: %w", profilesPath, err)
			}

			fmt.Printf("Initialized new profiles file at %s\n", profilesPath)
			return nil
		},
	}
}

func newDuplicateCommand(getEditor func() (*ProfilesEditor, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "duplicate <source-profile> <new-profile>",
		Short: "Duplicate an existing profile with a new name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := getEditor()
			if err != nil {
				// Cannot duplicate if editor failed to load
				return err
			}

			sourceProfile := args[0]
			newProfile := args[1]

			if err := editor.DuplicateProfile(sourceProfile, newProfile); err != nil {
				return err // Error includes context from DuplicateProfile
			}

			if err := editor.Save(); err != nil {
				return fmt.Errorf("failed to save profiles after duplication: %w", err)
			}

			fmt.Printf("Duplicated profile '%s' to '%s'\n", sourceProfile, newProfile)
			return nil
		},
	}
}
