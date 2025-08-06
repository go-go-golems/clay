package sql

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

type DatabaseConfig struct {
	Host            string `glazed.parameter:"host"`
	Database        string `glazed.parameter:"database"`
	User            string `glazed.parameter:"user"`
	Password        string `glazed.parameter:"password"`
	Port            int    `glazed.parameter:"port"`
	Schema          string `glazed.parameter:"schema"`
	Type            string `glazed.parameter:"db-type"`
	DSN             string `glazed.parameter:"dsn"`
	Driver          string `glazed.parameter:"driver"`
	DbtProfilesPath string `glazed.parameter:"dbt-profiles-path"`
	DbtProfile      string `glazed.parameter:"dbt-profile"`
	UseDbtProfiles  bool   `glazed.parameter:"use-dbt-profiles"`
}

// LogVerbose just outputs information about the database config to the
// debug logging level.
func (c *DatabaseConfig) LogVerbose() {
	if c.UseDbtProfiles {
		log.Debug().
			Str("dbt-profiles-path", c.DbtProfilesPath).
			Str("dbt-profile", c.DbtProfile).
			Msg("Using dbt profiles")

		log.Debug().
			Str("host", c.Host).
			Str("database", c.Database).
			Str("user", c.User).
			Int("port", c.Port).
			Str("schema", c.Schema).
			Str("type", c.Type).
			Msg("Using connection string")
	} else if c.DSN != "" {
		log.Debug().
			Str("dsn", c.DSN).
			Str("driver", c.Driver).
			Msg("Using DSN")
	} else {
		log.Debug().
			Str("host", c.Host).
			Str("database", c.Database).
			Str("user", c.User).
			Int("port", c.Port).
			Str("schema", c.Schema).
			Str("type", c.Type).
			Msg("Using connection string")
	}
}

func (c *DatabaseConfig) ToString() string {
	if c.UseDbtProfiles {
		s, err := c.GetSource()
		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		sourceString := fmt.Sprintf("%s@%s:%d/%s", s.Username, s.Hostname, s.Port, s.Database)

		if c.DbtProfilesPath != "" {
			return fmt.Sprintf("dbt-profiles-path: %s, dbt-profile: %s, %s", c.DbtProfilesPath, c.DbtProfile, sourceString)
		} else {
			return fmt.Sprintf("dbt-profile: %s, %s", c.DbtProfile, sourceString)
		}
	} else if c.DSN != "" {
		return fmt.Sprintf("dsn: %s, driver: %s", c.DSN, c.Driver)
	} else {
		return fmt.Sprintf("%s@%s:%d/%s", c.User, c.Host, c.Port, c.Database)
	}
}

func (c *DatabaseConfig) GetSource() (*Source, error) {
	var source *Source

	if c.UseDbtProfiles {
		if c.DbtProfile == "" {
			return nil, errors.Errorf("No dbt profile specified")
		}

		sources, err := ParseDbtProfiles(c.DbtProfilesPath)
		if err != nil {
			return nil, err
		}

		for _, s := range sources {
			log.Trace().
				Str("Profile", c.DbtProfile).
				Str("name", s.Name).
				Msg("Checking source")
			if s.Name == c.DbtProfile {
				source = s
				break
			}
		}

		if source == nil {
			return nil, errors.Errorf("Source %s not found", c.DbtProfile)
		}
	} else {
		source = &Source{
			Type:     c.Type,
			Hostname: c.Host,
			Port:     c.Port,
			Username: c.User,
			Password: c.Password,
			Database: c.Database,
			Schema:   c.Schema,
		}
	}

	if source.Type == "sqlite" {
		source.Type = "sqlite3"
	}

	return source, nil
}

// GetConnectionString returns the connection string for this database config
func (c *DatabaseConfig) GetConnectionString() (string, error) {
	if c.DSN != "" {
		return c.DSN, nil
	}

	s, err := c.GetSource()
	if err != nil {
		return "", err
	}

	return s.ToConnectionString(), nil
}

func (c *DatabaseConfig) Connect(ctx context.Context) (*sqlx.DB, error) {
	// DEBUG: Monitor context cancellation
	go func() {
		<-ctx.Done()
		fmt.Printf("[CONFIG DEBUG] Context cancelled in Connect() - signal received: %v\n", ctx.Err())
	}()
	// enforce driver-level timeout for unreachable endpoints
	if c.Driver == "pgx" {
		// append connect_timeout if missing
		if !strings.Contains(c.DSN, "connect_timeout") {
			c.DSN = c.DSN + " connect_timeout=5"
		}
	}
	c.LogVerbose()

	var dbType string
	connectionString, err := c.GetConnectionString()
	if err != nil {
		return nil, err
	}

	if c.DSN != "" {
		dbType = c.Driver
	} else {
		s, err := c.GetSource()
		if err != nil {
			return nil, err
		}
		dbType = s.Type
	}

	log.Debug().Msg("Opening database connection")
	db, err := sqlx.Open(dbType, connectionString)
	if err != nil {
		return nil, err
	}
	log.Debug().Msg("Database connection established")
	// use context with timeout for ping
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// DEBUG: Monitor ping context cancellation
	go func() {
		<-pingCtx.Done()
		fmt.Printf("[CONFIG DEBUG] PingContext cancelled - reason: %v\n", pingCtx.Err())
	}()
	
	fmt.Printf("[CONFIG DEBUG] Starting PingContext call\n")
	if err := db.PingContext(pingCtx); err != nil {
		fmt.Printf("[CONFIG DEBUG] PingContext failed with error: %v\n", err)
		_ = db.Close()
		return nil, errors.Wrap(err, "failed to ping database")
	}
	fmt.Printf("[CONFIG DEBUG] PingContext succeeded\n")

	// TODO(2022-12-18, manuel): this is where we would add support for a ro connection
	// https://github.com/wesen/sqleton/issues/24

	return db, err
}

func NewConfigFromParsedLayers(parsedLayers ...*layers.ParsedLayer) (*DatabaseConfig, error) {
	config := &DatabaseConfig{}
	for _, layer := range parsedLayers {
		err := layer.Parameters.InitializeStruct(config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
