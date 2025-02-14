---
Title: SQL Commands in Clay
Slug: sql-commands
Short: Learn how to define and execute SQL queries using Clay's powerful templating and parameter system
Topics:
  - sql
  - database
  - templating
  - parameters
Commands:
  - BuildCobraCommandWithSqletonMiddlewares
Flags:
  - sql-connection
  - dbt
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Overview

Clay's SQL command system (`github.com/go-go-golems/clay/pkg/sql`) provides a powerful and flexible way to define and execute SQL queries using YAML configuration files. It combines the power of Go templates with SQL, allowing for dynamic query generation with proper parameter handling and SQL injection prevention.

To use the SQL functionality in your project, import the package:

```go
import "github.com/go-go-golems/clay/pkg/sql"
```

## Table of Contents

1. Getting Started
2. Command Structure
3. SQL Parameter Layers
4. Query Definition and Templating
5. Advanced Features
6. Best Practices
7. Error Handling
8. Integration with Cobra

## Getting Started

SQL commands in Clay are defined in YAML files. These files describe SQL queries, their parameters, and how to execute them with proper templating and safety features.

### Prerequisites

- Clay installed
- Basic understanding of YAML
- Basic SQL knowledge
- Database connection details

### Directory Setup

Create a directory for your SQL queries:

```bash
mkdir -p queries/
cd queries/
```

## Command Structure

A SQL command YAML file has the following structure:

```yaml
# Metadata
name: query-name           # Required: Command name (use lowercase and hyphens)
short: Short description   # Required: One-line description
long: |                    # Optional: Detailed multi-line description
  Detailed description that can
  span multiple lines

# Parameter Definition
flags:                     # Optional: Query parameters
  - name: parameter_name   # Required: Parameter name
    type: string          # Required: Parameter type
    help: Description     # Required: Parameter description
    required: true        # Optional: Whether the parameter is required
    default: value        # Optional: Default value
    choices: [a, b, c]    # Optional: For choice/choiceList types

# Query Definition
query: |                  # Required: The SQL query template
  SELECT column1, column2
  FROM table
  WHERE condition = {{ .parameter_name }}

# Subqueries Definition
subqueries:              # Optional: Named subqueries that can be referenced
  user_types: |
    SELECT DISTINCT user_type 
    FROM users
```

### Parameter Types

The following parameter types are supported:

- **Basic Types**
  - `string`: Text values
  - `int`: Integer numbers
  - `float`: Floating point numbers
  - `bool`: True/false values
  - `date`: Date values
  - `datetime`: Date and time values
- **List Types**
  - `stringList`: List of strings
  - `intList`: List of integers
  - `floatList`: List of floating point numbers
- **Choice Types**
  - `choice`: Single selection from predefined options
  - `choiceList`: Multiple selections from predefined options

## SQL Parameter Layers

The SQL implementation uses several parameter layers to organize and manage different aspects of SQL queries:

### Connection Layer
The SQL connection layer manages database connection parameters through the `SqlConnectionSettings` struct:

```go
import "github.com/go-go-golems/clay/pkg/sql"

type sql.SqlConnectionSettings struct {
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
```

Create a new SQL connection layer:

```go
layer, err := sql.NewSqlConnectionParameterLayer()
if err != nil {
    // Handle error
}
```

### DBT Layer
The DBT layer manages dbt-specific configurations through the `DbtSettings` struct:

```go
type sql.DbtSettings struct {
    DbtProfilesPath string `glazed.parameter:"dbt-profiles-path"`
    UseDbtProfiles  bool   `glazed.parameter:"use-dbt-profiles"`
    DbtProfile      string `glazed.parameter:"dbt-profile"`
}
```

Create a new DBT layer:

```go
layer, err := sql.NewDbtParameterLayer()
if err != nil {
    // Handle error
}
```

### Opening Database Connections

Clay provides convenient functions to open database connections using the parsed layers:

1. Using the default layers:
```go
db, err := sql.OpenDatabaseFromDefaultSqlConnectionLayer(parsedLayers)
if err != nil {
    // Handle error
}
```

2. Using custom layer names:
```go
db, err := sql.OpenDatabaseFromSqlConnectionLayer(
    parsedLayers,
    "custom-sql-connection-layer",
    "custom-dbt-layer",
)
if err != nil {
    // Handle error
}
```

You can also create a custom database connection factory:

```go
type sql.DBConnectionFactory func(parsedLayers *layers.ParsedLayers) (*sqlx.DB, error)

// Example usage:
var connectionFactory sql.DBConnectionFactory = sql.OpenDatabaseFromDefaultSqlConnectionLayer
db, err := connectionFactory(parsedLayers)
```

## Query Definition and Templating

SQL queries in Clay use Go's template language with additional SQL-specific functions for safe value handling.

### Template Functions

Clay provides several template functions for safe SQL value handling:

#### String Handling
```sql
{{ sqlString value }}      -- 'value'
{{ sqlEscape value }}      -- Escapes quotes
{{ sqlStringLike value }} -- '%value%'
{{ sqlStringIn list }}    -- 'value1','value2'
```

#### Date Handling
```sql
{{ sqlDate value }}        -- '2023-01-01'
{{ sqlDateTime value }}    -- '2023-01-01T12:00:00'
{{ sqliteDate value }}    -- SQLite format
{{ sqliteDateTime value }} -- SQLite format
```

#### List Handling
```sql
{{ sqlIn values }}        -- value1,value2,value3
{{ sqlIntIn values }}     -- 1,2,3
```

### Control Flow

Use Go template syntax for conditional queries:

```sql
SELECT * 
FROM posts
WHERE 1=1
{{ if .category }}
  AND category = {{ .category | sqlString }}
{{ end }}
{{ if .tags }}
  AND tags && {{ .tags | sqlStringIn }}
{{ end }}
ORDER BY 
  {{ .sort_by | default "created_at" }} 
  {{ .sort_order | default "DESC" }}
```

## Advanced Features

### Subqueries
Define and use subqueries for complex operations:

```yaml
subqueries:
  active_users: |
    SELECT user_id 
    FROM user_status 
    WHERE status = 'active'

query: |
  SELECT * 
  FROM users
  WHERE id IN ({{ sqlColumn (subQuery "active_users") }})
```

### Dynamic Columns
Use template functions to generate dynamic column lists:

```yaml
query: |
  SELECT 
    {{ range $col := sqlColumn "SELECT column_name FROM information_schema.columns" }}
    {{ $col }},
    {{ end }}
  FROM table
```

## Examples

### Basic Query
```yaml
name: get-user
short: Get user by ID
flags:
  - name: user_id
    type: int
    help: User ID to fetch
    required: true
query: |
  SELECT id, username, email
  FROM users
  WHERE id = {{ .user_id }}
```

### Complex Query with Multiple Parameters
```yaml
name: search-orders
short: Search orders with filters
flags:
  - name: start_date
    type: date
    help: Start date for order search
    required: true
  - name: end_date
    type: date
    help: End date for order search
    required: true
  - name: status
    type: stringList
    help: Order status filter
    default: ["pending", "processing"]
  - name: min_amount
    type: float
    help: Minimum order amount
query: |
  SELECT 
    o.id,
    o.created_at,
    o.status,
    o.total_amount,
    c.name as customer_name
  FROM orders o
  JOIN customers c ON o.customer_id = c.id
  WHERE 
    o.created_at BETWEEN {{ .start_date | sqlDate }} AND {{ .end_date | sqlDate }}
    {{ if .status }}
    AND o.status IN ({{ .status | sqlStringIn }})
    {{ end }}
    {{ if .min_amount }}
    AND o.total_amount >= {{ .min_amount }}
    {{ end }}
  ORDER BY o.created_at DESC
```

## Best Practices

1. **Parameter Validation**
   - Always specify parameter types
   - Use appropriate default values
   - Mark required parameters as `required: true`

2. **Security**
   - Always use template functions for parameter interpolation
   - Never concatenate raw strings into queries
   - Use `sqlEscape` for free-form text

3. **Query Organization**
   - Group related queries in directories
   - Use clear, descriptive names
   - Include helpful descriptions in `short` and `long` fields

4. **Template Usage**
   - Use conditional blocks for optional clauses
   - Leverage subqueries for reusable components
   - Keep complex logic in Go code rather than templates

## Error Handling

The SQL package provides detailed error messages for common issues:
- Template parsing errors
- SQL syntax errors
- Parameter type mismatches
- Missing required parameters
- Database connection issues

## Integration with Cobra

The SQL package integrates with Cobra for CLI applications:

```go
package main

import (
    "github.com/go-go-golems/clay/pkg/sql"
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
)

cobraCmd, err := sql.BuildCobraCommandWithSqletonMiddlewares(
    sqlCmd,
    cli.WithCobraShortHelpLayers(
        layers.DefaultSlug,
        sql.DbtSlug,
        sql.SqlConnectionSlug,
    ),
)
```

This integration provides:
- Automatic flag parsing
- Help text generation
- Parameter validation
- Environment variable support
- Configuration file loading

## Conclusion

Clay's SQL command system provides a robust and flexible way to define and execute SQL queries. By combining YAML configuration, Go templates, and proper parameter handling, it enables developers to create safe and maintainable database queries while maintaining full flexibility and power.
