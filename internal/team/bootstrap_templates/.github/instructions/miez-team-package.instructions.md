---
description: "miez team package contract. Use when creating or editing miez.yaml, workers/, skills/, workflows/, or rules/, when running miez team build, or when a build or check fails. Covers frontmatter schemas, id rules, model ids, and what miez renders for Copilot."
applyTo: ["miez.yaml", "miez.generated.yaml", "workers/**", "skills/**", "workflows/**", "rules/**"]
---

# miez team package contract

This repository is a miez team package: Markdown artifacts plus `miez.yaml`,
compiled by `miez team build .` into `miez.generated.yaml`.

## Authored vs generated

| Path | Role |
|---|---|
| `miez.yaml` | Authored package metadata, model catalog, MCP declarations |
| `workers/*.md` | Authored workers |
| `skills/<skill-id>/SKILL.md` | Authored skills |
| `workflows/*.md` | Authored workflows |
| `rules/*.md` | Optional always-on team rules |
| `miez.generated.yaml` | **Generated. Never edit by hand.** |

`workers/`, `skills/`, and `workflows/` must all exist, and a package must
declare at least one workflow. Run `miez team build .` after any change; it
validates everything and rewrites `miez.generated.yaml`. A failed build leaves
the previous generated file untouched.

`miez check` validates an *installed* team under `.miez/miez_modules/`. While
authoring this package, use `miez team build .`.

## Identifiers

- Team, worker, skill, workflow, and phase ids: `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`
  — kebab-case, starts with a letter. No underscores, dots, or capitals.
- Model ids also allow dots: `gpt-5.1`, `claude-sonnet-4.5`.

## Frontmatter schemas

Worker frontmatter is strict — an unknown field fails the build. The worker
id is the file name (`workers/architect.md` becomes `architect`); `id` is not
a frontmatter field.

```yaml
---
kind: agent          # required: command | agent
model: gpt-5         # agent only; never on a command worker
skills: [architecture]  # optional; ids that exist under skills/
tools: [github]      # optional; ids declared in miez.yaml mcp:
name: Architect      # optional, human-readable only
description: ...     # optional, human-readable only
---
```

Workflow frontmatter is strict; the workflow id is the file name:

```yaml
---
name: Default        # required, non-empty
description: ...     # optional
phases:              # required, at least one
  - id: design       # kebab-case, unique within the workflow
    workers: [architect]   # at least one, each an existing worker id
---
```

A skill's id is its directory name. SKILL.md frontmatter needs no `id`; extra
fields like `description` are tolerated:

```yaml
---
description: ...
---
```

## Rules that break the build

- Every `.md` under `workers/` must be a valid worker, and every `.md` under
  `workflows/` a valid workflow — the scan is recursive. Never park templates
  or notes in those folders.
- A skill must sit exactly at `skills/<skill-id>/SKILL.md`. Other files under
  a skill folder (`assets/`, `references/`) are ignored by the build.
- Never declare `id` in artifact frontmatter — worker, skill, and workflow
  ids are derived from the file path, and a stray `id` field fails the build.
- `skills:` and `tools:` must reference things that already exist.
- An agent's model must resolve in the merged catalog: built-ins are `gpt-5`
  (default), `gpt-5.1`, `claude-sonnet-4.5`, `claude-opus-4.5`, `gemini-3-pro`.
  A team adds or overrides ids in `miez.yaml` with a `copilot:` mapping.
- `miez.generated.yaml` metadata, models, and MCP must still match `miez.yaml`.

## What miez renders

miez strips frontmatter and renders the Markdown **body**:

| Artifact | Output |
|---|---|
| `kind: command` worker | `.github/prompts/<id>.prompt.md` |
| `kind: agent` worker | `.github/agents/<id>.md`, with the resolved Copilot model prepended |
| assigned skill | `.github/skills/<skill-id>/SKILL.md`, linked from the worker |
| selected workflow | `.github/instructions/miez-workflow.instructions.md` |
| `rules/*.md` | `.github/instructions/miez-rules.instructions.md` |

Because frontmatter is dropped, the body must stand on its own: open a worker
with its role, not with a bare "Describe this worker".

## Separation

- A worker owns durable persona: identity, motivation, goals, beliefs,
  boundaries. Not procedures, not workflow order.
- A skill owns one reusable procedure, independent of any worker's personality.
- A workflow owns ordered phases and worker membership.

Assign skills in worker frontmatter; do not paste skill bodies into workers.
Do not add a worker to a workflow unless asked.
