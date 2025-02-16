package repositories

import (
	"context"

	"github.com/go-go-golems/clay/pkg/repositories/mcp"
	"github.com/go-go-golems/clay/pkg/repositories/trie"
	"github.com/go-go-golems/clay/pkg/watcher"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
)

// CommandRepository is a simple repository that just manages commands in memory.
// It doesn't deal with files or watching, just provides a way to add and organize commands.
type CommandRepository struct {
	root *trie.TrieNode
	name string
}

type CommandRepositoryOption func(*CommandRepository)

func WithCommandRepositoryName(name string) CommandRepositoryOption {
	return func(r *CommandRepository) {
		r.name = name
	}
}

// NewCommandRepository creates a new command repository that just manages commands in memory
func NewCommandRepository(options ...CommandRepositoryOption) *CommandRepository {
	ret := &CommandRepository{
		root: trie.NewTrieNode([]cmds.Command{}, nil),
	}

	for _, opt := range options {
		opt(ret)
	}

	return ret
}

// LoadCommands is a no-op for CommandRepository since it doesn't load from files
func (r *CommandRepository) LoadCommands(_ *help.HelpSystem, _ ...cmds.CommandDescriptionOption) error {
	return nil
}

// Add adds one or more commands to the repository, optionally under a specific path
func (r *CommandRepository) Add(commands ...cmds.Command) {
	for _, command := range commands {
		prefix := command.Description().Parents
		r.root.InsertCommand(prefix, command)
	}
}

// AddUnderPath adds commands under a specific path prefix
func (r *CommandRepository) AddUnderPath(pathPrefix []string, commands ...cmds.Command) {
	for _, command := range commands {
		// Create a new slice to avoid modifying the original command's parents
		newPrefix := append([]string{}, pathPrefix...)
		newPrefix = append(newPrefix, command.Description().Parents...)
		r.root.InsertCommand(newPrefix, command)
	}
}

// Remove removes commands with the given prefixes from the repository
func (r *CommandRepository) Remove(prefixes ...[]string) {
	for _, prefix := range prefixes {
		r.root.Remove(prefix)
	}
}

// CollectCommands returns all commands under a given prefix
func (r *CommandRepository) CollectCommands(prefix []string, recurse bool) []cmds.Command {
	return r.root.CollectCommands(prefix, recurse)
}

// GetCommand returns a single command by its full path name
func (r *CommandRepository) GetCommand(name string) (cmds.Command, bool) {
	if name == "" {
		return nil, false
	}

	prefix := []string{name}
	commands := r.CollectCommands(prefix, false)
	if len(commands) == 0 {
		return nil, false
	}

	return commands[0], true
}

// FindNode returns the TrieNode at the given prefix
func (r *CommandRepository) FindNode(prefix []string) *trie.TrieNode {
	return r.root.FindNode(prefix)
}

// GetRenderNode returns a RenderNode for visualization purposes
func (r *CommandRepository) GetRenderNode(prefix []string) (*trie.RenderNode, bool) {
	node := r.root.FindNode(prefix)
	if node == nil {
		return nil, false
	}

	ret := node.ToRenderNode()
	if len(prefix) > 0 {
		ret.Name = prefix[len(prefix)-1]
	}
	cmd, ok := r.root.FindCommand(prefix)
	if ok {
		ret.Command = cmd
		ret.Name = cmd.Description().Name
	}

	return ret, true
}

// ListTools returns all commands as tools for MCP compatibility
func (r *CommandRepository) ListTools(ctx context.Context, cursor string) ([]mcp.Tool, string, error) {
	commands := r.root.CollectCommands([]string{}, true)
	tools := make([]mcp.Tool, 0, len(commands))

	for _, cmd := range commands {
		desc := cmd.Description()
		tools = append(tools, mcp.Tool{
			Name:        desc.FullPath(),
			Description: desc.Short,
		})
	}

	return tools, "", nil
}

// Watch is a no-op since CommandRepository doesn't support file watching
func (r *CommandRepository) Watch(ctx context.Context, options ...watcher.Option) error {
	return nil
}
