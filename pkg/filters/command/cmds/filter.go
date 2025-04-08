package cmds

import (
	"context"
	"strings"

	"github.com/go-go-golems/clay/pkg/filters/command"
	"github.com/go-go-golems/clay/pkg/filters/command/builder"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
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

	// Create filter parameter layer
	filterLayer, err := builder.NewFilterParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create filter parameter layer")
	}

	return &FilterCommand{
		CommandDescription: cmds.NewCommandDescription(
			"filter",
			cmds.WithShort("Filter commands based on various criteria"),
			cmds.WithLong(`Filter commands using various criteria like type, tags, path, etc.
Supports complex filtering with pattern matching and metadata search.`),
			cmds.WithLayersList(
				glazedLayer,
				filterLayer,
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
	s, err := builder.GetFilterSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "could not initialize filter settings")
	}

	// Build filter
	b := builder.New()
	filter := builder.BuildFilterFromSettings(s, b)

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
