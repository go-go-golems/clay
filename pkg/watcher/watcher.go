package watcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type WriteCallback func(path string) error
type RemoveCallback func(path string) error

// Watcher provides a way to recursively watch a set of paths for changes.
// It recursively adds all directories present (and created in the future)
// to provide coverage.
//
// You can provide a doublestar mask to filter out paths. For example, to
// only watch for changes to .txt files, you can provide "**/*.txt".
type Watcher struct {
	paths          []string
	masks          []string
	watchedDirs    map[string]bool // track directories we're already watching
	writeCallback  WriteCallback
	removeCallback RemoveCallback
	breakOnError   bool
}

// Run is a blocking loop that will watch the paths provided and call the
func (w *Watcher) Run(ctx context.Context) error {
	if w.writeCallback == nil {
		return errors.New("no writeCallback provided")
	}

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	// Add each path to the watcher
	for _, path := range w.paths {
		log.Debug().Str("path", path).Msg("Adding recursive path to watcher")
		err = w.addRecursive(watcher, path)
		if err != nil {
			return err
		}
	}

	log.Info().Strs("paths", w.paths).Strs("masks", w.masks).Msg("Watching paths")

	// Listen for events until the context is cancelled
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Context cancelled, stopping watcher")
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			log.Debug().Str("op", event.Op.String()).
				Str("name", event.Name).
				//Strs("watchList", watcher.WatchList()).
				Msg("Received fsnotify event")

			// if it is a deletion, remove the directory from the watcher
			// TODO(manuel, 2023-03-27) There's a race condition here where a rename is a Create followed by a Remove.
			// See also https://github.com/go-go-golems/cliopatra/issues/10
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				err = w.removePathsWithPrefix(watcher, event.Name)
				if err != nil {
					log.Warn().Err(err).Str("path", event.Name).Msg("Could not remove path from watcher")
					if w.breakOnError {
						return err
					}
				}
			}

			if event.Op&fsnotify.Rename == fsnotify.Rename {
				err = w.removePathsWithPrefix(watcher, event.Name)
				if err != nil {
					if errno, ok := err.(syscall.Errno); ok && errno == syscall.EINVAL {
						// This means that the file was already deleted, and the inotify already removed,
						// which can happen on a rename in linux.
						continue
					}
					log.Warn().Err(err).Str("path", event.Name).Msg("Could not remove path from watcher")
					if w.breakOnError {
						return err
					}
				}
			}

			// if a new directory is created, add it to the watcher
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)

				if err != nil {
					log.Debug().Err(err).Str("path", event.Name).Msg("Could not stat path")
					continue
				}

				// if a directory was created, we always add it. For files, we must first check if it matches
				// the globs.
				if info.IsDir() {
					log.Debug().Str("path", event.Name).Msg("Adding new directory to watcher")
					err = w.addRecursive(watcher, event.Name)
					if err != nil {
						log.Warn().Err(err).Str("path", event.Name).Msg("Could not add directory to watcher")
						if w.breakOnError {
							return err
						}
					}
					continue
				}
			}

			// check if the path matches the globs
			if len(w.masks) > 0 {
				matched := false
				for _, mask := range w.masks {
					doesMatch, err := doublestar.Match(mask, event.Name)
					if err != nil {
						log.Warn().Err(err).Str("path", event.Name).Str("mask", mask).Msg("Could not match path with mask")
						if w.breakOnError {
							return err
						}
						break
					}

					if doesMatch {
						matched = true
						break
					}

				}

				if !matched {
					log.Debug().Str("path", event.Name).Strs("masks", w.masks).Msg("Skipping event because it does not match the mask")
					continue
				}
			}

			if event.Name == "" {
				// This is a workaround for a bug in fsnotify where the name is empty.
				// This might be a race condition with removing the renamed file and adding it again.
				// The Rename event is triggered by MOVE_FROM. Ignoring the empty path seems to work fine,
				// the Rename on the proper path is triggered afterwards.
				continue
			}

			// if the new file is valid, add it to the watcher for changes and removal
			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Debug().Str("path", event.Name).Msg("Adding path to watchlist")
				err = watcher.Add(event.Name)
				if err != nil {
					log.Warn().Err(err).Str("path", event.Name).Msg("Could not add path to watcher")
					if w.breakOnError {
						return err
					}
				}
			}

			isWriteEvent := event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create
			isRemoveEvent := event.Op&fsnotify.Rename == fsnotify.Rename || event.Op&fsnotify.Remove == fsnotify.Remove

			if isWriteEvent && w.writeCallback != nil {
				err = w.writeCallback(event.Name)
				if err != nil {
					log.Warn().Err(err).Str("path", event.Name).Msg("Error while processing write event")
					if w.breakOnError {
						return err
					}
				}
			}

			if isRemoveEvent && w.removeCallback != nil {
				err = w.removeCallback(event.Name)
				if err != nil {
					log.Warn().Err(err).Str("path", event.Name).Msg("Error while processing remove event")
					if w.breakOnError {
						return err
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Error().Err(err).Msg("Received fsnotify error")
			if w.breakOnError {
				return err
			}
		}
	}
}

type Option func(w *Watcher)

func WithPaths(paths ...string) Option {
	return func(w *Watcher) {
		w.paths = append(w.paths, paths...)
	}
}

func WithMask(masks ...string) Option {
	return func(w *Watcher) {
		w.masks = masks
	}
}

func WithWriteCallback(callback WriteCallback) Option {
	return func(w *Watcher) {
		w.writeCallback = callback
	}
}

func WithRemoveCallback(callback RemoveCallback) Option {
	return func(w *Watcher) {
		w.removeCallback = callback
	}
}

func WithBreakOnError(breakOnError bool) Option {
	return func(w *Watcher) {
		w.breakOnError = breakOnError
	}
}

func NewWatcher(options ...Option) *Watcher {
	ret := &Watcher{
		paths:       []string{},
		masks:       []string{},
		watchedDirs: make(map[string]bool),
	}

	for _, opt := range options {
		opt(ret)
	}

	return ret
}

// removePathsWithPrefix removes `name` and all subdirectories from the watcher
func (w *Watcher) removePathsWithPrefix(watcher *fsnotify.Watcher, name string) error {
	if name == "" {
		log.Debug().Msg("Ignoring empty prefixes")
		return nil
	}

	watchlist := watcher.WatchList()
	log.Debug().Str("name", name).Msg("Removing paths with prefix")

	for _, path := range watchlist {
		if strings.HasPrefix(path, name) {
			log.Debug().Str("path", path).Msg("Removing path from watcher")
			err := watcher.Remove(path)
			if err != nil {
				return err
			}
			delete(w.watchedDirs, path)
		}
	}

	return nil
}

// Recursively add a path to the watcher
func (w *Watcher) addRecursive(watcher *fsnotify.Watcher, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Directory handling
		dirPath := path
		if !strings.HasSuffix(dirPath, string(os.PathSeparator)) {
			dirPath += string(os.PathSeparator)
		}

		if !w.watchedDirs[dirPath] {
			err = watcher.Add(dirPath)
			if err != nil {
				return err
			}
			w.watchedDirs[dirPath] = true
			log.Debug().Str("path", dirPath).Msg("Added directory to watcher")
		}

		// Continue with recursive directory handling...
		err = filepath.Walk(dirPath, func(subpath string, info os.FileInfo, err error) error {
			if err != nil {
				log.Warn().Err(err).Str("path", subpath).Msg("Error walking path")
				return nil
			}
			if subpath == dirPath {
				return nil
			}
			log.Trace().Str("path", subpath).Msg("Testing subpath to watcher")
			if info.IsDir() {
				log.Debug().Str("path", subpath).Msg("Adding subpath to watcher")
				err = w.addRecursive(watcher, subpath)
				if err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		// File handling - add parent directory
		parentDir := filepath.Dir(path)
		if !strings.HasSuffix(parentDir, string(os.PathSeparator)) {
			parentDir += string(os.PathSeparator)
		}

		if !w.watchedDirs[parentDir] {
			err = watcher.Add(parentDir)
			if err != nil {
				return err
			}
			w.watchedDirs[parentDir] = true
			log.Debug().Str("path", parentDir).Msg("Added parent directory to watcher for file")
		}
	}
	return nil
}
