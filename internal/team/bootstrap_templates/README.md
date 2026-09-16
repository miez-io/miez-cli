# {{TEAM_ID}}

This directory is a miez team authoring package. It contains the source files
that miez validates and turns into a portable GitHub Copilot team.

## Quick start

1. Edit `miez.yaml`, the files below `workers/`, `skills/`, `tasks/`, and
   `workflows/`.
2. Run `miez team build .` to validate the package and generate
   `miez.generated.yaml`.
3. Review the generated catalog, then commit the authored files and the
   generated index to the team repository.
4. Install the published team with `miez team install <github-url>`.

The generated index is derived output. Do not edit it by hand. A failed build
leaves the previous generated index unchanged.

## Package layout

```text
miez.yaml                         team metadata and model/MCP declarations
workers/<worker-id>.md            worker personas and capabilities
skills/<skill-id>/SKILL.md        reusable skills assigned to workers
tasks/<task-id>.md                reusable task objectives rendered as prompts
workflows/<workflow-id>.md        ordered phases and agent membership
rules/<name>.md                   optional always-on team rules
miez.generated.yaml               generated operational catalog
```

`workers/`, `skills/`, `tasks/`, and `workflows/` are required, and a package
must declare at least one workflow. `rules/` is optional: every top-level `.md`
there is concatenated into one always-on Copilot instruction, separate from
the active workflow.

The `.github/` directory in this starter package contains authoring support for
GitHub Copilot. It is not scanned as an operational worker, skill, or workflow
and is not added to `miez.generated.yaml`.

## Workers

Every worker is a persistent Copilot custom agent. The worker id is the file
name: `workers/architect.md` becomes `architect`. Use `kind: agent` and assign
reusable skills with `skills: [skill-id]`; workers may reference declared MCP
servers with `tools: [mcp-id]`.

Keep durable identity, motivation, goals, beliefs, and boundaries in the worker
body. Put repeatable procedures and technology-specific methods in skills.
Keep workflow phases and handoff order in workflow files.

Example:

```markdown
---
kind: agent
model: gpt-5
skills: [architecture]
---
# Architect

Describe the worker's durable persona and decision posture.
```

## Tasks

A task describes what to do: a reusable objective or partial todo independent
of worker identity and skill assignment. Its id is the file name, and its body
is rendered as a GitHub Copilot prompt under `.github/prompts/task-<id>.prompt.md`.
Select a worker agent, then invoke the task prompt in that agent's context.
Tasks do not invoke other prompts from inside their body.

```markdown
---
description: Analyze an existing solution and document its requirements.
---
# Analyze an existing solution

Describe the reusable objective and expected output.
```

## Skills and workflows

A skill is reusable and independent of a worker persona. Its id is its
directory name (`skills/<skill-id>/SKILL.md`); do not declare `id` in
frontmatter. Assign it from worker frontmatter or with
the local `miez worker skill add` command after installation.

A workflow's id is its file name; its frontmatter declares `name` and ordered
`phases`. Every phase
lists existing worker ids. Workflow membership belongs only in the workflow;
do not put routing or handoff choreography in a worker or skill.

## Local CLI controls

After installation, the active workspace can select a workflow and adjust local
worker participation, skill assignments, or agent models without changing the
team source:

```bash
miez workflow use <workflow-id>
miez workflow worker enable|disable <worker-id>
miez worker skill add|remove <worker-id> <skill-id>
miez worker model list
miez worker model set <worker-id> <model-id>
miez check
```

The `.miez/` directory stores workspace-local state. The team source remains
the package files and the generated catalog.

## Copilot authoring helpers

`.github/` holds the authoring support for this package. It is never scanned as
a worker, skill, or workflow and never appears in `miez.generated.yaml`.

| Helper | Purpose |
|---|---|
| `instructions/miez-team-package.instructions.md` | The package contract: layout, frontmatter schemas, id rules, what miez renders. Attaches automatically when you edit an authored artifact. |
| `prompts/new-worker.prompt.md` | Add one worker from the strict template. |
| `prompts/new-skill.prompt.md` | Add one reusable skill. |
| `prompts/new-task.prompt.md` | Add one reusable task prompt. |
| `prompts/new-workflow.prompt.md` | Create or restructure a workflow's phases. |
| `skills/write-workers/` | Persona authoring workflow plus `assets/worker-template.md`. |
| `skills/review-team-package/` | Pre-publish review, with a build-error reference. |

Workers follow a strict template because their body is rendered straight into a
Copilot prompt or agent. Skills are intentionally loose: the folder name is the
id and no frontmatter field is required.

The helpers may edit source files, but `miez.generated.yaml` is always produced
by `miez team build .`, never by hand.