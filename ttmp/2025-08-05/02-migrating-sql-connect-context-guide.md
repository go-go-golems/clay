# Database Connection Context Migration Guide

## Breaking Changes

Database opening functions now require a `context.Context` parameter to enable proper timeout handling and cancellation.

## Updated Function Signatures

```go
// Before
func (c *DatabaseConfig) Connect() (*sqlx.DB, error)
func OpenDatabaseFromDefaultSqlConnectionLayer(parsedLayers *layers.ParsedLayers) (*sqlx.DB, error)
type DBConnectionFactory func(parsedLayers *layers.ParsedLayers) (*sqlx.DB, error)

// After  
func (c *DatabaseConfig) Connect(ctx context.Context) (*sqlx.DB, error)
func OpenDatabaseFromDefaultSqlConnectionLayer(ctx context.Context, parsedLayers *layers.ParsedLayers) (*sqlx.DB, error)
type DBConnectionFactory func(ctx context.Context, parsedLayers *layers.ParsedLayers) (*sqlx.DB, error)
```

## How to Update Your Code

### Direct config.Connect() calls
```go
// Before
db, err := config.Connect()

// After
db, err := config.Connect(ctx)
// or in cobra commands:
db, err := config.Connect(cmd.Context())
```

### DBConnectionFactory usage
```go
// Before
db, err := factory(parsedLayers)

// After  
db, err := factory(ctx, parsedLayers)
```

### Custom DBConnectionFactory implementations
```go
// Before
func myFactory(parsedLayers *layers.ParsedLayers) (*sqlx.DB, error) {
    // implementation
}

// After
func myFactory(ctx context.Context, parsedLayers *layers.ParsedLayers) (*sqlx.DB, error) {
    // implementation - now can use ctx for timeouts
}
```

## Benefits

- Database connections now fail fast when endpoints are unreachable
- Proper context cancellation and timeout propagation
- Fixes hanging connections to unreachable Postgres endpoints
