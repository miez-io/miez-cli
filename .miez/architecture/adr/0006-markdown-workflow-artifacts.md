---
status: superseded
date: 2026-09-15
---

# Store workflow coordination as Markdown artifacts

> Superseded by [ADR 0007: Generate the team index from package metadata and
> frontmatter](0007-generated-team-index.md).

## Context and Problem Statement

The team manifest currently owns workflow identities, phases, and worker
participation. That couples the portable team contract to one workflow model
and places coordination data beside worker and model configuration. It also
requires the CLI compiler to reconstruct an instruction from structured
manifest data, even though the workflow is authored as guidance for Copilot.

The workflow source must instead own the coordination narrative, while miez
remains responsible for selecting and installing one always-on instruction.
This decision satisfies `artifact/workflow-artifacts-are-discoverable`,
`workflow/selection-and-routing`,
`workflow/selection-replaces-the-active-instruction`,
`workflow/completion-uses-the-installed-catalog`, and
`render-contract/selected-workflow-is-always-on`.

## Considered Options

* Keep workflow definitions in `miez.generated.yaml`
* Keep workflow Markdown but add a second structured frontmatter contract for
  phases and worker participation
* Discover direct Markdown workflow artifacts and treat their bodies as opaque
  coordination instructions

## Decision Outcome

Chosen option: **Discover direct Markdown workflow artifacts and treat their
bodies as opaque coordination instructions**, because workflow ownership stays
with the workflow artifact, team authors can describe coordination without
changing the manifest schema, and the CLI only needs to discover, select, and
install one instruction.

A workflow artifact is a regular `.md` file directly below `workflows/`. Its
filename stem is its stable workflow id. The team must provide at least one
valid workflow artifact. The selected artifact is rendered to the managed
Copilot workflow instruction with always-on `applyTo: "**"` frontmatter.

The installed module is the workflow catalog. miez does not create a second
persistent registry: `team use` validates and installs the complete source,
and completion reads the active installed module locally. Selecting another
workflow replaces the prior managed instruction atomically.

The old manifest `workflows` field, phase model, workflow enable/disable state,
and CLI worker participation toggles are retired. Worker and skill artifacts
remain independently owned; a workflow may refer to workers in its Markdown
body, but workers do not declare workflow membership.

### Consequences

* Good, because workflow coordination is authored where it is consumed and
  can contain rich Markdown without a second schema.
* Good, because workflow completion is local, deterministic, and naturally
  follows the installed team source.
* Good, because team manifests no longer mix worker persona configuration with
  workflow coordination.
* Bad, because miez cannot validate semantic phases or worker participation in
  an opaque workflow body.
* Bad, because the existing phase-based worker toggle commands are no longer
  meaningful and require a breaking command-surface change.
* Bad, because teams using the old manifest workflow shape need a migration to
  one Markdown file per selectable workflow.
