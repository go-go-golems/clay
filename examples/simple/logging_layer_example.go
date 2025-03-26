package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ExampleCommand implements a simple command that uses the logging layer
type ExampleCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*ExampleCommand)(nil)

func (c *ExampleCommand) Description() *cmds.CommandDescription {
	return c.CommandDescription
}

func (c *ExampleCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Log some information to show different log levels
	log.Debug().Msg("Running example command")
	log.Info().Msg("Running example command")
	log.Warn().Msg("This is a warning message")
	log.Error().Msg("This is an error message")

	// Don't use Fatal in an example as it will exit the program
	// log.Fatal().Msg("This is a fatal message")

	return nil
}

// NewExampleCommand creates a new example command
func NewExampleCommand() (*ExampleCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &ExampleCommand{
		CommandDescription: cmds.NewCommandDescription(
			"example",
			cmds.WithShort("Example command showing logging layer usage"),
			cmds.WithLong("This command demonstrates how to use the logging layer in a Glazed command. "+
				"It supports various logging formats, levels, and outputs including file and Logstash integration."),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}

	return cmd, nil
}

// Ensure ExampleCommand implements the Command interface
var _ cmds.Command = (*ExampleCommand)(nil)

func main() {
	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "logging-example",
		Short: "Example application with logging layer",
		Long: `Example application demonstrating the logging layer capabilities.
The logging layer supports:
- Different log levels (debug, info, warn, error, fatal)
- Various output formats (text, json)
- File logging with rotation
- Logstash integration for centralized logging`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := logging.InitLoggerFromViper()
			if err != nil {
				log.Error().Msgf("Error initializing logger: %v", err)
			}
			log.Debug().Msg("PersistentPreRun from main")
		},
	}

	// Set up Viper and initialize logging
	err := pkg.InitViper("logging-example", rootCmd)
	if err != nil {
		fmt.Printf("Error initializing viper: %v\n", err)
		os.Exit(1)
	}

	log.Debug().Msg("Debug message from main")
	log.Info().Msg("Info message from main")
	log.Warn().Msg("Warn message from main")
	log.Error().Msg("Error message from main")

	// Create the example command
	exampleCmd, err := NewExampleCommand()
	if err != nil {
		fmt.Printf("Error creating example command: %v\n", err)
		os.Exit(1)
	}

	// Method 1: Build a Cobra command with the logging layer
	// This adds the logging layer to the command and configures the help text
	cobraCmd, err := cli.BuildCobraCommandFromCommand(exampleCmd)
	if err != nil {
		fmt.Printf("Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	// Method 2: Add logging layer to the command directly and then build a Cobra command
	// This is useful if you need to customize the command further
	anotherExampleCmd, err := NewExampleCommand()
	if err != nil {
		fmt.Printf("Error creating another example command: %v\n", err)
		os.Exit(1)
	}

	// Build the Cobra command with short help for the logging layer
	anotherCobraCmd, err := cli.BuildCobraCommandFromCommand(
		anotherExampleCmd,
		cli.WithCobraShortHelpLayers("logging"),
	)
	if err != nil {
		fmt.Printf("Error building another cobra command: %v\n", err)
		os.Exit(1)
	}

	// Give it a different name
	anotherCobraCmd.Use = "another-example"
	rootCmd.AddCommand(anotherCobraCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}

	/*
		Examples:

		Basic usage:
		$ go run examples/logging_layer_example.go example --log-level debug

		Log to a file:
		$ go run examples/logging_layer_example.go example --log-level debug --log-file example.log

		Use JSON format:
		$ go run examples/logging_layer_example.go example --log-format json

		Log to Logstash:
		$ go run examples/logging_layer_example.go example --logstash-enabled --logstash-host localhost --logstash-port 5044 --app-name "my-app" --environment development

		For a dedicated Logstash example, see logstash_example.go
	*/
}
