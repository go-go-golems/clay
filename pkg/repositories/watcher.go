package repositories

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/clay/pkg/watcher"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// getProcessedPaths takes a path and returns the processed paths needed for command operations
func (r *Repository) getProcessedPaths(path string) (string, string, []string, error) {
	filePath := path
	if len(path) > 0 && path[0] != '/' {
		var err error
		filePath, err = filepath.Abs(path)
		if err != nil {
			return "", "", nil, err
		}
	}

	// try to strip all r.Directories from path
	strippedPath := path
	for _, dir := range r.Directories {
		if strings.HasPrefix(path, dir.WatchDirectory) && dir.WatchDirectory != "." {
			strippedPath = strings.TrimPrefix(path, dir.WatchDirectory)
			break
		}
	}
	fullPath := strings.TrimPrefix(filePath, "/")

	// get directory of file
	parents := loaders.GetParentsFromDir(filepath.Dir(strippedPath))
	return filePath, fullPath, parents, nil
}

func (r *Repository) Watch(
	ctx context.Context,
	options ...watcher.Option,
) error {
	if r.loader == nil {
		return errors.New("no command loader set")
	}

	paths := []string{}
	for _, dir := range r.Directories {
		// if the watch directory is not set, skip
		if dir.WatchDirectory != "" {
			paths = append(paths, dir.WatchDirectory)
		}
	}
	paths = append(paths, r.Files...)

	options = append(options,
		watcher.WithWriteCallback(func(path string) error {
			log.Debug().Msgf("Loading %s", path)
			filePath, fullPath, parents, err := r.getProcessedPaths(path)
			if err != nil {
				return err
			}

			// Check if this is an individually tracked file
			isTrackedFile := false
			for _, f := range r.Files {
				if f == path {
					isTrackedFile = true
					break
				}
			}

			cmdOptions_ := []cmds.CommandDescriptionOption{
				cmds.WithSource(fullPath),
			}
			aliasOptions := []alias.Option{
				alias.WithSource(fullPath),
			}

			// Only add parents if this isn't a tracked file
			if !isTrackedFile {
				cmdOptions_ = append(cmdOptions_, cmds.WithParents(parents...))
				aliasOptions = append(aliasOptions, alias.WithParents(parents...))
			}

			fs_, filePath, err := loaders.FileNameToFsFilePath(filePath)
			if err != nil {
				return errors.Wrapf(err, "could not get fs and file path for %s", filePath)
			}

			commands, err := r.loader.LoadCommands(fs_, filePath, cmdOptions_, aliasOptions)
			if err != nil {
				return err
			}
			r.Add(commands...)
			return nil
		}),
		watcher.WithRemoveCallback(func(path string) error {
			log.Debug().Msgf("Removing %s", path)
			_, _, parents, err := r.getProcessedPaths(path)
			if err != nil {
				return err
			}

			// XXX we would actually need to map the command name to the file name because at this point we assume that the command name is the file base name.
			parents = append(parents, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
			r.Remove(parents)
			return nil
		}),
		watcher.WithPaths(paths...),
	)
	w := watcher.NewWatcher(options...)

	err := w.Run(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			return errors.Wrapf(err, "could not run watcher for repository: %s", strings.Join(paths, ","))
		}
		return err
	}
	return nil
}
