package trie

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/stretchr/testify/assert"
)

// MockCommand implements cmds.Command for testing
type MockCommand struct {
	name string
}

func (m *MockCommand) Description() *cmds.CommandDescription {
	return &cmds.CommandDescription{
		Name: m.name,
	}
}

func (m *MockCommand) ToYAML(w io.Writer) error {
	return nil
}

func (m *MockCommand) ParseArguments(args []string) error {
	return nil
}

// Helper to create test commands
func createMockCommands(names ...string) []cmds.Command {
	commands := make([]cmds.Command, len(names))
	for i, name := range names {
		commands[i] = &MockCommand{name: name}
	}
	return commands
}

// Basic Node Operations Tests
func TestNewTrieNode(t *testing.T) {
	t.Run("empty node", func(t *testing.T) {
		node := NewTrieNode(nil, nil)
		assert.NotNil(t, node)
		assert.Empty(t, node.Children)
		assert.Empty(t, node.Commands)
	})

	t.Run("node with commands", func(t *testing.T) {
		commands := createMockCommands("cmd1", "cmd2")
		node := NewTrieNode(commands, nil)
		assert.NotNil(t, node)
		assert.Empty(t, node.Children)
		assert.Len(t, node.Commands, 2)
	})
}

func TestFindNode(t *testing.T) {
	root := NewTrieNode(nil, nil)

	// Create a path: root -> a -> b
	root.Children["a"] = NewTrieNode(nil, nil)
	root.Children["a"].Children["b"] = NewTrieNode(createMockCommands("cmdB"), nil)

	t.Run("find non-existent node", func(t *testing.T) {
		node := root.FindNode([]string{"nonexistent"})
		assert.Nil(t, node)
	})

	t.Run("find root", func(t *testing.T) {
		node := root.FindNode([]string{})
		assert.Equal(t, root, node)
	})

	t.Run("find existing node at depth", func(t *testing.T) {
		node := root.FindNode([]string{"a", "b"})
		assert.NotNil(t, node)
		assert.Len(t, node.Commands, 1)
		assert.Equal(t, "cmdB", node.Commands[0].Description().Name)
	})
}

// Command Operations Tests
func TestInsertCommand(t *testing.T) {
	t.Run("insert at root", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		cmd := &MockCommand{name: "root-cmd"}
		root.InsertCommand([]string{}, cmd)
		assert.Len(t, root.Commands, 1)
		assert.Equal(t, "root-cmd", root.Commands[0].Description().Name)
	})

	t.Run("insert at depth", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		cmd := &MockCommand{name: "deep-cmd"}
		root.InsertCommand([]string{"a", "b"}, cmd)

		node := root.FindNode([]string{"a", "b"})
		assert.NotNil(t, node)
		assert.Len(t, node.Commands, 1)
		assert.Equal(t, "deep-cmd", node.Commands[0].Description().Name)
	})

	t.Run("replace existing command", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		cmd1 := &MockCommand{name: "cmd"}
		cmd2 := &MockCommand{name: "cmd"} // Same name, different instance

		root.InsertCommand([]string{}, cmd1)
		root.InsertCommand([]string{}, cmd2)

		assert.Len(t, root.Commands, 1)
		assert.Equal(t, cmd2, root.Commands[0])
	})
}

func TestFindCommand(t *testing.T) {
	root := NewTrieNode(nil, nil)
	rootCmd := &MockCommand{name: "root-cmd"}
	deepCmd := &MockCommand{name: "deep-cmd"}

	root.InsertCommand([]string{}, rootCmd)
	root.InsertCommand([]string{"a", "b"}, deepCmd)

	t.Run("find root command", func(t *testing.T) {
		cmd, found := root.FindCommand([]string{"root-cmd"})
		assert.True(t, found)
		assert.Equal(t, rootCmd, cmd)
	})

	t.Run("find deep command", func(t *testing.T) {
		cmd, found := root.FindCommand([]string{"a", "b", "deep-cmd"})
		assert.True(t, found)
		assert.Equal(t, deepCmd, cmd)
	})

	t.Run("find non-existent command", func(t *testing.T) {
		cmd, found := root.FindCommand([]string{"not-exists"})
		assert.False(t, found)
		assert.Nil(t, cmd)
	})

	t.Run("find with empty path", func(t *testing.T) {
		cmd, found := root.FindCommand([]string{})
		assert.False(t, found)
		assert.Nil(t, cmd)
	})
}

func TestRemove(t *testing.T) {
	t.Run("remove from root", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("cmd1", "cmd2"), nil)
		removed := root.Remove([]string{"cmd1"})
		assert.Len(t, removed, 1)
		assert.Equal(t, "cmd1", removed[0].Description().Name)
		assert.Len(t, root.Commands, 1)
	})

	t.Run("remove from depth", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		cmd := &MockCommand{name: "deep-cmd"}
		root.InsertCommand([]string{"a", "b"}, cmd)

		removed := root.Remove([]string{"a", "b", "deep-cmd"})
		assert.Len(t, removed, 1)
		assert.Equal(t, "deep-cmd", removed[0].Description().Name)

		node := root.FindNode([]string{"a", "b"})
		assert.NotNil(t, node)
		assert.Empty(t, node.Commands)
	})

	t.Run("remove entire subtree", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		root.InsertCommand([]string{"a"}, &MockCommand{name: "cmd1"})
		root.InsertCommand([]string{"a", "b"}, &MockCommand{name: "cmd2"})

		removed := root.Remove([]string{"a"})
		assert.Len(t, removed, 2)
		assert.Nil(t, root.FindNode([]string{"a"}))
	})
}

// Command Collection Tests
func TestCollectCommands(t *testing.T) {
	t.Run("collect from empty trie", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		commands := root.CollectCommands([]string{}, true)
		assert.Empty(t, commands)
	})

	t.Run("collect from root with recursion", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("root1", "root2"), nil)
		root.Children["a"] = NewTrieNode(createMockCommands("a1", "a2"), nil)
		root.Children["b"] = NewTrieNode(createMockCommands("b1"), nil)

		commands := root.CollectCommands([]string{}, true)
		assert.Len(t, commands, 5)
		names := make([]string, len(commands))
		for i, cmd := range commands {
			names[i] = cmd.Description().Name
		}
		assert.ElementsMatch(t, []string{"root1", "root2", "a1", "a2", "b1"}, names)
	})

	t.Run("collect from root without recursion", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("root1", "root2"), nil)
		root.Children["a"] = NewTrieNode(createMockCommands("a1", "a2"), nil)

		commands := root.CollectCommands([]string{}, false)
		assert.Len(t, commands, 2)
		names := make([]string, len(commands))
		for i, cmd := range commands {
			names[i] = cmd.Description().Name
		}
		assert.ElementsMatch(t, []string{"root1", "root2"}, names)
	})

	t.Run("collect from specific prefix", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		root.Children["a"] = NewTrieNode(createMockCommands("a1", "a2"), nil)
		root.Children["a"].Children["b"] = NewTrieNode(createMockCommands("b1"), nil)

		commands := root.CollectCommands([]string{"a"}, true)
		assert.Len(t, commands, 3)
		names := make([]string, len(commands))
		for i, cmd := range commands {
			names[i] = cmd.Description().Name
		}
		assert.ElementsMatch(t, []string{"a1", "a2", "b1"}, names)
	})

	t.Run("collect with non-existent prefix", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("root1"), nil)
		commands := root.CollectCommands([]string{"nonexistent"}, true)
		assert.Empty(t, commands)
	})
}

// Node Insertion and Rendering Tests
func TestInsertNode(t *testing.T) {
	t.Run("insert at root", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		nodeToInsert := NewTrieNode(createMockCommands("cmd1", "cmd2"), nil)

		root.InsertNode([]string{}, nodeToInsert)
		assert.Len(t, root.Commands, 2)
		assert.Equal(t, "cmd1", root.Commands[0].Description().Name)
		assert.Equal(t, "cmd2", root.Commands[1].Description().Name)
	})

	t.Run("insert at depth", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		nodeToInsert := NewTrieNode(createMockCommands("deep1"), nil)

		root.InsertNode([]string{"a", "b"}, nodeToInsert)
		node := root.FindNode([]string{"a", "b"})
		assert.NotNil(t, node)
		assert.Len(t, node.Commands, 1)
		assert.Equal(t, "deep1", node.Commands[0].Description().Name)
	})

	t.Run("insert with overlapping commands", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("existing"), nil)
		nodeToInsert := NewTrieNode(createMockCommands("new"), nil)

		root.InsertNode([]string{}, nodeToInsert)
		assert.Len(t, root.Commands, 2)
		names := []string{root.Commands[0].Description().Name, root.Commands[1].Description().Name}
		assert.ElementsMatch(t, []string{"existing", "new"}, names)
	})

	t.Run("insert with overlapping children", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		root.Children["a"] = NewTrieNode(createMockCommands("existing"), nil)

		nodeToInsert := NewTrieNode(nil, nil)
		nodeToInsert.Children["b"] = NewTrieNode(createMockCommands("new"), nil)

		root.InsertNode([]string{"a"}, nodeToInsert)
		node := root.FindNode([]string{"a"})
		assert.NotNil(t, node)
		assert.Len(t, node.Commands, 1)
		assert.Len(t, node.Children, 1)
		assert.Equal(t, "existing", node.Commands[0].Description().Name)

		childNode := node.Children["b"]
		assert.NotNil(t, childNode)
		assert.Len(t, childNode.Commands, 1)
		assert.Equal(t, "new", childNode.Commands[0].Description().Name)
	})
}

func TestToRenderNode(t *testing.T) {
	t.Run("convert empty trie", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		renderNode := root.ToRenderNode()
		assert.Empty(t, renderNode.Name)
		assert.Nil(t, renderNode.Command)
		assert.Empty(t, renderNode.Children)
	})

	t.Run("convert single level", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("cmd1", "cmd2"), nil)
		renderNode := root.ToRenderNode()

		assert.Empty(t, renderNode.Name)
		assert.Nil(t, renderNode.Command)
		assert.Len(t, renderNode.Children, 2)

		names := []string{renderNode.Children[0].Name, renderNode.Children[1].Name}
		assert.ElementsMatch(t, []string{"cmd1", "cmd2"}, names)
	})

	t.Run("convert multiple levels", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		root.Children["a"] = NewTrieNode(createMockCommands("a1"), nil)
		root.Children["a"].Children["b"] = NewTrieNode(createMockCommands("b1"), nil)

		renderNode := root.ToRenderNode()
		assert.Len(t, renderNode.Children, 1)
		assert.Equal(t, "a", renderNode.Children[0].Name)

		aNode := renderNode.Children[0]
		assert.Len(t, aNode.Children, 2) // "a1" command and "b" child

		var bNode *RenderNode
		for _, child := range aNode.Children {
			if child.Name == "b" {
				bNode = child
				break
			}
		}

		assert.NotNil(t, bNode)
		assert.Len(t, bNode.Children, 1)
		assert.Equal(t, "b1", bNode.Children[0].Name)
	})

	t.Run("verify sorting of children", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("c", "a", "b"), nil)
		renderNode := root.ToRenderNode()

		assert.Len(t, renderNode.Children, 3)
		names := []string{
			renderNode.Children[0].Name,
			renderNode.Children[1].Name,
			renderNode.Children[2].Name,
		}
		assert.Equal(t, []string{"a", "b", "c"}, names)
	})
}

// Edge Cases and Helper Functions Tests
func TestEdgeCases(t *testing.T) {
	t.Run("deep nested structure", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		path := []string{"a", "b", "c", "d", "e"}
		cmd := &MockCommand{name: "deep-cmd"}

		// Insert deep command
		root.InsertCommand(path, cmd)

		// Verify we can find it
		foundCmd, found := root.FindCommand(append(path, "deep-cmd"))
		assert.True(t, found)
		assert.Equal(t, cmd, foundCmd)

		// Verify collection works at different levels
		assert.Len(t, root.CollectCommands([]string{}, true), 1)
		assert.Len(t, root.CollectCommands([]string{"a", "b"}, true), 1)
		assert.Empty(t, root.CollectCommands([]string{"a", "b"}, false))

		// Remove from middle of path
		removed := root.Remove([]string{"a", "b"})
		assert.Len(t, removed, 1)
		assert.Equal(t, "deep-cmd", removed[0].Description().Name)
	})

	t.Run("large number of commands", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		numCommands := 1000
		commands := make([]cmds.Command, numCommands)

		// Create and insert many commands
		for i := 0; i < numCommands; i++ {
			cmd := &MockCommand{name: fmt.Sprintf("cmd-%d", i)}
			commands[i] = cmd
			root.InsertCommand([]string{}, cmd)
		}

		// Verify all commands are present
		collected := root.CollectCommands([]string{}, false)
		assert.Len(t, collected, numCommands)

		// Verify we can find specific commands
		cmd, found := root.FindCommand([]string{"cmd-42"})
		assert.True(t, found)
		assert.Equal(t, "cmd-42", cmd.Description().Name)

		// Remove half the commands
		for i := 0; i < numCommands/2; i++ {
			name := fmt.Sprintf("cmd-%d", i)
			removed := root.Remove([]string{name})
			assert.Len(t, removed, 1)
			assert.Equal(t, name, removed[0].Description().Name)
		}

		// Verify remaining commands
		remaining := root.CollectCommands([]string{}, false)
		assert.Len(t, remaining, numCommands/2)
	})

	t.Run("empty prefix handling", func(t *testing.T) {
		root := NewTrieNode(createMockCommands("root-cmd"), nil)

		// Empty prefix for various operations
		assert.Equal(t, root, root.FindNode([]string{}))
		assert.Len(t, root.CollectCommands([]string{}, false), 1)

		// Empty prefix in path
		root.InsertCommand([]string{"", "a", "", "b"}, &MockCommand{name: "test"})
		cmd, found := root.FindCommand([]string{"", "a", "", "b", "test"})
		assert.True(t, found)
		assert.Equal(t, "test", cmd.Description().Name)
	})

	t.Run("concurrent operations", func(t *testing.T) {
		root := NewTrieNode(nil, nil)
		numGoroutines := 10
		numOperations := 100
		var wg sync.WaitGroup

		// Concurrent insertions
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					cmd := &MockCommand{name: fmt.Sprintf("cmd-%d-%d", id, j)}
					root.InsertCommand([]string{fmt.Sprintf("group-%d", id)}, cmd)
				}
			}(i)
		}
		wg.Wait()

		// Verify all commands were inserted
		allCommands := root.CollectCommands([]string{}, true)
		assert.Len(t, allCommands, numGoroutines*numOperations)

		// Concurrent reads and removals
		var removedCount int32
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					// Mix of reads and removals
					if j%2 == 0 {
						removed := root.Remove([]string{fmt.Sprintf("group-%d", id),
							fmt.Sprintf("cmd-%d-%d", id, j)})
						if len(removed) > 0 {
							atomic.AddInt32(&removedCount, 1)
						}
					} else {
						root.FindCommand([]string{fmt.Sprintf("group-%d", id),
							fmt.Sprintf("cmd-%d-%d", id, j)})
					}
				}
			}(i)
		}
		wg.Wait()

		// Verify expected number of commands were removed
		remaining := root.CollectCommands([]string{}, true)
		assert.Len(t, remaining, numGoroutines*numOperations-int(removedCount))
	})
}
