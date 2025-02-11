package yaml_editor

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLEditor provides utilities for manipulating YAML files while preserving comments and structure
type YAMLEditor struct {
	root *yaml.Node
}

// NewYAMLEditor creates a new YAMLEditor from raw YAML data
func NewYAMLEditor(data []byte) (*YAMLEditor, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("could not parse YAML: %w", err)
	}
	return &YAMLEditor{root: &root}, nil
}

// NewYAMLEditorFromFile creates a new YAMLEditor from a file
func NewYAMLEditorFromFile(filename string) (*YAMLEditor, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	return NewYAMLEditor(data)
}

// Save writes the YAML content to a file
func (e *YAMLEditor) Save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(e.root); err != nil {
		return fmt.Errorf("could not encode YAML: %w", err)
	}
	return nil
}

// GetNode returns the node at the given path
func (e *YAMLEditor) GetNode(path ...string) (*yaml.Node, error) {
	if len(path) == 0 {
		return e.root, nil
	}

	current := e.root
	if current.Kind == yaml.DocumentNode && len(current.Content) > 0 {
		current = current.Content[0]
	}

	for _, key := range path {
		if current.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("expected mapping node at path %v", path)
		}

		found := false
		for i := 0; i < len(current.Content); i += 2 {
			if current.Content[i].Value == key {
				current = current.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("key %s not found at path %v", key, path)
		}
	}

	return current, nil
}

// SetNode sets the node at the given path
func (e *YAMLEditor) SetNode(node *yaml.Node, path ...string) error {
	if len(path) == 0 {
		e.root = node
		return nil
	}

	parent, err := e.GetNode(path[:len(path)-1]...)
	if err != nil {
		return err
	}

	if parent.Kind != yaml.MappingNode {
		return fmt.Errorf("parent at path %v is not a mapping node", path[:len(path)-1])
	}

	lastKey := path[len(path)-1]
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == lastKey {
			parent.Content[i+1] = node
			return nil
		}
	}

	// Key doesn't exist, append it
	parent.Content = append(parent.Content,
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: lastKey,
		},
		node)

	return nil
}

// AppendToSequence appends a node to a sequence at the given path
func (e *YAMLEditor) AppendToSequence(node *yaml.Node, path ...string) error {
	target, err := e.GetNode(path...)
	if err != nil {
		return err
	}

	if target.Kind != yaml.SequenceNode {
		return fmt.Errorf("node at path %v is not a sequence node", path)
	}

	target.Content = append(target.Content, node)
	return nil
}

// RemoveFromSequence removes a node from a sequence at the given path and index
func (e *YAMLEditor) RemoveFromSequence(index int, path ...string) error {
	target, err := e.GetNode(path...)
	if err != nil {
		return err
	}

	if target.Kind != yaml.SequenceNode {
		return fmt.Errorf("node at path %v is not a sequence node", path)
	}

	if index < 0 || index >= len(target.Content) {
		return fmt.Errorf("index %d out of range for sequence at path %v", index, path)
	}

	target.Content = append(target.Content[:index], target.Content[index+1:]...)
	return nil
}

// GetMapNode returns the value node for a key in a mapping node
func (e *YAMLEditor) GetMapNode(key string, mapNode *yaml.Node) (*yaml.Node, error) {
	if mapNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("not a mapping node")
	}

	for i := 0; i < len(mapNode.Content); i += 2 {
		if mapNode.Content[i].Value == key {
			return mapNode.Content[i+1], nil
		}
	}

	return nil, fmt.Errorf("key %s not found", key)
}

// CreateMap creates a new mapping node with the given key-value pairs
func (e *YAMLEditor) CreateMap(pairs ...interface{}) (*yaml.Node, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("pairs must be key-value pairs")
	}

	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("key at index %d must be a string", i)
		}

		var valueNode *yaml.Node
		switch v := pairs[i+1].(type) {
		case string:
			valueNode = &yaml.Node{Kind: yaml.ScalarNode, Value: v}
		case int:
			valueNode = &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", v)}
		case bool:
			valueNode = &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", v)}
		case *yaml.Node:
			valueNode = v
		default:
			return nil, fmt.Errorf("unsupported value type at index %d", i+1)
		}

		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			valueNode)
	}

	return node, nil
}

// CreateSequence creates a new sequence node with the given items
func (e *YAMLEditor) CreateSequence(items ...*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: items,
	}
}

// CreateScalar creates a new scalar node with the given value
func (e *YAMLEditor) CreateScalar(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}
}

// DeepCopyNode creates a deep copy of a node
func (e *YAMLEditor) DeepCopyNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}

	copy_ := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor,
		Alias:       e.DeepCopyNode(node.Alias),
		Content:     make([]*yaml.Node, len(node.Content)),
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
		Line:        node.Line,
		Column:      node.Column,
	}

	for i, child := range node.Content {
		copy_.Content[i] = e.DeepCopyNode(child)
	}

	return copy_
}
