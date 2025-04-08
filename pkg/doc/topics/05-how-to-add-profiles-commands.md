---
Title: How to Add Reusable Profile Management Commands
Slug: how-to-add-profiles-commands
Short: Learn how to integrate the reusable profile management commands provided by the `clay/pkg/cmds/profiles` package into your Cobra-based command-line applications
Topics:
  - profiles
  - yaml
  - cobra
  - command-line-applications
Commands:
  - NewProfilesCommand
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

This document explains how to integrate the reusable profile management commands provided by the `clay/pkg/cmds/profiles` package into your Cobra-based command-line applications.

## Overview

Many applications benefit from using configuration profiles to manage different settings for various environments or use cases (e.g., development vs. production databases, different API keys, specific model parameters). The `clay` library provides a standardized set of commands (`profiles list`, `get`, `set`, `delete`, `init`, `edit`, `duplicate`) built on top of `clay/pkg/yaml-editor` to manage these profiles stored in a YAML file.

By using this package, you get consistent profile management across different tools with minimal boilerplate code.

## Profile Storage

The profiles are stored in a standard location within the user's configuration directory:

```
~/.config/<appName>/profiles.yaml
```

Where `<appName>` is the name you provide when initializing the commands.

## Profile Structure

The `profiles.yaml` file is expected to have a simple, hierarchical structure:

```yaml
<profileName1>:
  <layerName1>:
    <settingKey1>: <settingValue1>
    <settingKey2>: <settingValue2>
  <layerName2>:
    <settingKey3>: <settingValue3>

<profileName2>:
  <layerName1>: # Layers can be reused across profiles
    <settingKey1>: <overrideValue1>
    # settingKey2 uses default if not overridden
  <layerName3>:
    <settingKey4>: <settingValue4>
```

- **Profiles:** Top-level keys represent profile names (e.g., `production-db`, `anyscale-mixtral`).
- **Layers:** Each profile contains one or more layers (e.g., `sql-connection`, `openai-chat`). Layers group related settings. Your application code will typically look for settings within specific layer names.
- **Settings:** Each layer contains key-value pairs for specific configuration parameters.

## Integrating the Commands

Integrating the profile commands into your application involves two main steps:

1.  **Define Initial Content:** Create a function that returns the default content (as a string) for a new `profiles.yaml` file when the `profiles init` command is run. This content should ideally include comments explaining the profile structure and some examples relevant to your application.
2.  **Create and Add the Command Group:** In your Cobra command setup (usually in `main.go` or a `cmd/root.go`), call the `profiles.NewProfilesCommand` factory function and add the resulting command group to your root command.

### Step 1: Define Initial Content Provider

Create a function that returns a string. This string will be written to `profiles.yaml` when a user runs `your-app profiles init` for the first time.

**Example (`pinocchio/cmd/pinocchio/main.go`):**

```go
// pinocchioInitialProfilesContent provides the default YAML content for a new pinocchio profiles file.
func pinocchioInitialProfilesContent() string {
	return `# Pinocchio Profiles Configuration
#
# This file contains profile configurations for Pinocchio.
# Each profile can override layer parameters for different components (like AI models).
# ... (rest of the example content) ...
#
# You can manage this file using the 'pinocchio profiles' commands:
# - list: List all profiles
# - get <profile> [layer] [key]: Get profile settings
# - set <profile> <layer> <key> <value>: Set a profile setting
# - delete <profile> [layer] [key]: Delete a profile, layer, or setting
# - edit: Open this file in your editor
# - init: Create this file if it doesn't exist
# - duplicate <source> <new>: Copy an existing profile
`
}
```

### Step 2: Add the Command Group

Import the `clay/pkg/cmds/profiles` package (e.g., `clay_profiles "github.com/go-go-golems/clay/pkg/cmds/profiles"`).

Call `clay_profiles.NewProfilesCommand`, passing your application's name and the initial content provider function you just defined.

Add the returned command to your Cobra root command.

**Example (`pinocchio/cmd/pinocchio/main.go` within `initAllCommands` or similar):**

```go
import (
	// ... other imports
	clay_profiles "github.com/go-go-golems/clay/pkg/cmds/profiles"
)

func initAllCommands(helpSystem *help.HelpSystem) error {
	// ... other command initializations ...

	// Add profiles command from clay
	// "pinocchio" is the appName, used for ~/.config/pinocchio/profiles.yaml
	profilesCmd, err := clay_profiles.NewProfilesCommand("pinocchio", pinocchioInitialProfilesContent)
	if err != nil {
		return fmt.Errorf("error initializing profiles command: %w", err)
	}
	rootCmd.AddCommand(profilesCmd)

	// ... add other commands ...

	return nil
}
```

## Available Commands

Once integrated, your application will have the following subcommands under `profiles`:

- `your-app profiles list [-c | --concise]`: Lists all defined profile names. With `-c`, only names are shown; otherwise, full content is displayed.
- `your-app profiles get <profile> [layer] [key]`: Retrieves and displays settings. Shows all layers for a profile, all settings for a layer, or a specific setting's value.
- `your-app profiles set <profile> <layer> <key> <value>`: Sets a specific setting's value. Creates the profile and/or layer if they don't exist.
- `your-app profiles delete <profile> [layer] [key]`: Deletes an entire profile, a specific layer within a profile, or a single setting within a layer.
- `your-app profiles init`: Creates the `profiles.yaml` file in `~/.config/<appName>/` with the default content, but only if it doesn't already exist.
- `your-app profiles edit`: Opens the `profiles.yaml` file in the default system editor (`$EDITOR`, fallback to `vim`). Creates the directory and an empty file if they don't exist.
- `your-app profiles duplicate <source-profile> <new-profile>`: Creates a new profile by copying all layers and settings from an existing source profile.

## Using Profiles in Your Application

While this package provides the commands to _manage_ the profiles file, your application logic needs to _read_ and _apply_ these profiles. This typically involves:

1.  Adding a `--profile <name>` flag (often using `glazed/pkg/cmds/layers` or similar parameter layer mechanisms).
2.  Loading the `profiles.yaml` file corresponding to your application name.
3.  Reading the settings from the specified profile and the relevant layers.
4.  Merging these profile settings with default values and other configuration sources (like environment variables or command-line flags).

The `pinocchio`, `sqleton`, and `escuse-me` applications provide examples of how profiles are loaded and applied to override default layer parameters.
