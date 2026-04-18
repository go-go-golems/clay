---
Title: DuckDB migration analysis and risk assessment
Ticket: CLAY-DUCKDB-001
Status: active
Topics:
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-minitrace/Makefile
      Note: Workspace build target used to validate symbol conflicts
    - Path: go.mod
      Note: Replaces the legacy DuckDB binding with duckdb-go/v2
    - Path: pkg/sql/config.go
      Note: Blank-import registration point for the DuckDB driver
    - Path: pkg/sql/config_test.go
      Note: Regression coverage for DuckDB DSN normalization
    - Path: pkg/sql/template.go
      Note: Confirms the SQL template layer is driver-agnostic
ExternalSources: []
Summary: Inventory of clay SQL/DuckDB usage, migration scope, and build risks.
LastUpdated: 2026-04-18T07:48:09.003176516-04:00
WhatFor: ""
WhenToUse: ""
---


# DuckDB migration analysis and risk assessment

## Executive Summary

Clay only uses DuckDB as a SQL driver registration point in `pkg/sql/config.go`, but that one import is enough to pull the legacy `github.com/marcboeker/go-duckdb` CGO binding into the module graph. The rest of the DuckDB-specific behavior is driver-agnostic: DSN normalization, driver-name canonicalization, and tests that assert the `duckdb://` URL shape.

The migration goal is therefore narrow: replace the old blank import with `github.com/duckdb/duckdb-go/v2`, refresh module metadata, and then verify that the workspace still builds cleanly, especially `go-minitrace`, which already depends on the new DuckDB binding and is the canary for symbol clashes.

## Problem Statement

The repository currently depends on `github.com/marcboeker/go-duckdb v1.8.5` in `clay/go.mod` and imports it in `pkg/sql/config.go` as a side-effect-only driver registration. That keeps the old DuckDB C symbols in the build even though the public API only refers to the generic database/sql driver name `duckdb`.

In this workspace, `go-minitrace` already uses `github.com/duckdb/duckdb-go/v2`. Leaving clay on the older binding creates a risk of duplicate/competing DuckDB symbols when the two modules are built together via `go.work` or when module caches are refreshed. The practical symptom we want to avoid is a build that fails late with CGO/linker conflicts instead of compiling cleanly.

## Current Clay DuckDB Surface

I found the DuckDB touch points in clay are small and centralized:

- `clay/go.mod` — pins the old driver dependency.
- `clay/pkg/sql/config.go` — blank-imports the driver and normalizes `duckdb`/`duck` into the generic driver name.
- `clay/pkg/sql/config_test.go` — verifies DuckDB DSN normalization.
- `clay/pkg/sql/sources.go` — treats `duckdb` as a connection type, but does not import a driver directly.
- `clay/pkg/sql/template.go` / `pkg/sql/query.go` — generic SQL templating and query execution; no DuckDB-specific API usage.
- `clay/pkg/sql/flags/sql-connection.yaml` — docs/help text listing DuckDB as a supported DB type.

This means the migration should not require any query/template refactor. The code path is registering a driver, not calling driver-specific methods.

## Proposed Solution

1. Replace the blank import in `clay/pkg/sql/config.go` with `github.com/duckdb/duckdb-go/v2`.
2. Update `clay/go.mod` to require the v2 binding and drop `github.com/marcboeker/go-duckdb`.
3. Run `go mod tidy` so `go.sum` reflects the new binding and the old package is removed from the dependency graph.
4. Build clay and then build `go-minitrace` from the workspace with the normal `make build` path to confirm the symbol set is clean.
5. If build issues remain, inspect for any stale imports or cached references to the old package before making broader changes.

## Design Decisions

### Keep the `duckdb` driver name unchanged

The application already opens DuckDB connections with `sqlx.Open("duckdb", ...)` or equivalent database/sql calls. The migration should preserve that public contract so configuration, flags, and DSN handling remain stable.

### Keep the existing DSN parsing and normalization logic

The current driver code already normalizes `duckdb://...` into the path shape expected by the Go binding. That logic is independent of the specific DuckDB module and should stay in place unless the new driver shows a different requirement during validation.

### Avoid adding an adapter layer

There is no evidence that clay needs to support both bindings simultaneously. The cleanest path is a direct replacement, not a compatibility shim or wrapper package.

## Alternatives Considered

### Leave clay on `marcboeker/go-duckdb` and only update consumers

Rejected. `go-minitrace` already uses the newer binding, so keeping the old one in clay preserves the symbol conflict risk we are trying to remove.

### Add a compatibility layer or driver abstraction

Rejected. This would add maintenance cost without solving the underlying CGO duplication problem.

### Switch only the import but keep the old module in `go.mod`

Rejected. That would leave unnecessary transitive metadata in the graph and would not fully address the build hygiene issue.

## Implementation Plan

1. Edit `clay/pkg/sql/config.go` to import `github.com/duckdb/duckdb-go/v2`.
2. Update `clay/go.mod` to the new binding and remove the old one.
3. Regenerate `go.sum` with `go mod tidy`.
4. Run focused verification:
   - `go test ./pkg/sql`
   - `go build ./...`
5. Run the workspace-level regression target for `go-minitrace`:
   - `cd ../go-minitrace && make build`
6. If the build still complains about conflicting symbols, inspect the final module graph and any remaining DuckDB imports before widening the change.

## Open Questions

- None at the code level; this looks like a straight dependency swap.
- The only real validation question is whether `go-minitrace` still sees any old DuckDB symbols through transitive metadata or cached artifacts after the module swap.

## References

- `/home/manuel/workspaces/2026-04-18/fix-clay-duckdb/clay/go.mod`
- `/home/manuel/workspaces/2026-04-18/fix-clay-duckdb/clay/pkg/sql/config.go`
- `/home/manuel/workspaces/2026-04-18/fix-clay-duckdb/clay/pkg/sql/config_test.go`
- `/home/manuel/workspaces/2026-04-18/fix-clay-duckdb/go-minitrace/go.mod`
- `/home/manuel/workspaces/2026-04-18/fix-clay-duckdb/go-minitrace/Makefile`
