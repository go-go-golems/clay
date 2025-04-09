---
Title: Managing Repositories 
Slug: managing-repositories
Short: Learn how to add, remove, and list repositories in Clay's configuration
Topics:
- configuration
- repositories
Commands:
- repositories add
- repositories remove
- repositories get
Flags:
- none
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Managing Repositories 

This program provides a set of commands to manage repositories in your configuration file. These commands allow you to add new repositories, remove existing ones, and list all currently configured repositories.

## Adding Repositories

To add one or more repositories to your configuration:

```
<program> repositories add [directories...]
```

This command will:
- Convert relative paths to absolute paths
- Skip directories that already exist in the configuration
- Add new directories to the repository list
- Display the updated list of repositories

## Removing Repositories

To remove one or more repositories from your configuration:

```
<program> repositories remove [directories...]
```

This command will:
- Convert relative paths to absolute paths
- Remove specified directories from the repository list
- Display the updated list of repositories

## Listing Repositories

To view all currently configured repositories:

```
<program> repositories get
```

This command will display a list of all repositories currently in your configuration file.

## Configuration File

These commands interact with the program's configuration file, which is typically located in your home directory. The repositories are stored under the `repositories` key in YAML format.

Note: If no configuration file is found, these commands will return an error.

By using these commands, you can easily manage the set of repositories that Clay works with, allowing you to organize and access your clay files efficiently.