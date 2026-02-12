package sql

import (
	"context"
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

//go:embed "flags/sql-connection.yaml"
var connectionFlagsYaml []byte

type SqlConnectionParameterLayer struct {
	schema.SectionImpl `yaml:",inline"`
}

const SqlConnectionSlug = "sql-connection"

type SqlConnectionSettings struct {
	Host       string `glazed:"host"`
	Port       int    `glazed:"port"`
	Database   string `glazed:"database"`
	User       string `glazed:"user"`
	Password   string `glazed:"password"` // #nosec G117 -- Password is a required connection setting.
	Schema     string `glazed:"schema"`
	DbType     string `glazed:"db-type"`
	Repository string `glazed:"repository"`
	Dsn        string `glazed:"dsn"`
	Driver     string `glazed:"driver"`
	SSLDisable bool   `glazed:"ssl-disable"`
}

func NewSqlConnectionParameterLayer(
	options ...schema.SectionOption,
) (*SqlConnectionParameterLayer, error) {
	layer, err := schema.NewSectionFromYAML(connectionFlagsYaml, options...)
	if err != nil {
		return nil, err
	}
	ret := &SqlConnectionParameterLayer{}
	ret.SectionImpl = *layer

	return ret, nil
}

//go:embed "flags/dbt.yaml"
var dbtFlagsYaml []byte

type DbtParameterLayer struct {
	schema.SectionImpl `yaml:",inline"`
}

const DbtSlug = "dbt"

type DbtSettings struct {
	DbtProfilesPath string `glazed:"dbt-profiles-path"`
	UseDbtProfiles  bool   `glazed:"use-dbt-profiles"`
	DbtProfile      string `glazed:"dbt-profile"`
}

func NewDbtParameterLayer(
	options ...schema.SectionOption,
) (*DbtParameterLayer, error) {
	ret, err := schema.NewSectionFromYAML(dbtFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize dbt section")
	}
	return &DbtParameterLayer{
		SectionImpl: *ret,
	}, nil
}

type DBConnectionFactory func(ctx context.Context, parsedValues *values.Values) (*sqlx.DB, error)

func OpenDatabaseFromDefaultSqlConnectionLayer(
	ctx context.Context,
	parsedValues *values.Values,
) (*sqlx.DB, error) {
	return OpenDatabaseFromSqlConnectionLayer(ctx, parsedValues, SqlConnectionSlug, DbtSlug)
}

var _ DBConnectionFactory = OpenDatabaseFromDefaultSqlConnectionLayer

func OpenDatabaseFromSqlConnectionLayer(
	ctx context.Context,
	parsedValues *values.Values,
	sqlConnectionLayerName string,
	dbtLayerName string,
) (*sqlx.DB, error) {
	sqlConnectionLayer, ok := parsedValues.Get(sqlConnectionLayerName)
	if !ok {
		return nil, errors.New("No sql-connection section found")
	}
	dbtLayer, ok := parsedValues.Get(dbtLayerName)
	if !ok {
		return nil, errors.New("No dbt section found")
	}

	config, err2 := NewConfigFromParsedLayers(sqlConnectionLayer, dbtLayer)
	if err2 != nil {
		return nil, err2
	}
	return config.Connect(ctx)
}

func NewConfigFromRawParsedLayers(parsedValues *values.Values) (*DatabaseConfig, error) {
	sqlConnectionLayer, ok := parsedValues.Get(SqlConnectionSlug)
	if !ok {
		return nil, errors.New("No sql-connection section found")
	}
	dbtLayer, ok := parsedValues.Get(DbtSlug)
	if !ok {
		return nil, errors.New("No dbt section found")
	}

	config, err := NewConfigFromParsedLayers(sqlConnectionLayer, dbtLayer)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetConnectionStringFromParsedLayers extracts a connection string from parsed values.
// This is useful for tools like River that need a connection string directly
func GetConnectionStringFromParsedLayers(parsedValues *values.Values) (string, error) {
	config, err := NewConfigFromRawParsedLayers(parsedValues)
	if err != nil {
		return "", err
	}

	return config.GetConnectionString()
}

// GetConnectionStringFromSqlConnectionLayer extracts a connection string from specific section names
func GetConnectionStringFromSqlConnectionLayer(
	parsedValues *values.Values,
	sqlConnectionLayerName string,
	dbtLayerName string,
) (string, error) {
	log.Debug().Str("sqlConnectionLayerName", sqlConnectionLayerName).Str("dbtLayerName", dbtLayerName).Msg("Opening database from sql connection layer")
	sqlConnectionLayer, ok := parsedValues.Get(sqlConnectionLayerName)
	if !ok {
		return "", errors.New("No sql-connection section found")
	}
	dbtLayer, ok := parsedValues.Get(dbtLayerName)
	if !ok {
		return "", errors.New("No dbt section found")
	}

	config, err := NewConfigFromParsedLayers(sqlConnectionLayer, dbtLayer)
	if err != nil {
		return "", err
	}

	return config.GetConnectionString()
}
