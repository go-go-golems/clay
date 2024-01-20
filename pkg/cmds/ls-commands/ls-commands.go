package ls_commands

import (
	"context"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"strings"
)

type AddCommandToRowFunc func(cmd glazed_cmds.Command, row types.Row, parsedLayers *layers.ParsedLayers) ([]types.Row, error)

type ListCommandsCommand struct {
	*glazed_cmds.CommandDescription
	commands            []glazed_cmds.Command
	AddCommandToRowFunc AddCommandToRowFunc
}

var _ glazed_cmds.GlazeCommand = (*ListCommandsCommand)(nil)

type ListCommandsCommandOption func(*ListCommandsCommand) error

func WithCommandDescriptionOptions(options ...glazed_cmds.CommandDescriptionOption) ListCommandsCommandOption {
	return func(q *ListCommandsCommand) error {
		description := q.CommandDescription.Description()
		for _, option := range options {
			option(description)
		}
		return nil
	}
}

func WithAddCommandToRowFunc(f AddCommandToRowFunc) ListCommandsCommandOption {
	return func(q *ListCommandsCommand) error {
		q.AddCommandToRowFunc = f
		return nil
	}
}

func NewListCommandsCommand(
	allCommands []glazed_cmds.Command,
	options ...ListCommandsCommandOption,
) (*ListCommandsCommand, error) {
	glazeParameterLayer, err := settings.NewGlazedParameterLayers(
		settings.WithFieldsFiltersParameterLayerOptions(
			layers.WithDefaults(
				&settings.FieldsFilterFlagsDefaults{
					Fields: []string{"name", "type", "short", "source"},
				},
			),
		),
	)
	if err != nil {
		return nil, err
	}

	ret := &ListCommandsCommand{
		commands: allCommands,
		CommandDescription: glazed_cmds.NewCommandDescription(
			"ls-commands", glazed_cmds.WithLayersList(glazeParameterLayer),
		),
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func (q *ListCommandsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	tableProcessor, ok := gp.(*middlewares.TableProcessor)
	if !ok {
		return errors.New("expected a table processor")
	}

	// check if there already is a ReorderColumnOrderMiddleware in the row processors
	hasReorderColumnOrderMiddleware := false
	for _, middleware := range tableProcessor.RowMiddlewares {
		if _, ok := middleware.(*row.ReorderColumnOrderMiddleware); ok {
			hasReorderColumnOrderMiddleware = true
			break
		}
	}
	if !hasReorderColumnOrderMiddleware {
		tableProcessor.AddRowMiddleware(
			row.NewReorderColumnOrderMiddleware(
				[]string{"name", "short", "long", "source", "query"}),
		)
	}

	for _, command := range q.commands {
		description := command.Description()
		obj := types.NewRow(
			types.MRP("name", strings.Join(append(description.Parents, description.Name), " ")),
			types.MRP("short", description.Short),
			types.MRP("long", description.Long),
			types.MRP("source", description.Source),
			types.MRP("type", "unknown"),
			types.MRP("parents", description.Parents),
		)

		switch c := command.(type) {
		case *alias.CommandAlias:
			obj.Set("aliasFor", c.AliasFor)
			obj.Set("type", "alias")
		default:
		}

		rows := []types.Row{obj}
		if q.AddCommandToRowFunc != nil {
			var err error
			rows, err = q.AddCommandToRowFunc(command, obj, parsedLayers)
			if err != nil {
				return err
			}
		}

		for _, row_ := range rows {
			err := gp.AddRow(ctx, row_)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
