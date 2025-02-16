package trie

import (
	"sort"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/rs/zerolog/log"
)

type TrieNode struct {
	Children map[string]*TrieNode
	Commands []cmds.Command
}

type RenderNode struct {
	Name     string
	Command  cmds.Command
	Children []*RenderNode
}

// NewTrieNode creates a new trie node.
func NewTrieNode(commands []cmds.Command, aliases []*alias.CommandAlias) *TrieNode {
	return &TrieNode{
		Children: make(map[string]*TrieNode),
		Commands: commands,
	}
}

// Remove removes a command from the trie.
func (t *TrieNode) Remove(prefix []string) []cmds.Command {
	if len(prefix) == 0 {
		commands := t.CollectCommands(prefix, true)
		t.Commands = make([]cmds.Command, 0)
		t.Children = make(map[string]*TrieNode)

		return commands
	}

	removedCommands := make([]cmds.Command, 0)

	// try to get parent node
	path := prefix[:len(prefix)-1]
	parentNode := t.findNode(path, false)
	name := prefix[len(prefix)-1]
	if parentNode == nil {
		log.Debug().Msgf("parent node not found for %s", name)
		return []cmds.Command{}
	}

	childNode, ok := parentNode.Children[name]
	if ok {

		// remove the node
		commands := childNode.CollectCommands([]string{}, true)
		removedCommands = append(removedCommands, commands...)

		delete(parentNode.Children, name)
	}
	// check if this is an actual command or alias
	for i, c := range parentNode.Commands {
		if c.Description().Name == name {
			removedCommands = append(removedCommands, c)
			parentNode.Commands = append(parentNode.Commands[:i], parentNode.Commands[i+1:]...)
		}
	}

	return removedCommands
}

// InsertCommand inserts a command in the trie, replacing it if it already exists.
func (t *TrieNode) InsertCommand(prefix []string, command cmds.Command) {
	node := t.findNode(prefix, true)

	// check if the command is already in the trie
	for i, c := range node.Commands {
		if c.Description().Name == command.Description().Name {
			node.Commands[i] = command
			return
		}
	}

	node.Commands = append(node.Commands, command)
}

// findNode finds the node corresponding to the given prefix, creating it if it doesn't exist.
func (t *TrieNode) findNode(prefix []string, createNewNodes bool) *TrieNode {
	node := t
	for _, p := range prefix {
		if _, ok := node.Children[p]; !ok {
			if !createNewNodes {
				log.Debug().Msgf("node %s not found", p)
				return nil
			}
			node.Children[p] = NewTrieNode([]cmds.Command{}, []*alias.CommandAlias{})
		}
		node = node.Children[p]
	}
	return node
}

func (t *TrieNode) FindNode(prefix []string) *TrieNode {
	return t.findNode(prefix, false)
}

func (t *TrieNode) FindCommand(path []string) (cmds.Command, bool) {
	if len(path) == 0 {
		return nil, false
	}
	parentPath := path[:len(path)-1]
	commandName := path[len(path)-1]
	node := t.findNode(parentPath, false)
	if node == nil {
		return nil, false
	}

	for _, c := range node.Commands {
		if c.Description().Name == commandName {
			return c, true
		}
	}

	return nil, false
}

// CollectCommands collects all commands and aliases under the given prefix.
func (t *TrieNode) CollectCommands(prefix []string, recurse bool) []cmds.Command {
	ret := make([]cmds.Command, 0)

	// Check if the prefix identifies a single command.
	if len(prefix) > 0 {
		// try to get parent node
		path := prefix[:len(prefix)-1]
		parentNode := t.findNode(path, false)
		name := prefix[len(prefix)-1]
		if parentNode != nil {
			for _, c := range parentNode.Commands {
				if c.Description().Name == name {
					ret = append(ret, c)
					break
				}
			}
		}

		if !recurse {
			return ret
		}
	}

	node := t.findNode(prefix, false)
	if node == nil {
		return ret
	}

	if !recurse {
		return node.Commands
	}

	// recurse into node to collect all commands and aliases
	for _, child := range node.Children {
		c := child.CollectCommands([]string{}, true)
		ret = append(ret, c...)
	}

	// add commands and aliases from current node
	ret = append(ret, node.Commands...)

	return ret
}

func (r *TrieNode) ToRenderNode() *RenderNode {
	ret := &RenderNode{
		Name:     "",
		Command:  nil,
		Children: nil,
	}
	childrenMap := make(map[string]*RenderNode)

	for _, c := range r.Commands {
		childrenMap[c.Description().Name] = &RenderNode{
			Name:     c.Description().Name,
			Command:  c,
			Children: nil,
		}
	}

	for k, v := range r.Children {
		existingNode, ok := childrenMap[k]
		newNode := v.ToRenderNode()
		newNode.Name = k
		if ok {
			// merge command
			newNode.Command = existingNode.Command
			childrenMap[k] = newNode
		} else {
			childrenMap[k] = newNode
		}
	}

	ret.Children = make([]*RenderNode, 0, len(childrenMap))
	for _, v := range childrenMap {
		ret.Children = append(ret.Children, v)
	}

	// sort by name
	sort.Slice(ret.Children, func(i, j int) bool {
		return ret.Children[i].Name < ret.Children[j].Name
	})

	return ret
}

// InsertNode inserts a node at the given prefix path
func (t *TrieNode) InsertNode(prefix []string, node *TrieNode) {
	current := t
	for _, component := range prefix {
		if child, ok := current.Children[component]; ok {
			current = child
		} else {
			newNode := NewTrieNode([]cmds.Command{}, nil)
			current.Children[component] = newNode
			current = newNode
		}
	}
	// Copy commands and children from the node to insert
	for k, v := range node.Children {
		current.Children[k] = v
	}
	current.Commands = append(current.Commands, node.Commands...)
}
