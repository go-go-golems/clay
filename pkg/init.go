package pkg

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Deprecated: config loading is Glazed territory now. Use Glazed's
// cli.CobraParserConfig with AppName for environment variables and a
// ConfigPlanBuilder / config sources for files instead of initializing Viper
// through Clay. For logging, add Glazed's logging section directly with
// logging.AddLoggingSectionToRootCommand(rootCmd, appName) and initialize it in
// PersistentPreRunE with logging.InitLoggerFromCobra(cmd).
func InitViperWithAppName(appName string, configFile string) error {
	log.Warn().Msg("clay.InitViperWithAppName is deprecated; use Glazed CobraParserConfig/config sources and Glazed logging setup directly")
	viper.SetEnvPrefix(appName)

	if configFile != "" {
		viper.SetConfigFile(configFile)
		viper.SetConfigType("yaml")
	} else {
		viper.SetConfigType("yaml")
		viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", appName))
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", appName))

		xdgConfigPath, err := os.UserConfigDir()
		if err == nil {
			viper.AddConfigPath(fmt.Sprintf("%s/%s", xdgConfigPath, appName))
		}
	}

	// Read the configuration file into Viper
	err := viper.ReadInConfig()
	// if the file does not exist, continue normally
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found; ignore error
	} else if err != nil {
		// Config file was found but another error was produced
		return err
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return nil
}

// Deprecated: config loading is Glazed territory now. Avoid Viper instances for
// command configuration; use Glazed's cli.CobraParserConfig with AppName for
// environment variables and a ConfigPlanBuilder / config sources for files.
// Keep Clay imports only for Clay-specific packages such as pkg/sql.
func InitViperInstanceWithAppName(appName string, configFile string) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix(appName)

	if configFile != "" {
		v.SetConfigFile(configFile)
		v.SetConfigType("yaml")
	} else {
		v.SetConfigType("yaml")
		v.AddConfigPath(fmt.Sprintf("$HOME/.%s", appName))
		v.AddConfigPath(fmt.Sprintf("/etc/%s", appName))

		xdgConfigPath, err := os.UserConfigDir()
		if err == nil {
			v.AddConfigPath(fmt.Sprintf("%s/%s", xdgConfigPath, appName))
		}
	}

	// Read the configuration file into Viper
	err := v.ReadInConfig()
	// if the file does not exist, continue normally
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found; ignore error
	} else if err != nil {
		// Config file was found but another error was produced
		return nil, err
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	return v, nil
}

// InitGlazed adds the logging section to the root command without wiring Viper.
//
// Deprecated: logging and command config setup are Glazed territory. Replace
//
//	clay.InitGlazed(appName, rootCmd)
//
// with
//
//	logging.AddLoggingSectionToRootCommand(rootCmd, appName)
//
// from github.com/go-go-golems/glazed/pkg/cmds/logging, and keep initializing
// logging in PersistentPreRunE with logging.InitLoggerFromCobra(cmd). Configure
// environment variables and config files through Glazed's cli.CobraParserConfig
// (AppName plus ConfigPlanBuilder / config sources), not through Clay/Viper.
func InitGlazed(appName string, rootCmd *cobra.Command) error {
	return logging.AddLoggingSectionToRootCommand(rootCmd, appName)
}
