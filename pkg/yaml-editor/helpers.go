package yaml_editor

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// CreateMapNode creates a new mapping node with the given key-value pairs
func CreateMapNode(pairs ...interface{}) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("pairs must be key-value pairs")
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
		case float64:
			valueNode = &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%g", v)}
		case *yaml.Node:
			valueNode = v
		case []interface{}:
			seq := &yaml.Node{Kind: yaml.SequenceNode}
			for _, item := range v {
				switch it := item.(type) {
				case string:
					seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: it})
				case int:
					seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", it)})
				case bool:
					seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", it)})
				case float64:
					seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%g", it)})
				case *yaml.Node:
					seq.Content = append(seq.Content, it)
				default:
					return nil, fmt.Errorf("unsupported sequence item type at index %d", i+1)
				}
			}
			valueNode = seq
		case map[string]interface{}:
			var err error
			valueNode, err = CreateMapNode(mapToPairs(v)...)
			if err != nil {
				return nil, fmt.Errorf("error creating nested map at index %d: %w", i+1, err)
			}
		default:
			return nil, fmt.Errorf("unsupported value type at index %d", i+1)
		}

		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			valueNode)
	}

	return node, nil
}

// CreateSequenceNode creates a new sequence node with the given items
func CreateSequenceNode(items ...*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: items,
	}
}

// CreateScalarNode creates a new scalar node with the given value
func CreateScalarNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}
}

// DeepCopyNode creates a deep copy of a node
func DeepCopyNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}

	copy_ := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor,
		Alias:       DeepCopyNode(node.Alias),
		Content:     make([]*yaml.Node, len(node.Content)),
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
		Line:        node.Line,
		Column:      node.Column,
	}

	for i, child := range node.Content {
		copy_.Content[i] = DeepCopyNode(child)
	}

	return copy_
}

// GetNodeAtPath returns the node at the given path
func GetNodeAtPath(root *yaml.Node, path ...string) (*yaml.Node, error) {
	if len(path) == 0 {
		return root, nil
	}

	current := root
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

// mapToPairs converts a map to a slice of alternating keys and values
func mapToPairs(m map[string]interface{}) []interface{} {
	pairs := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		pairs = append(pairs, k, v)
	}
	return pairs
}

// SetComment sets a comment on a node
func SetComment(node *yaml.Node, comment string, position CommentPosition) {
	switch position {
	case CommentHead:
		node.HeadComment = comment
	case CommentLine:
		node.LineComment = comment
	case CommentFoot:
		node.FootComment = comment
	}
}

// CommentPosition specifies where to place a comment relative to a node
type CommentPosition int

const (
	CommentHead CommentPosition = iota // Comment before the node
	CommentLine                        // Comment at the end of the node's line
	CommentFoot                        // Comment after the node
)

// GetValueByPath gets a value from a YAML node by path, with type conversion
func GetValueByPath(root *yaml.Node, path ...string) (interface{}, error) {
	node, err := GetNodeAtPath(root, path...)
	if err != nil {
		return nil, err
	}

	//nolint:exhaustive
	switch node.Kind {
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str":
			return node.Value, nil
		case "!!int":
			var v int
			if err := node.Decode(&v); err != nil {
				return nil, err
			}
			return v, nil
		case "!!bool":
			var v bool
			if err := node.Decode(&v); err != nil {
				return nil, err
			}
			return v, nil
		case "!!float":
			var v float64
			if err := node.Decode(&v); err != nil {
				return nil, err
			}
			return v, nil
		default:
			return node.Value, nil
		}
	case yaml.SequenceNode:
		var v []interface{}
		if err := node.Decode(&v); err != nil {
			return nil, err
		}
		return v, nil
	case yaml.MappingNode:
		var v map[string]interface{}
		if err := node.Decode(&v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported node kind: %v", node.Kind)
	}
}
