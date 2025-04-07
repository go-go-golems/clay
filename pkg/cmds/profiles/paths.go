package profiles

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetProfilesPathForApp returns the default path for the profiles YAML file
// for a given application name. It typically uses ~/.config/<appName>/profiles.yaml.
func GetProfilesPathForApp(appName string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory if UserConfigDir fails
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user config or home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
		// We don't create the .config dir here, let the caller handle directory creation.
	}

	// Use ~/.config/<appName>/profiles.yaml
	appConfigDir := filepath.Join(configDir, appName)
	return filepath.Join(appConfigDir, "profiles.yaml"), nil
}
