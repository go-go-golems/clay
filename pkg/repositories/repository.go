package repositories

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// A repository is a collection of commands and aliases, that can optionally be reloaded
// through a watcher (and for which you can register callbacks, for example to update a potential
// cobra command or REST route).

type UpdateCallback func(cmd cmds.Command) error
type RemoveCallback func(cmd cmds.Command) error

type Directory struct {
	FS fs.FS
	// Root directories are relative to the FS
	RootDirectory    string
	RootDocDirectory string
	Name             string
	SourcePrefix     string
	WatchDirectory   string
}

type Repository struct {
	Name        string
	Directories []Directory
	Files       []string // New field for individual files
	// The root of the repository.
	Root           *TrieNode
	updateCallback UpdateCallback
	removeCallback RemoveCallback

	// loader is used to load all commands on startup
	loader loaders.CommandLoader
}

type RepositoryOption func(*Repository)

func WithName(name string) RepositoryOption {
	return func(r *Repository) {
		r.Name = name
	}
}

func WithDirectories(directories ...Directory) RepositoryOption {
	return func(r *Repository) {
		r.Directories = directories
	}
}

// WithCommandLoader sets the command loader to use when loading commands from
// the filesystem on startup or when a directory changes.
func WithCommandLoader(loader loaders.CommandLoader) RepositoryOption {
	return func(r *Repository) {
		r.loader = loader
	}
}

func WithUpdateCallback(callback UpdateCallback) RepositoryOption {
	return func(r *Repository) {
		r.updateCallback = callback
	}
}

func WithRemoveCallback(callback RemoveCallback) RepositoryOption {
	return func(r *Repository) {
		r.removeCallback = callback
	}
}

func WithFiles(files ...string) RepositoryOption {
	return func(r *Repository) {
		r.Files = files
	}
}

// NewRepository creates a new repository.
func NewRepository(options ...RepositoryOption) *Repository {
	ret := &Repository{
		Root: NewTrieNode([]cmds.Command{}, []*alias.CommandAlias{}),
	}
	for _, opt := range options {
		opt(ret)
	}
	return ret
}

// LoadCommands initializes the repository by loading all commands from the loader,
// if available.
func (r *Repository) LoadCommands(helpSystem *help.HelpSystem, options ...cmds.CommandDescriptionOption) error {
	if r.loader != nil {
		commands := make([]cmds.Command, 0)
		aliases := make([]*alias.CommandAlias, 0)

		// Load from directories
		for _, directory := range r.Directories {
			source := ""
			if directory.SourcePrefix != "" {
				source = directory.SourcePrefix + ":"
			}
			if r.Name != "" {
				source = source + r.Name + ":"
			}
			if directory.Name != "" {
				source = source + directory.Name
			}
			base := filepath.Base(directory.RootDirectory)
			if base != "." {
				source = source + "/" + base
			}

			options_ := append([]cmds.CommandDescriptionOption{
				cmds.WithStripParentsPrefix([]string{directory.RootDirectory}),
			}, options...)

			aliasOptions := []alias.Option{
				alias.WithStripParentsPrefix([]string{directory.RootDirectory}),
			}

			commands_, err := loaders.LoadCommandsFromFS(
				directory.FS,
				directory.RootDirectory,
				source,
				r.loader,
				options_, aliasOptions)
			if err != nil {
				return err
			}
			for _, command := range commands_ {
				switch v := command.(type) {
				case *alias.CommandAlias:
					aliases = append(aliases, v)
				case cmds.Command:
					commands = append(commands, v)
				default:
					return errors.New(fmt.Sprintf("unknown command type %T", v))
				}
			}

			// Check if the RootDocDirectory exists
			file, err := directory.FS.Open(directory.RootDocDirectory)
			if err != nil {
				if os.IsNotExist(err) {
					// Directory doesn't exist, skip loading
					continue
				}
				// Return other errors
				return err
			}
			_ = file.Close()

			// If directory exists, proceed with loading sections
			err = helpSystem.LoadSectionsFromFS(directory.FS, directory.RootDocDirectory)
			if err != nil {
				return err
			}
		}

		// Load from individual files
		for _, file := range r.Files {
			fs, filePath, err := loaders.FileNameToFsFilePath(file)
			if err != nil {
				return errors.Wrapf(err, "could not get fs and file path for %s", file)
			}

			source := ""
			if r.Name != "" {
				source = r.Name + ":"
			}
			source = source + "file:" + file

			commands_, err := r.loader.LoadCommands(
				fs,
				filePath,
				append([]cmds.CommandDescriptionOption{
					cmds.WithSource(source),
				}, options...),
				[]alias.Option{},
			)
			if err != nil {
				return errors.Wrapf(err, "could not load commands from file %s", file)
			}

			for _, command := range commands_ {
				switch v := command.(type) {
				case *alias.CommandAlias:
					aliases = append(aliases, v)
				case cmds.Command:
					commands = append(commands, v)
				default:
					return errors.New(fmt.Sprintf("unknown command type %T", v))
				}
			}
		}

		r.Add(commands...)
		for _, alias_ := range aliases {
			r.Add(alias_)
		}
	}

	return nil
}

func (r *Repository) Add(commands ...cmds.Command) {
	aliases := []*alias.CommandAlias{}

	for _, command := range commands {
		_, isAlias := command.(*alias.CommandAlias)
		if isAlias {
			aliases = append(aliases, command.(*alias.CommandAlias))
			continue
		}

		prefix := command.Description().Parents
		r.Root.InsertCommand(prefix, command)
		if r.updateCallback != nil {
			err := r.updateCallback(command)
			if err != nil {
				log.Warn().Err(err).Msg("error while updating command")
			}
		}
	}

	for _, alias_ := range aliases {
		prefix := alias_.Parents
		aliasedCommand, ok := r.Root.FindCommand(prefix)
		if !ok {
			name := alias_.Name
			log.Warn().Msgf("alias %s (prefix: %v, source %s) for %s not found", name, prefix, alias_.Source, alias_.AliasFor)
			continue
		}
		alias_.AliasedCommand = aliasedCommand

		r.Root.InsertCommand(prefix, alias_)
		if r.updateCallback != nil {
			err := r.updateCallback(alias_)
			if err != nil {
				log.Warn().Err(err).Msg("error while updating command")
			}
		}
	}
}

func (r *Repository) Remove(prefixes ...[]string) {
	for _, prefix := range prefixes {
		removedCommands := r.Root.Remove(prefix)
		for _, command := range removedCommands {
			if r.removeCallback != nil {
				err := r.removeCallback(command)
				if err != nil {
					log.Warn().Err(err).Msg("error while removing command")
				}
			}
		}
	}
}

func (r *Repository) CollectCommands(prefix []string, recurse bool) []cmds.Command {
	return r.Root.CollectCommands(prefix, recurse)
}

// GetCommand returns a single command by its full path name (components separated by /).
// It returns the command and true if found, nil and false otherwise.
func (r *Repository) GetCommand(name string) (cmds.Command, bool) {
	if name == "" {
		return nil, false
	}

	prefix := strings.Split(name, "/")
	commands := r.CollectCommands(prefix, false)
	if len(commands) == 0 {
		return nil, false
	}

	return commands[0], true
}

func (r *Repository) FindNode(prefix []string) *TrieNode {
	return r.Root.FindNode(prefix)
}

func (r *Repository) GetRenderNode(prefix []string) (*RenderNode, bool) {
	node := r.Root.FindNode(prefix)
	if node == nil {
		return nil, false
	}

	ret := node.ToRenderNode()
	if len(prefix) > 0 {
		ret.Name = prefix[len(prefix)-1]
	}
	cmd, ok := r.Root.FindCommand(prefix)
	if ok {
		ret.Command = cmd
		ret.Name = cmd.Description().Name
	}

	return ret, true
}
