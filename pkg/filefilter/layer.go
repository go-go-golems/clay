package filefilter

import (
	"fmt"
	"os"

	"github.com/denormal/go-gitignore"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

type FileFilterSettings struct {
	MaxFileSize           int64    `glazed.parameter:"max-file-size"`
	DisableGitIgnore      bool     `glazed.parameter:"disable-gitignore"`
	DisableDefaultFilters bool     `glazed.parameter:"disable-default-filters"`
	Include               []string `glazed.parameter:"include"`
	Exclude               []string `glazed.parameter:"exclude"`
	MatchFilename         []string `glazed.parameter:"match-filename"`
	MatchPath             []string `glazed.parameter:"match-path"`
	ExcludeDirs           []string `glazed.parameter:"exclude-dirs"`
	ExcludeMatchFilename  []string `glazed.parameter:"exclude-match-filename"`
	ExcludeMatchPath      []string `glazed.parameter:"exclude-match-path"`
	FilterBinary          bool     `glazed.parameter:"filter-binary"`
	Verbose               bool     `glazed.parameter:"verbose"`
}

const FileFilterSlug = "file-filter"

func NewFileFilterParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		FileFilterSlug,
		"File Filter Options",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"max-file-size",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum size of individual files in bytes"),
				parameters.WithDefault(int64(1024*1024)),
			),
			parameters.NewParameterDefinition(
				"disable-gitignore",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Disable .gitignore filter"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"disable-default-filters",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Disable default file and directory filters"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"include",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of file extensions to include (e.g., .go,.js)"),
				parameters.WithShortFlag("i"),
			),
			parameters.NewParameterDefinition(
				"exclude",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of file extensions to exclude (e.g., .exe,.dll)"),
				parameters.WithShortFlag("e"),
			),
			parameters.NewParameterDefinition(
				"match-filename",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of regular expressions to match filenames"),
				parameters.WithShortFlag("f"),
			),
			parameters.NewParameterDefinition(
				"match-path",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of regular expressions to match full paths"),
				parameters.WithShortFlag("p"),
			),
			parameters.NewParameterDefinition(
				"exclude-dirs",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of directories to exclude"),
				parameters.WithShortFlag("x"),
			),
			parameters.NewParameterDefinition(
				"exclude-match-filename",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of regular expressions to exclude matching filenames"),
				parameters.WithShortFlag("F"),
			),
			parameters.NewParameterDefinition(
				"exclude-match-path",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of regular expressions to exclude matching full paths"),
				parameters.WithShortFlag("P"),
			),
			parameters.NewParameterDefinition(
				"filter-binary",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Filter out binary files"),
				parameters.WithDefault(true),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose logging of filtered/unfiltered paths"),
				parameters.WithDefault(false),
				parameters.WithShortFlag("v"),
			),
		),
	)
}

func CreateFileFilterFromSettings(parsedLayer *layers.ParsedLayer) (*FileFilter, error) {
	s := &FileFilterSettings{}
	err := parsedLayer.InitializeStruct(s)
	if err != nil {
		return nil, err
	}

	ff := NewFileFilter()

	ff.MaxFileSize = s.MaxFileSize
	ff.IncludeExts = s.Include
	ff.ExcludeExts = s.Exclude
	ff.MatchFilenames = compileRegexps(s.MatchFilename)
	ff.MatchPaths = compileRegexps(s.MatchPath)
	ff.ExcludeDirs = s.ExcludeDirs
	ff.ExcludeMatchFilenames = compileRegexps(s.ExcludeMatchFilename)
	ff.ExcludeMatchPaths = compileRegexps(s.ExcludeMatchPath)
	ff.DisableGitIgnore = s.DisableGitIgnore
	ff.DisableDefaultFilters = s.DisableDefaultFilters
	ff.Verbose = s.Verbose
	ff.FilterBinaryFiles = s.FilterBinary

	if !ff.DisableGitIgnore {
		gitIgnoreFilter, err := initGitIgnoreFilter()
		if err != nil {
			return nil, fmt.Errorf("error initializing gitignore filter: %w", err)
		}
		ff.GitIgnoreFilter = gitIgnoreFilter
	}

	return ff, nil
}

func initGitIgnoreFilter() (gitignore.GitIgnore, error) {
	if _, err := os.Stat(".gitignore"); err == nil {
		gitIgnoreFilter, err := gitignore.NewFromFile(".gitignore")
		if err != nil {
			return nil, fmt.Errorf("error initializing gitignore filter from file: %w", err)
		}
		return gitIgnoreFilter, nil
	}

	gitIgnoreFilter, err := gitignore.NewRepository(".")
	if err != nil {
		return nil, fmt.Errorf("error initializing gitignore filter: %w", err)
	}
	return gitIgnoreFilter, nil
}
