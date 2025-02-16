package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindNode(t *testing.T) {
	tests := []struct {
		name       string
		setupNodes map[string]*TrieNode
		findPrefix []string
		shouldFind bool
	}{
		{
			name: "root node",
			setupNodes: map[string]*TrieNode{
				"/": NewTrieNode(nil, nil),
			},
			findPrefix: []string{},
			shouldFind: true,
		},
		{
			name: "mounted node",
			setupNodes: map[string]*TrieNode{
				"/test": NewTrieNode(nil, nil),
			},
			findPrefix: []string{"test"},
			shouldFind: true,
		},
		{
			name: "nonexistent node",
			setupNodes: map[string]*TrieNode{
				"/": NewTrieNode(nil, nil),
			},
			findPrefix: []string{"nonexistent"},
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()

			for path, node := range tt.setupNodes {
				mockRepo := NewMockRepository(nil)
				mockRepo.findNodeRet = node
				mr.Mount(path, mockRepo)
			}

			node := mr.FindNode(tt.findPrefix)
			if tt.shouldFind {
				assert.NotNil(t, node)
			} else {
				assert.Nil(t, node)
			}
		})
	}
}

func TestGetRenderNode(t *testing.T) {
	tests := []struct {
		name         string
		setupNodes   map[string]*RenderNode
		findPrefix   []string
		shouldFind   bool
		expectedName string
	}{
		{
			name: "root render node",
			setupNodes: map[string]*RenderNode{
				"/": {
					Name:     "root",
					Children: []*RenderNode{},
				},
			},
			findPrefix:   []string{},
			shouldFind:   true,
			expectedName: "/",
		},
		{
			name: "mounted render node",
			setupNodes: map[string]*RenderNode{
				"/test": {
					Name:     "test",
					Children: []*RenderNode{},
				},
			},
			findPrefix:   []string{"test"},
			shouldFind:   true,
			expectedName: "test",
		},
		{
			name: "nonexistent render node",
			setupNodes: map[string]*RenderNode{
				"/": {
					Name:     "root",
					Children: []*RenderNode{},
				},
			},
			findPrefix: []string{"nonexistent"},
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMultiRepository()

			for path, node := range tt.setupNodes {
				mockRepo := NewMockRepository(nil)
				mockRepo.renderNode = node
				mockRepo.renderNodeOk = true
				mr.Mount(path, mockRepo)
			}

			node, found := mr.GetRenderNode(tt.findPrefix)
			assert.Equal(t, tt.shouldFind, found)

			if tt.shouldFind {
				assert.Equal(t, tt.expectedName, node.Name)
			}
		})
	}
}
