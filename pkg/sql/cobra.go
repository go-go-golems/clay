package sql

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/sqleton/pkg/flags"
	"github.com/spf13/cobra"
)

func BuildCobraCommandWithSqletonMiddlewares(
	cmd cmds.Command,
	options ...cli.CobraParserOption,
) (*cobra.Command, error) {
	options_ := append([]cli.CobraParserOption{
		cli.WithCobraMiddlewaresFunc(GetCobraCommandSqletonMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug, DbtSlug, SqlConnectionSlug, flags.SqlHelpersSlug),
		cli.WithCreateCommandSettingsLayer(),
		cli.WithProfileSettingsLayer(),
	}, options...)

	return cli.BuildCobraCommandFromCommand(cmd, options_...)
}

func GetCobraCommandSqletonMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {

	// Start with cobra-specific middlewares
	middlewares_ := []middlewares.Middleware{
		middlewares.ParseFromCobraCommand(cmd,
			parameters.WithParseStepSource("cobra"),
		),
		middlewares.GatherArguments(args,
			parameters.WithParseStepSource("arguments"),
		),
	}

	// Get the common sqleton middlewares
	additionalMiddlewares, err := GetSqletonMiddlewares(parsedCommandLayers)
	if err != nil {
		return nil, err
	}
	middlewares_ = append(middlewares_, additionalMiddlewares...)

	return middlewares_, nil
}

// GetSqletonMiddlewares returns the common middleware chain used by sqleton commands
func GetSqletonMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
) ([]middlewares.Middleware, error) {
	commandSettings := &cli.CommandSettings{}
	err := parsedCommandLayers.InitializeStruct(cli.CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}
	middlewares_ := []middlewares.Middleware{}

	if commandSettings.LoadParametersFromFile != "" {
		middlewares_ = append(middlewares_,
			middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
	}

	profileSettings := &cli.ProfileSettings{}
	err = parsedCommandLayers.InitializeStruct(cli.ProfileSettingsSlug, profileSettings)
	if err != nil {
		return nil, err
	}

	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	defaultProfileFile := fmt.Sprintf("%s/sqleton/profiles.yaml", xdgConfigPath)
	if profileSettings.ProfileFile == "" {
		profileSettings.ProfileFile = defaultProfileFile
	}
	if profileSettings.Profile == "" {
		profileSettings.Profile = "default"
	}
	middlewares_ = append(middlewares_,
		middlewares.GatherFlagsFromProfiles(
			defaultProfileFile,
			profileSettings.ProfileFile,
			profileSettings.Profile,
			parameters.WithParseStepSource("profiles"),
			parameters.WithParseStepMetadata(map[string]interface{}{
				"profileFile": profileSettings.ProfileFile,
				"profile":     profileSettings.Profile,
			}),
		),
	)

	middlewares_ = append(middlewares_,
		middlewares.WrapWithWhitelistedLayers(
			[]string{
				DbtSlug,
				SqlConnectionSlug,
			},
			middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper")),
		),
		middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	)

	return middlewares_, nil
}
