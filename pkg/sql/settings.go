package sql

import (
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

//go:embed "flags/sql-connection.yaml"
var connectionFlagsYaml []byte

type SqlConnectionParameterLayer struct {
	layers.ParameterLayerImpl `yaml:",inline"`
}

const SqlConnectionSlug = "sql-connection"

type SqlConnectionSettings struct {
	Host       string `glazed.parameter:"host"`
	Port       int    `glazed.parameter:"port"`
	Database   string `glazed.parameter:"database"`
	User       string `glazed.parameter:"user"`
	Password   string `glazed.parameter:"password"`
	Schema     string `glazed.parameter:"schema"`
	DbType     string `glazed.parameter:"db-type"`
	Repository string `glazed.parameter:"repository"`
	Dsn        string `glazed.parameter:"dsn"`
	Driver     string `glazed.parameter:"driver"`
}

func NewSqlConnectionParameterLayer(
	options ...layers.ParameterLayerOptions,
) (*SqlConnectionParameterLayer, error) {
	layer, err := layers.NewParameterLayerFromYAML(connectionFlagsYaml, options...)
	if err != nil {
		return nil, err
	}
	ret := &SqlConnectionParameterLayer{}
	ret.ParameterLayerImpl = *layer

	return ret, nil
}

//go:embed "flags/dbt.yaml"
var dbtFlagsYaml []byte

type DbtParameterLayer struct {
	layers.ParameterLayerImpl `yaml:",inline"`
}

const DbtSlug = "dbt"

type DbtSettings struct {
	DbtProfilesPath string `glazed.parameter:"dbt-profiles-path"`
	UseDbtProfiles  bool   `glazed.parameter:"use-dbt-profiles"`
	DbtProfile      string `glazed.parameter:"dbt-profile"`
}

func NewDbtParameterLayer(
	options ...layers.ParameterLayerOptions,
) (*DbtParameterLayer, error) {
	ret, err := layers.NewParameterLayerFromYAML(dbtFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize dbt parameter layer")
	}
	return &DbtParameterLayer{
		ParameterLayerImpl: *ret,
	}, nil
}

type DBConnectionFactory func(parsedLayers *layers.ParsedLayers) (*sqlx.DB, error)

func OpenDatabaseFromDefaultSqlConnectionLayer(
	parsedLayers *layers.ParsedLayers,
) (*sqlx.DB, error) {
	return OpenDatabaseFromSqlConnectionLayer(parsedLayers, SqlConnectionSlug, DbtSlug)
}

var _ DBConnectionFactory = OpenDatabaseFromDefaultSqlConnectionLayer

func OpenDatabaseFromSqlConnectionLayer(
	parsedLayers *layers.ParsedLayers,
	sqlConnectionLayerName string,
	dbtLayerName string,
) (*sqlx.DB, error) {
	sqlConnectionLayer, ok := parsedLayers.Get(sqlConnectionLayerName)
	if !ok {
		return nil, errors.New("No sql-connection layer found")
	}
	dbtLayer, ok := parsedLayers.Get(dbtLayerName)
	if !ok {
		return nil, errors.New("No dbt layer found")
	}

	config, err2 := NewConfigFromParsedLayers(sqlConnectionLayer, dbtLayer)
	if err2 != nil {
		return nil, err2
	}
	return config.Connect()
}

func NewConfigFromRawParsedLayers(parsedLayers *layers.ParsedLayers) (*DatabaseConfig, error) {
	sqlConnectionLayer, ok := parsedLayers.Get(SqlConnectionSlug)
	if !ok {
		return nil, errors.New("No sql-connection layer found")
	}
	dbtLayer, ok := parsedLayers.Get(DbtSlug)
	if !ok {
		return nil, errors.New("No dbt layer found")
	}

	config, err := NewConfigFromParsedLayers(sqlConnectionLayer, dbtLayer)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetConnectionStringFromParsedLayers extracts a connection string from parsed layers
// This is useful for tools like River that need a connection string directly
func GetConnectionStringFromParsedLayers(parsedLayers *layers.ParsedLayers) (string, error) {
	config, err := NewConfigFromRawParsedLayers(parsedLayers)
	if err != nil {
		return "", err
	}

	return config.GetConnectionString()
}

// GetConnectionStringFromSqlConnectionLayer extracts a connection string from specific layer names
func GetConnectionStringFromSqlConnectionLayer(
	parsedLayers *layers.ParsedLayers,
	sqlConnectionLayerName string,
	dbtLayerName string,
) (string, error) {
	sqlConnectionLayer, ok := parsedLayers.Get(sqlConnectionLayerName)
	if !ok {
		return "", errors.New("No sql-connection layer found")
	}
	dbtLayer, ok := parsedLayers.Get(dbtLayerName)
	if !ok {
		return "", errors.New("No dbt layer found")
	}

	config, err := NewConfigFromParsedLayers(sqlConnectionLayer, dbtLayer)
	if err != nil {
		return "", err
	}

	return config.GetConnectionString()
}
