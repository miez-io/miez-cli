---
status: accepted
date: 2026-09-15
---

# Keep miez distribution state inside `.miez`

## Context and Problem Statement

The current distribution layout places `miez_modules/` and
`miez.lock.yaml` at repository root, while initialization also creates
compiler bookkeeping and workflow-runtime paths below `.miez/`. This mixes
dependency state, compiler internals, and worker-owned work areas across the
repository and makes `.miez/` look like a partially obsolete state model.

The workspace still needs durable state for the active team, target, workflow,
worker overrides, and locally resolved MCP values. Removing that state would
require a separate redesign of the active-team contract and is not needed to
remove the obsolete compiler and workflow scaffolding.

This decision supersedes [ADR 0002: Keep installed modules and the lockfile at
repository root](0002-root-modules-and-lockfile.md).

## Considered Options

* Keep distribution state at repository root and continue creating all current
  `.miez/` records
* Move `miez_modules/` and `miez.lock.yaml` below `.miez/`, retain only the
  required workspace and secret state, and let workers and skills own any
  additional runtime paths
* Remove all persisted `.miez/` state and derive the active team and local
  overrides from installed source on every invocation

## Decision Outcome

Chosen option: **Move distribution state below `.miez/` while retaining only
required workspace state**, because it gives miez one tool-owned home without
removing state that the active-team, local-override, and MCP contracts require.

The target layout is:

* `.miez/miez_modules/<team-id>/` for downloaded team source;
* `.miez/miez.lock.yaml` for resolved revisions and source hashes;
* `.miez/config.yaml` for active team, target, workflow, and local worker,
  skill, and model choices;
* `.miez/mcp.local.yaml` for ignored local MCP values;
* no miez-created `.miez/changes/`, `.miez/runs/`, `manifest.json`, or
  `compiled.json` records.

Extra work directories are created only by the active worker or skill
artifacts that explicitly need them. The compiler manages its deterministic
Copilot output without a persistent compiler manifest or summary file.

### Consequences

* Good, because installed source and its lock record have one clear miez-owned
  location.
* Good, because worker-specific work state is not imposed on every repository
  during initialization.
* Good, because active selections and MCP values remain durable and local
  without being confused with downloaded team source.
* Bad, because every distribution, audit, update, and migration path must
  change its filesystem root and existing workspaces need a compatibility
  migration.
* Bad, because compiler cleanup and rollback can no longer rely on a persisted
  generated-file manifest and must derive the managed output set from the
  previous and next team states.