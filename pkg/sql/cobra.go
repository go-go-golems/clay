package sql

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/sqleton/pkg/flags"
	"github.com/spf13/cobra"
)

func BuildCobraCommandWithSqletonMiddlewares(
	cmd cmds.Command,
	options ...cli.CobraOption,
) (*cobra.Command, error) {
	options_ := append([]cli.CobraOption{
		cli.WithCobraMiddlewaresFunc(GetCobraCommandSqletonMiddlewares),
		cli.WithCobraShortHelpSections(schema.DefaultSlug, DbtSlug, SqlConnectionSlug, flags.SqlHelpersSlug),
		cli.WithCreateCommandSettingsSection(),
		cli.WithProfileSettingsSection(),
	}, options...)

	return cli.BuildCobraCommandFromCommand(cmd, options_...)
}

func GetCobraCommandSqletonMiddlewares(
	parsedCommandValues *values.Values,
	cmd *cobra.Command,
	args []string,
) ([]sources.Middleware, error) {

	// Start with cobra-specific middlewares
	middlewares_ := []sources.Middleware{
		sources.FromCobra(cmd,
			fields.WithSource("cobra"),
		),
		sources.FromArgs(args,
			fields.WithSource("arguments"),
		),
	}

	// Get the common sqleton middlewares
	additionalMiddlewares, err := GetSqletonMiddlewares(parsedCommandValues)
	if err != nil {
		return nil, err
	}
	middlewares_ = append(middlewares_, additionalMiddlewares...)

	return middlewares_, nil
}

// GetSqletonMiddlewares returns the common middleware chain used by sqleton commands
func GetSqletonMiddlewares(
	parsedCommandValues *values.Values,
) ([]sources.Middleware, error) {
	commandSettings := &cli.CommandSettings{}
	err := parsedCommandValues.DecodeSectionInto(cli.CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}
	middlewares_ := []sources.Middleware{}

	profileSettings := &cli.ProfileSettings{}
	err = parsedCommandValues.DecodeSectionInto(cli.ProfileSettingsSlug, profileSettings)
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
		sources.GatherFlagsFromProfiles(
			defaultProfileFile,
			profileSettings.ProfileFile,
			profileSettings.Profile,
			"default",
			fields.WithSource("profiles"),
			fields.WithMetadata(map[string]interface{}{
				"profileFile": profileSettings.ProfileFile,
				"profile":     profileSettings.Profile,
			}),
		),
	)

	middlewares_ = append(middlewares_,
		sources.WrapWithWhitelistedSections(
			[]string{
				DbtSlug,
				SqlConnectionSlug,
			},
			// Load from environment using SQLETON_* prefix
			sources.FromEnv(strings.ToUpper("sqleton"),
				fields.WithSource("env"),
			),
		),
		sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)

	return middlewares_, nil
}
