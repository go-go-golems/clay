---
Title: Investigation diary
Ticket: SQLETON-02-VIPER-APP-CONFIG-CLEANUP
Status: active
Topics:
    - backend
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Chronological notes for the Viper/app-config cleanup follow-up ticket.
LastUpdated: 2026-04-02T17:10:00-04:00
WhatFor: Record the reasoning and evidence behind the follow-up Viper removal ticket.
WhenToUse: Read this when reviewing why the ticket exists or how the implementation plan was derived.
---

# Investigation diary

## 2026-04-02 17:10 Follow-up Ticket Creation

### Goal

Create a follow-up ticket for the remaining sqleton startup/config cleanup after the SQL command loader work was finished.

### Why this ticket is needed

The previous ticket uncovered one design issue that was outside the SQL command loader cleanup itself:

- `sqleton` still uses `clay.InitViper("sqleton", rootCmd)` for startup
- command parsing still uses the default Glazed `AppName: "sqleton"` config loading behavior
- the same `config.yaml` file is therefore interpreted both as app config and as section-config

That becomes visible with the `repositories:` key:

```yaml
repositories:
  - /path/to/repo
```

This is valid app config for repository discovery but invalid section-config for Glazed field loading, because the config middleware expects top-level section maps.

### Reference comparison

I compared this with `pinocchio`, which already uses the healthier pattern:

- `clay.InitGlazed(...)`
- app-owned repository config loading
- app-owned parser middleware decisions

That makes `pinocchio` the right reference implementation for the migration direction, even though sqleton’s concrete sections differ.

### Decision

The new ticket should not try to preserve old startup/config behavior.

The requested direction is:

- no backward compatibility layer
- remove Viper directly
- make app config ownership explicit
- make command config ownership explicit

### Deliverables created

- ticket workspace
- design/implementation guide
- this diary

