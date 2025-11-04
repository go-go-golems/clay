package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Create a new root command
	rootCmd := &cobra.Command{
		Use:   "logstash-example",
		Short: "Example application showing Logstash integration",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Log some messages at different levels
			log.Debug().Msg("This is a debug message")
			log.Info().Msg("This is an info message")
			log.Warn().Msg("This is a warning message")
			log.Error().Msg("This is an error message")

			// Log a structured message with fields
			log.Info().
				Str("component", "example").
				Int("count", 42).
				Float64("value", 3.14).
				Bool("enabled", true).
				Time("timestamp", time.Now()).
				Msg("This is a structured log message")

			// Sleep for a moment to allow logs to be sent
			time.Sleep(500 * time.Millisecond)
		},
	}

	// Initialize logging flags on the root command (no Viper)
	err := pkg.InitGlazed("logstash-example", rootCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing configuration: %v\n", err)
		os.Exit(1)
	}

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

/*
To run this example with Logstash enabled:

$ go run examples/logstash_example.go --logstash-enabled --logstash-host localhost --logstash-port 5044 --app-name "my-app" --environment development --log-level debug

If you don't have a Logstash server available, you can use netcat to test the connection:

$ nc -lk 5044

This will listen for connections on port 5044 and print received data to the console.
*/
