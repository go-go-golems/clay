package sql

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

		// Get the actual source values from DBT profile
		source, err := c.GetSource()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to get source from dbt profile")
			return
		}

		log.Debug().
			Str("host", source.Hostname).
			Str("database", source.Database).
			Str("user", source.Username).
			Int("port", source.Port).
			Str("schema", source.Schema).
			Str("type", source.Type).
			Msg("Using connection string from dbt profile")
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

	// Normalize driver/type names
	switch strings.ToLower(source.Type) {
	case "sqlite":
		source.Type = "sqlite3"
	case "postgres", "postgresql", "pg":
		source.Type = "pgx"
	case "mariadb":
		source.Type = "mysql"
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
	// Normalize driver based on provided value or DSN
	if c.DSN != "" {
		// Infer driver from DSN scheme if not explicitly provided
		if c.Driver == "" {
			lower := strings.ToLower(c.DSN)
			switch {
			case strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://"):
				c.Driver = "pgx"
			case strings.HasPrefix(lower, "mysql://") || strings.HasPrefix(lower, "mariadb://"):
				// For DSN with URL scheme, the mysql std driver expects a different format,
				// but many connectors accept it; we still set driver appropriately.
				c.Driver = "mysql"
			case strings.HasPrefix(lower, "sqlite://") || strings.HasPrefix(lower, "sqlite3://"):
				c.Driver = "sqlite3"
			}
		}
		// Canonicalize driver aliases
		switch strings.ToLower(c.Driver) {
		case "postgres", "postgresql", "pg":
			c.Driver = "pgx"
		case "sqlite":
			c.Driver = "sqlite3"
		case "mariadb":
			c.Driver = "mysql"
		}

		// Enforce driver-level timeout for unreachable pgx endpoints
		if c.Driver == "pgx" && !strings.Contains(c.DSN, "connect_timeout") {
			if strings.HasPrefix(c.DSN, "postgres://") || strings.HasPrefix(c.DSN, "postgresql://") {
				// Append as URL query parameter
				if u, err := url.Parse(c.DSN); err == nil {
					q := u.Query()
					if q.Get("connect_timeout") == "" {
						q.Set("connect_timeout", "5")
						u.RawQuery = q.Encode()
						c.DSN = u.String()
					}
				} else {
					// Fallback: append as key/value
					c.DSN = c.DSN + " connect_timeout=5"
				}
			} else {
				// Key/value style DSN
				c.DSN = c.DSN + " connect_timeout=5"
			}
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

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, errors.Wrap(err, "failed to ping database")
	}

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
