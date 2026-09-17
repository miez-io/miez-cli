---
description: "miez team package contract. Use when creating or editing miez.yaml, workers/, skills/, tasks/, workflows/, or rules/, when running miez team build, or when a build or check fails. Covers frontmatter fields, id rules, model ids, and what miez renders for Copilot."
applyTo: ["miez.yaml", "miez.generated.yaml", "workers/**", "skills/**", "tasks/**", "workflows/**", "rules/**"]
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
| `tasks/*.md` | Authored reusable task prompts |
| `workflows/*.md` | Authored workflows |
| `rules/*.md` | Optional always-on team rules |
| `miez.generated.yaml` | **Generated. Never edit by hand.** |

`workers/`, `skills/`, `tasks/`, and `workflows/` must all exist, and a package
must declare at least one workflow. Run `miez team build .` after any change; it
validates everything and rewrites `miez.generated.yaml`. A failed build leaves
the previous generated file untouched.

`miez check` validates an *installed* team under `.miez/miez_modules/`. While
authoring this package, use `miez team build .`.

## Identifiers

- Team, worker, skill, workflow, and phase ids: `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`
  — kebab-case, starts with a letter. No underscores, dots, or capitals.
- Model ids also allow dots: `gpt-5.1`, `claude-sonnet-4.5`.

## Frontmatter schemas

Worker frontmatter contains miez-owned fields and optional provider fields. The
worker id is the file name (`workers/architect.md` becomes `architect`); `id` is
not a frontmatter field. Every worker renders as a Copilot custom agent, so
provider fields such as `name`, `description`, `model`, `reasoning-effort`, and
future supported header fields are copied to the generated agent.
Miez consumes `skills` and MCP `tools`; those fields are not copied. The legacy
`kind: agent` field is accepted for compatibility but is no longer required and
is never generated.

```yaml
---
name: Architect
description: ...
model: gpt-5
reasoning-effort: xhigh
skills: [architecture]  # miez-owned skill ids
tools: [github]         # miez-owned MCP ids
---
```

Task frontmatter is strict — the task id is the file name and `description` is
optional:

```yaml
---
description: Analyze an existing solution and document its requirements.
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

- Every `.md` under `workers/` must be a valid worker, every `.md` under
  `tasks/` a valid task, and every `.md` under `workflows/` a valid workflow —
  the scan is recursive. Never park templates
  or notes in those folders.
- A skill must sit exactly at `skills/<skill-id>/SKILL.md`. Other files under
  a skill folder (`assets/`, `references/`) are ignored by the build.
- Never declare `id` in artifact frontmatter — worker, skill, and workflow
  ids are derived from the file path, and a stray `id` field fails the build.
- `skills:` and miez `tools:` must reference things that already exist.
- An agent's model must resolve in the merged catalog: built-ins are `gpt-5`
  (default), `gpt-5.1`, `claude-sonnet-4.5`, `claude-opus-4.5`, `gemini-3-pro`.
  A team adds or overrides ids in `miez.yaml` with a `copilot:` mapping.
- `miez.generated.yaml` metadata, models, and MCP must still match `miez.yaml`.

The package also includes `.vscode/settings.json`, which associates worker,
workflow, task, skill, and rule source with VS Code's built-in prompt languages.
The `chatagent` language provides syntax and highlighting, but not a miez
frontmatter schema or live model completion for arbitrary `workers/*.md` files.

The package's `.vscode/miez-team.code-snippets` adds templates for worker
frontmatter, including miez-only fields. Open the folder containing `miez.yaml`
as a VS Code workspace folder and use the `Agent` language mode if the
association is not picked up. It does not know the current skill or MCP ids;
use `miez team build` for the authoritative cross-reference check.

## What miez renders

miez consumes miez-owned frontmatter and renders the Markdown **body**. Worker
provider frontmatter is retained in the generated Copilot agent:

| Artifact | Output |
|---|---|
| worker agent | `.github/agents/<id>.md`, with provider frontmatter copied, miez-owned fields removed, and the effective Copilot model applied |
| task | `.github/prompts/task-<id>.prompt.md` |
| assigned skill | `.github/skills/<skill-id>/SKILL.md`, linked from the worker |
| selected workflow | `.github/instructions/miez-workflow.instructions.md` |
| `rules/*.md` | `.github/instructions/miez-rules.instructions.md` |

Because miez-owned frontmatter is consumed, the body must stand on its own:
open a worker with its role, not with a bare "Describe this worker".

## Separation

- A worker owns durable persona: identity, motivation, goals, beliefs,
  boundaries. Not procedures, not workflow order.
- A skill owns one reusable procedure, independent of any worker's personality.
- A task owns one reusable objective or partial todo, independent of a worker
  persona and skill assignment.
- A workflow owns ordered phases and worker membership.

Assign skills in worker frontmatter; do not paste skill bodies into workers.
Do not add a worker to a workflow unless asked.
