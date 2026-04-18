# Changelog

## 2026-04-18

- Initial workspace created


## 2026-04-18

Switched clay's DuckDB binding from github.com/marcboeker/go-duckdb to github.com/duckdb/duckdb-go/v2, removed the legacy module from go.mod/go.sum, and verified the workspace with go build ./... plus go-minitrace make build to ensure there are no DuckDB symbol conflicts.


## 2026-04-18

Ticket closed

