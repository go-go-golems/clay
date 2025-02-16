package repositories

import (
	"context"

	"github.com/go-go-golems/clay/pkg/repositories/mcp"
	"github.com/go-go-golems/clay/pkg/repositories/trie"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
)

// RepositoryInterface defines the core functionality that all repositories must implement
type RepositoryInterface interface {
	// LoadCommands initializes the repository by loading all commands
	LoadCommands(helpSystem *help.HelpSystem, options ...cmds.CommandDescriptionOption) error

	// Add adds one or more commands to the repository
	Add(commands ...cmds.Command)

	// Remove removes commands with the given prefixes from the repository
	Remove(prefixes ...[]string)

	// CollectCommands returns all commands under a given prefix
	CollectCommands(prefix []string, recurse bool) []cmds.Command

	// GetCommand returns a single command by its full path name
	GetCommand(name string) (cmds.Command, bool)

	// FindNode returns the TrieNode at the given prefix
	FindNode(prefix []string) *trie.TrieNode

	// GetRenderNode returns a RenderNode for visualization purposes
	GetRenderNode(prefix []string) (*trie.RenderNode, bool)

	// ListTools returns all commands as tools for MCP compatibility
	ListTools(ctx context.Context, cursor string) ([]mcp.Tool, string, error)
}
