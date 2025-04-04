package multi_repository

import (
	"context"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/clay/pkg/repositories/mcp"
	"github.com/go-go-golems/clay/pkg/repositories/trie"
	"github.com/go-go-golems/clay/pkg/watcher"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
)

type MockRepository struct {
	commands     []cmds.Command
	loadError    error
	helpSystem   *help.HelpSystem
	loadOptions  []cmds.CommandDescriptionOption
	addCalls     [][]cmds.Command
	removeCalls  [][]string
	findNodeRet  *trie.TrieNode
	renderNode   *trie.RenderNode
	renderNodeOk bool
	tools        []mcp.Tool
	toolsError   error
}

func NewMockRepository(commands []cmds.Command) *MockRepository {
	return &MockRepository{
		commands:    commands,
		addCalls:    make([][]cmds.Command, 0),
		removeCalls: make([][]string, 0),
	}
}

var _ repositories.RepositoryInterface = (*MockRepository)(nil)

func (m *MockRepository) LoadCommands(helpSystem *help.HelpSystem, options ...cmds.CommandDescriptionOption) error {
	m.helpSystem = helpSystem
	m.loadOptions = options
	return m.loadError
}

func (m *MockRepository) Add(commands ...cmds.Command) {
	m.addCalls = append(m.addCalls, commands)
	m.commands = append(m.commands, commands...)
}

func (m *MockRepository) Remove(prefixes ...[]string) {
	m.removeCalls = append(m.removeCalls, prefixes...)
}

func (m *MockRepository) CollectCommands(prefix []string, recurse bool) []cmds.Command {
	return m.commands
}

func (m *MockRepository) GetCommand(name string) (cmds.Command, bool) {
	for _, cmd := range m.commands {
		if cmd.Description().FullPath() == name {
			return cmd, true
		}
	}
	return nil, false
}

func (m *MockRepository) FindNode(prefix []string) *trie.TrieNode {
	return m.findNodeRet
}

func (m *MockRepository) GetRenderNode(prefix []string) (*trie.RenderNode, bool) {
	return m.renderNode, m.renderNodeOk
}

func (m *MockRepository) ListTools(ctx context.Context, cursor string) ([]mcp.Tool, string, error) {
	return m.tools, "", m.toolsError
}

func (m *MockRepository) Watch(ctx context.Context, options ...watcher.Option) error {
	return nil
}
