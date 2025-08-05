# Fix DB Ping Context Timeout Issue

We observed that `OpenDatabaseFromSqlConnectionLayer` hangs indefinitely when the Postgres endpoint is unreachable. The root cause is clay's call to `sqlx.Connect`, which uses `db.DB.Ping()` by default.

## Solution

Wrap the ping in a `context.Context` with a timeout before proceeding:

```go
import (
    "context"
    "fmt"
    "time"
)

// after obtaining `db`:
pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
if err := db.DB.PingContext(pingCtx); err != nil {
    db.Close()
    return nil, fmt.Errorf("failed to ping database: %w", err)
}
```

This ensures the call fails fast after the configured deadline. Adjust `5*time.Second` as needed.
