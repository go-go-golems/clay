---
Title: Remove Viper and separate sqleton app config from command config
Ticket: SQLETON-02-VIPER-APP-CONFIG-CLEANUP
Status: active
Topics:
    - backend
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Analysis and implementation plan for removing Viper from sqleton, separating app-level config from command-section config, and aligning sqleton with the non-Viper startup pattern already used in pinocchio.
LastUpdated: 2026-04-02T16:58:13.174284484-04:00
WhatFor: Plan a no-backwards-compat cleanup that removes `clay.InitViper(...)` from sqleton and prevents the current `repositories:` config collision between app config and Glazed section config.
WhenToUse: Use this ticket when implementing or reviewing sqleton startup/config cleanup, replacing Viper-based repository discovery, or aligning sqleton with the app-owned config pattern used by pinocchio.
---

# Remove Viper and separate sqleton app config from command config

## Overview

This ticket documents the follow-up cleanup that remains after the SQL command loader work:

1. Remove `clay.InitViper(...)` from `sqleton`.
2. Separate app-level config such as `repositories` from command-section config such as `sql-connection`, `dbt`, and `glazed-command-settings`.
3. Replace the default `AppName: "sqleton"` config-loading behavior with an app-owned Glazed middleware strategy, following the same architectural direction already used in `pinocchio`.

The target state is intentionally not backward-compatible. The goal is a cleaner startup/config model, not another compatibility shim.

## Key Links

- Design doc: `design/01-sqleton-viper-removal-and-app-config-cleanup-design.md`
- Diary: `reference/01-investigation-diary.md`
- Related files: see frontmatter `RelatedFiles`

## Status

Current status: **active**

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
