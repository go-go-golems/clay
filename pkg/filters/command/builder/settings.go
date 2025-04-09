package builder

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// FilterSettings contains all the filter parameters used by the list command.
// These correspond to the filter methods provided by the Builder.
type FilterSettings struct {
	Type          string   `glazed.parameter:"type" help:"Filter by command type"`
	Types         []string `glazed.parameter:"types" help:"Filter by multiple types (OR)"`
	Tag           string   `glazed.parameter:"tag" help:"Filter by single tag"`
	Tags          []string `glazed.parameter:"tags" help:"Filter by any of multiple tags (OR)"`
	AllTags       []string `glazed.parameter:"all-tags" help:"Must have all specified tags (AND)"`
	AnyTags       []string `glazed.parameter:"any-tags" help:"Must have any of specified tags (OR) (alias for --tags)"`
	Path          string   `glazed.parameter:"path" help:"Exact path match (e.g., 'queries es')"`
	PathGlob      string   `glazed.parameter:"path-glob" help:"Path glob pattern (e.g., 'queries/*')"`
	PathPrefix    string   `glazed.parameter:"path-prefix" help:"Path prefix match (e.g., 'queries/')"`
	Name          string   `glazed.parameter:"name" help:"Exact command name match (last part of path)"`
	NamePattern   string   `glazed.parameter:"name-pattern" help:"Command name pattern match (e.g., 'list*')"`
	MetadataKey   string   `glazed.parameter:"metadata-key" help:"Metadata key to match"`
	MetadataValue string   `glazed.parameter:"metadata-value" help:"Metadata value to match (requires --metadata-key)"`
}

// FilterLayerSlug is the slug for the filter parameter layer.
const FilterLayerSlug = "filter"

// NewFilterParameterLayer creates a new parameter layer for command filtering.
func NewFilterParameterLayer(options ...layers.ParameterLayerOptions) (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(FilterLayerSlug, "Command Filtering Options",
		append([]layers.ParameterLayerOptions{
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition(
					"type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by command type"),
				),
				parameters.NewParameterDefinition(
					"types",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by multiple types (OR)"),
				),
				parameters.NewParameterDefinition(
					"tag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by single tag"),
				),
				parameters.NewParameterDefinition(
					"tags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by any of multiple tags (OR)"),
				),
				parameters.NewParameterDefinition(
					"all-tags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Must have all specified tags (AND)"),
				),
				parameters.NewParameterDefinition(
					"any-tags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Must have any of specified tags (OR) (alias for --tags)"),
				),
				parameters.NewParameterDefinition(
					"path",
					parameters.ParameterTypeString,
					parameters.WithHelp("Exact path match (e.g., 'queries es')"),
				),
				parameters.NewParameterDefinition(
					"path-glob",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path glob pattern (e.g., 'queries/*')"),
				),
				parameters.NewParameterDefinition(
					"path-prefix",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path prefix match (e.g., 'queries/')"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithHelp("Exact command name match (last part of path)"),
				),
				parameters.NewParameterDefinition(
					"name-pattern",
					parameters.ParameterTypeString,
					parameters.WithHelp("Command name pattern match (e.g., 'list*')"),
				),
				parameters.NewParameterDefinition(
					"metadata-key",
					parameters.ParameterTypeString,
					parameters.WithHelp("Metadata key to match"),
				),
				parameters.NewParameterDefinition(
					"metadata-value",
					parameters.ParameterTypeString,
					parameters.WithHelp("Metadata value to match (requires --metadata-key)"),
				),
			),
		}, options...)...,
	)
}

// GetFilterSettingsFromParsedLayers extracts filter settings from parsed layers.
func GetFilterSettingsFromParsedLayers(parsedLayers *layers.ParsedLayers) (*FilterSettings, error) {
	s := &FilterSettings{}
	err := parsedLayers.InitializeStruct(FilterLayerSlug, s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize filter settings")
	}
	return s, nil
}

// BuildFilterFromSettings creates a filter from the settings provided via command-line flags.
// It uses the Builder to construct the appropriate query.
func BuildFilterFromSettings(s *FilterSettings, b *Builder) *FilterBuilder {
	var filters []*FilterBuilder

	// Add type filters
	if s.Type != "" {
		filters = append(filters, b.Type(s.Type))
	}
	if len(s.Types) > 0 {
		filters = append(filters, b.Types(s.Types...))
	}

	// Add tag filters
	if s.Tag != "" {
		filters = append(filters, b.Tag(s.Tag))
	}
	// Combine Tags and AnyTags as they are aliases
	tags := s.Tags
	if len(s.AnyTags) > 0 {
		tags = append(tags, s.AnyTags...)
	}
	if len(tags) > 0 {
		// Remove duplicates if any
		uniqueTags := map[string]struct{}{}
		for _, tag := range tags {
			uniqueTags[tag] = struct{}{}
		}
		finalTags := make([]string, 0, len(uniqueTags))
		for tag := range uniqueTags {
			finalTags = append(finalTags, tag)
		}
		if len(finalTags) > 0 {
			filters = append(filters, b.Tags(finalTags...))
		}
	}

	if len(s.AllTags) > 0 {
		filters = append(filters, b.AllTags(s.AllTags...))
	}

	// Add path filters
	if s.Path != "" {
		filters = append(filters, b.Path(s.Path))
	}
	if s.PathGlob != "" {
		filters = append(filters, b.PathGlob(s.PathGlob))
	}
	if s.PathPrefix != "" {
		filters = append(filters, b.PathPrefix(s.PathPrefix))
	}

	// Add name filters
	if s.Name != "" {
		filters = append(filters, b.Name(s.Name))
	}
	if s.NamePattern != "" {
		filters = append(filters, b.NamePattern(s.NamePattern))
	}

	// Add metadata filters
	if s.MetadataKey != "" && s.MetadataValue != "" {
		// TODO(manuel, 2024-07-18) Consider supporting non-string metadata values later
		filters = append(filters, b.Metadata(s.MetadataKey, s.MetadataValue))
	} else if s.MetadataKey != "" || s.MetadataValue != "" {
		// Warn if only one metadata flag is provided?
		log.Warn().Msg("Both --metadata-key and --metadata-value must be provided for metadata filtering")
	}

	// Combine all filters with AND logic
	if len(filters) == 0 {
		// No filters specified, match everything
		return b.MatchAll()
	}

	// Start with the first filter
	result := filters[0]

	// AND it with the rest
	for i := 1; i < len(filters); i++ {
		result = result.And(filters[i])
	}

	return result
}
