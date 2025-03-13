package pkg

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	WithCaller bool
	Level      string
	LogFormat  string
	LogFile    string
}

func InitLoggerWithConfig(config *LogConfig) error {
	if config.WithCaller {
		log.Logger = log.With().Caller().Logger()
	}

	// Set timestamp format to include milliseconds
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// default is json
	var logWriter io.Writer
	if config.LogFormat == "text" {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	if config.LogFile != "" {
		fileLogger := &lumberjack.Logger{
			Filename:   config.LogFile,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28,    //days
			Compress:   false, // disabled by default
		}
		var writer io.Writer
		writer = fileLogger
		if config.LogFormat == "text" {
			log.Info().Str("file", config.LogFile).Msg("Logging to file")
			writer = zerolog.ConsoleWriter{
				NoColor:    true,
				Out:        fileLogger,
				TimeFormat: time.RFC3339Nano,
			}
		}
		// TODO(manuel, 2024-07-05) We used to support logging to file *and* stderr, but disabling that for now
		// because it makes logging in UI apps tricky.
		// logWriter = io.MultiWriter(logWriter, writer)
		logWriter = writer
	}

	log.Logger = log.Output(logWriter)

	switch config.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	}

	log.Logger.Debug().Str("format", config.LogFormat).
		Str("level", config.Level).
		Str("file", config.LogFile).
		Msg("Logger initialized")

	return nil
}

func InitLogger() error {
	logLevel := viper.GetString("log-level")
	logLevel = strings.ToLower(logLevel)
	verbose := viper.GetBool("verbose")
	if verbose && logLevel != "trace" {
		logLevel = "debug"
	}

	return InitLoggerWithConfig(&LogConfig{
		Level:      logLevel,
		LogFile:    viper.GetString("log-file"),
		LogFormat:  viper.GetString("log-format"),
		WithCaller: viper.GetBool("with-caller"),
	})
}

func InitViperWithAppName(appName string, configFile string) error {
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

func InitViper(appName string, rootCmd *cobra.Command) error {
	rootCmd.PersistentFlags().Bool("with-caller", false, "Log caller")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (json, text)")
	rootCmd.PersistentFlags().String("log-file", "", "Log file (default: stderr)")

	rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")

	rootCmd.PersistentFlags().String("config", "",
		fmt.Sprintf("Path to config file (default ~/.%s/config.yml)", appName))

	// parse the flags one time just to catch --config
	configFile := ""
	for idx, arg := range os.Args {
		if arg == "--config" {
			if len(os.Args) > idx+1 {
				configFile = os.Args[idx+1]
			}
		}
	}

	err := InitViperWithAppName(appName, configFile)
	if err != nil {
		return err
	}

	// Bind the variables to the command-line flags
	err = viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		return err
	}

	return nil
}

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
