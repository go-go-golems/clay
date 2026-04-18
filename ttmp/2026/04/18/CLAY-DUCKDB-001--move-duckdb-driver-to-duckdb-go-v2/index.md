---
Title: Move DuckDB driver to duckdb-go/v2
Ticket: CLAY-DUCKDB-001
Status: complete
Topics:
    - backend
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/18/CLAY-DUCKDB-001--move-duckdb-driver-to-duckdb-go-v2/design-doc/01-duckdb-migration-analysis-and-risk-assessment.md
      Note: Primary analysis and implementation plan for the DuckDB migration
ExternalSources: []
Summary: Replaced clay's DuckDB driver binding with duckdb-go/v2 and verified workspace builds.
LastUpdated: 2026-04-18T07:51:31.41743099-04:00
WhatFor: Track the DuckDB binding migration and build validation.
WhenToUse: Use when reviewing the DuckDB driver swap or symbol-conflict regression checks.
---



# Move DuckDB driver to duckdb-go/v2

## Overview

Clay's SQL layer only needed a driver registration swap, but that one import was enough to drag the legacy DuckDB CGO binding into the workspace. This ticket moved clay to duckdb-go/v2 and verified the build paths that mattered most: clay itself and the go-minitrace consumer that had previously exposed symbol conflicts.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **complete**

## Topics

- backend

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
