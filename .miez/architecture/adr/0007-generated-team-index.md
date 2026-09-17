---
status: accepted
date: 2026-09-15
---

# Generate the team index from package metadata and frontmatter

## Context and Problem Statement

A team needs a compact, deterministic catalog for installation, validation,
workflow selection, shell completion, worker skill assignment, model selection,
and MCP resolution. Manually maintaining that catalog in `miez.generated.yaml` duplicates
configuration already present in Markdown frontmatter and allows the package
to drift from its artifacts.

At the same time, installation must not parse arbitrary authoring frontmatter.
It needs a stable generated contract that can be locked, audited, and consumed
without a second discovery pass. A team also needs package metadata such as
name, version, author, model catalog, and MCP declarations that cannot be
inferred from one worker or workflow file.

This decision satisfies `artifact/team-index-is-generated`,
`artifact/frontmatter-declares-cli-configuration`,
`team-authoring/build-produces-installable-package`,
`team-distribution/install-uses-generated-team-index`, and
`workspace/active-selection-is-local-state`.

## Considered Options

* Continue authoring `miez.generated.yaml` manually and treat Markdown as referenced content
* Discover Markdown frontmatter during every install and CLI operation without a
  generated index
* Author package metadata in `miez.yaml`, declare artifact configuration in
  Markdown frontmatter, and generate `miez.generated.yaml` as the installable index

## Decision Outcome

Chosen option: **Author package metadata in `miez.yaml`, declare artifact
configuration in Markdown frontmatter, and generate `miez.generated.yaml` as the
installable index**, because it removes duplicated artifact configuration while
keeping installation deterministic and offline after download.

A team authoring package contains:

- `miez.yaml` for team metadata, author information, version, model catalog,
  default model, and MCP declarations;
- `workers/` Markdown files whose frontmatter declares worker id, optional
  miez skills, model, and MCP tools; provider frontmatter remains available to
  the Copilot renderer;
- `skills/` Markdown files whose frontmatter declares skill identity;
- `workflows/` Markdown files whose frontmatter declares workflow identity,
  display name, and ordered worker phases.

`miez team bootstrap <team-id>` creates the package metadata file and one
frontmatter template in each artifact directory. `miez team build` validates
all package metadata, frontmatter, identifiers, paths, references, models,
tools, and phases, then writes `miez.generated.yaml` atomically. The generated index
contains the complete operational catalog and is the only team catalog used by
installation and CLI completion. Its referenced Markdown bodies remain the
runtime content.

The package's `miez.yaml` is included in the installed `.miez/miez_modules/<team-id>/`
source for metadata and provenance. The installing repository's root
`miez.yaml` is a workspace/package manifest containing metadata and named team
sources; `.miez/miez.lock.yaml` remains the source of resolved commit and integrity
state. Active team/workflow selections remain in `.miez/config.yaml` so local
state never mutates or invalidates installed package files.

The workflow body remains authored coordination content, but its frontmatter
is machine-readable. Generated workflow entries retain phase metadata, so
miez can provide deterministic completion, validate worker references, and
apply local worker participation overrides without adding workflow fields to
worker frontmatter.

### Consequences

* Good, because authors configure an artifact once and the CLI index is derived
  from the same source.
* Good, because installation can rely on `miez.generated.yaml` without scanning or
  interpreting authoring frontmatter.
* Good, because generated catalogs can be compared, locked, audited, and
  reproduced deterministically.
* Good, because package metadata and workspace dependency metadata can use the
  familiar `miez.yaml` package-manifest shape.
* Bad, because `miez.generated.yaml` becomes a generated file and manual edits are
  overwritten or rejected by the next build.
* Bad, because build must validate cross-file references and provide actionable
  frontmatter errors before a package can be installed.
* Bad, because `miez.yaml` has two contextual roles, team package and consuming
  workspace, which requires an explicit manifest type and clear validation.
