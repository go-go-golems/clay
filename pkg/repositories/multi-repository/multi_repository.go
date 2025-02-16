package multi_repository

import (
	"context"
	"path"
	"strings"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/clay/pkg/repositories/mcp"
	"github.com/go-go-golems/clay/pkg/repositories/trie"
	"github.com/go-go-golems/clay/pkg/watcher"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type MountedRepository struct {
	Path       string
	Repository repositories.RepositoryInterface
}

type MultiRepository struct {
	repositories []MountedRepository
}

func NewMultiRepository() *MultiRepository {
	return &MultiRepository{
		repositories: []MountedRepository{},
	}
}

func (m *MultiRepository) Mount(mountPath string, repo repositories.RepositoryInterface) {
	// Ensure the path starts with a slash and doesn't end with one
	mountPath = path.Clean("/" + mountPath)
	m.repositories = append(m.repositories, MountedRepository{
		Path:       mountPath,
		Repository: repo,
	})
}

func (m *MultiRepository) Unmount(mountPath string) {
	mountPath = path.Clean("/" + mountPath)
	for i, repo := range m.repositories {
		if repo.Path == mountPath {
			m.repositories = append(m.repositories[:i], m.repositories[i+1:]...)
			return
		}
	}
}

func (m *MultiRepository) LoadCommands(helpSystem *help.HelpSystem, options ...cmds.CommandDescriptionOption) error {
	for _, repo := range m.repositories {
		if err := repo.Repository.LoadCommands(helpSystem, options...); err != nil {
			return errors.Wrapf(err, "failed to load commands for repository mounted at %s", repo.Path)
		}
	}
	return nil
}

func (m *MultiRepository) Add(commands ...cmds.Command) {
	// For now, add commands to the first repository
	// TODO(manuel) - might want to make this smarter
	if len(m.repositories) > 0 {
		m.repositories[0].Repository.Add(commands...)
	} else {
		log.Warn().Msg("attempting to add commands to empty multi-repository")
	}
}

func (m *MultiRepository) Remove(prefixes ...[]string) {
	for _, repo := range m.repositories {
		repo.Repository.Remove(prefixes...)
	}
}

func (m *MultiRepository) CollectCommands(prefix []string, recurse bool) []cmds.Command {
	var allCommands []cmds.Command

	// If prefix is empty or "/", collect from all repositories
	if len(prefix) == 0 || (len(prefix) == 1 && prefix[0] == "/") {
		for _, repo := range m.repositories {
			commands := repo.Repository.CollectCommands([]string{}, recurse)
			// Prepend mount path to each command's parents, unless it's root mounted
			for _, cmd := range commands {
				desc := cmd.Description()
				if repo.Path != "/" {
					desc.Parents = append(strings.Split(repo.Path, "/")[1:], desc.Parents...)
				}
			}
			allCommands = append(allCommands, commands...)
		}
		return allCommands
	}

	// Otherwise, find the appropriate repository and delegate
	for _, repo := range m.repositories {
		mountComponents := strings.Split(repo.Path, "/")[1:] // Skip empty first component
		if len(prefix) >= len(mountComponents) {
			match := true
			for i, comp := range mountComponents {
				if prefix[i] != comp {
					match = false
					break
				}
			}
			if match {
				// Remove mount path from prefix
				subPrefix := prefix[len(mountComponents):]
				commands := repo.Repository.CollectCommands(subPrefix, recurse)
				// Prepend mount path to each command's parents, unless it's root mounted
				for _, cmd := range commands {
					desc := cmd.Description()
					if repo.Path != "/" {
						desc.Parents = append(strings.Split(repo.Path, "/")[1:], desc.Parents...)
					}
				}
				allCommands = append(allCommands, commands...)
			}
		}
	}

	return allCommands
}

func (m *MultiRepository) GetCommand(name string) (cmds.Command, bool) {
	if name == "" {
		return nil, false
	}

	// Handle absolute paths
	name = path.Clean("/" + name)
	for _, repo := range m.repositories {
		if strings.HasPrefix(name, repo.Path) {
			subPath := strings.TrimPrefix(name, repo.Path)
			if subPath == "" {
				return nil, false
			}
			subPath = strings.TrimPrefix(subPath, "/")
			return repo.Repository.GetCommand(subPath)
		}
	}

	return nil, false
}

func (m *MultiRepository) FindNode(prefix []string) *trie.TrieNode {
	if len(prefix) == 0 {
		// Create a root node that contains all mounted repositories
		root := trie.NewTrieNode([]cmds.Command{}, nil)
		for _, repo := range m.repositories {
			mountComponents := strings.Split(repo.Path, "/")[1:] // Skip empty first component
			if len(mountComponents) > 0 {
				subNode := repo.Repository.FindNode([]string{})
				if subNode != nil {
					root.InsertNode(mountComponents, subNode)
				}
			}
		}
		return root
	}

	for _, repo := range m.repositories {
		mountComponents := strings.Split(repo.Path, "/")[1:] // Skip empty first component
		if len(prefix) >= len(mountComponents) {
			match := true
			for i, comp := range mountComponents {
				if prefix[i] != comp {
					match = false
					break
				}
			}
			if match {
				subPrefix := prefix[len(mountComponents):]
				return repo.Repository.FindNode(subPrefix)
			}
		}
	}

	return nil
}

func (m *MultiRepository) GetRenderNode(prefix []string) (*trie.RenderNode, bool) {
	if len(prefix) == 0 {
		// Create a root render node that contains all mounted repositories
		root := &trie.RenderNode{
			Name:     "/",
			Children: make([]*trie.RenderNode, 0),
		}
		for _, repo := range m.repositories {
			mountComponents := strings.Split(repo.Path, "/")[1:] // Skip empty first component
			if len(mountComponents) > 0 {
				renderNode, ok := repo.Repository.GetRenderNode([]string{})
				if ok {
					renderNode.Name = mountComponents[len(mountComponents)-1]
					root.Children = append(root.Children, renderNode)
				}
			}
		}
		return root, true
	}

	for _, repo := range m.repositories {
		mountComponents := strings.Split(repo.Path, "/")[1:] // Skip empty first component
		if len(prefix) >= len(mountComponents) {
			match := true
			for i, comp := range mountComponents {
				if prefix[i] != comp {
					match = false
					break
				}
			}
			if match {
				subPrefix := prefix[len(mountComponents):]
				return repo.Repository.GetRenderNode(subPrefix)
			}
		}
	}

	return nil, false
}

func (m *MultiRepository) ListTools(ctx context.Context, cursor string) ([]mcp.Tool, string, error) {
	var allTools []mcp.Tool
	for _, repo := range m.repositories {
		tools, _, err := repo.Repository.ListTools(ctx, cursor)
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to list tools for repository mounted at %s", repo.Path)
		}

		// Prepend mount path to each tool's name, unless it's root mounted
		for i := range tools {
			if repo.Path != "/" {
				tools[i].Name = path.Join(repo.Path, tools[i].Name)
			}
		}
		allTools = append(allTools, tools...)
	}

	return allTools, "", nil
}

func (m *MultiRepository) Watch(ctx context.Context, options ...watcher.Option) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, repo := range m.repositories {
		r := repo // Create new variable to avoid closure issues
		g.Go(func() error {
			if err := r.Repository.Watch(ctx, options...); err != nil {
				return errors.Wrapf(err, "failed to watch repository mounted at %s", r.Path)
			}
			return nil
		})
	}

	return g.Wait()
}

var _ repositories.RepositoryInterface = (*MultiRepository)(nil)
