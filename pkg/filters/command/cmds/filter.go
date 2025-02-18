package cmds

import (
	"context"
	"strings"

	"github.com/go-go-golems/clay/pkg/filters/command"
	"github.com/go-go-golems/clay/pkg/filters/command/builder"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
)

// FilterCommand implements a glazed command for filtering other commands
type FilterCommand struct {
	*cmds.CommandDescription
	commands []*cmds.CommandDescription
	index    *command.CommandIndex
}

// FilterSettings contains all the filter parameters
type FilterSettings struct {
	Type          string   `glazed.parameter:"type"`
	Types         []string `glazed.parameter:"types"`
	Tag           string   `glazed.parameter:"tag"`
	Tags          []string `glazed.parameter:"tags"`
	AllTags       []string `glazed.parameter:"all-tags"`
	AnyTags       []string `glazed.parameter:"any-tags"`
	Path          string   `glazed.parameter:"path"`
	PathGlob      string   `glazed.parameter:"path-glob"`
	PathPrefix    string   `glazed.parameter:"path-prefix"`
	Name          string   `glazed.parameter:"name"`
	NamePattern   string   `glazed.parameter:"name-pattern"`
	MetadataKey   string   `glazed.parameter:"metadata-key"`
	MetadataValue string   `glazed.parameter:"metadata-value"`
}

// NewFilterCommand creates a new filter command with the given list of commands to filter
func NewFilterCommand(commands []*cmds.CommandDescription) (*FilterCommand, error) {
	// Create the command index
	index, err := command.NewCommandIndex(commands)
	if err != nil {
		return nil, errors.Wrap(err, "could not create command index")
	}

	// Create glazed parameter layer
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &FilterCommand{
		CommandDescription: cmds.NewCommandDescription(
			"filter",
			cmds.WithShort("Filter commands based on various criteria"),
			cmds.WithLong(`Filter commands using various criteria like type, tags, path, etc.
Supports complex filtering with pattern matching and metadata search.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by command type"),
				),
				parameters.NewParameterDefinition(
					"types",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by multiple types"),
				),
				parameters.NewParameterDefinition(
					"tag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by single tag"),
				),
				parameters.NewParameterDefinition(
					"tags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by multiple tags"),
				),
				parameters.NewParameterDefinition(
					"all-tags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Must have all specified tags"),
				),
				parameters.NewParameterDefinition(
					"any-tags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Must have any of specified tags"),
				),
				parameters.NewParameterDefinition(
					"path",
					parameters.ParameterTypeString,
					parameters.WithHelp("Exact path match"),
				),
				parameters.NewParameterDefinition(
					"path-glob",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path glob pattern"),
				),
				parameters.NewParameterDefinition(
					"path-prefix",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path prefix match"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithHelp("Exact name match"),
				),
				parameters.NewParameterDefinition(
					"name-pattern",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name pattern match"),
				),
				parameters.NewParameterDefinition(
					"metadata-key",
					parameters.ParameterTypeString,
					parameters.WithHelp("Metadata key to match"),
				),
				parameters.NewParameterDefinition(
					"metadata-value",
					parameters.ParameterTypeString,
					parameters.WithHelp("Metadata value to match"),
				),
			),
			cmds.WithLayersList(
				glazedLayer,
			),
		),
		commands: commands,
		index:    index,
	}, nil
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *FilterCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &FilterSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "could not initialize filter settings")
	}

	// Build filter
	b := builder.New()
	filter := c.buildFilterFromSettings(s, b)

	// Execute search
	matches, err := c.index.Search(ctx, filter, c.commands)
	if err != nil {
		return errors.Wrap(err, "could not search commands")
	}

	// Output results as rows
	for _, cmd := range matches {
		row := types.NewRow(
			types.MRP("name", cmd.Name),
			types.MRP("type", cmd.Type),
			types.MRP("path", cmd.FullPath()),
			types.MRP("tags", strings.Join(cmd.Tags, ",")),
			types.MRP("short", cmd.Short),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrap(err, "could not add row")
		}
	}

	return nil
}

// buildFilterFromSettings creates a filter from the settings
func (c *FilterCommand) buildFilterFromSettings(s *FilterSettings, b *builder.Builder) *builder.FilterBuilder {
	var filters []*builder.FilterBuilder

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
	if len(s.Tags) > 0 {
		filters = append(filters, b.Tags(s.Tags...))
	}
	if len(s.AllTags) > 0 {
		filters = append(filters, b.AllTags(s.AllTags...))
	}
	if len(s.AnyTags) > 0 {
		filters = append(filters, b.AnyTags(s.AnyTags...))
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
		filters = append(filters, b.Metadata(s.MetadataKey, s.MetadataValue))
	}

	// Combine all filters with AND
	if len(filters) == 0 {
		return b.MatchAll()
	}

	result := filters[0]
	for i := 1; i < len(filters); i++ {
		result = result.And(filters[i])
	}
	return result
}
