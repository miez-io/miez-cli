# {{TEAM_ID}}

## What Miez CLI means

Have you ever felt overwhelmed by the many possibilities for structuring your
work around AI agents? There are plenty of tools and approaches, but none of
them felt complete. If you have had the same experience, you are in the right
place.

Miez is a small, simple CLI built around a domain for structuring and building
Markdown artifacts.

Instead of starting with questions about where to place and how to write `AGENTS.md`,
instructions, prompts, rules, skills, or agents, Miez starts with a realistic,
easy-to-understand model of a team. It defines four artifact roles:

- **Workers** answer "who I am and what I believe". They are personas with a
  durable identity, motivation, goals, beliefs, judgment, and boundaries. Each
  worker becomes a selectable GitHub Copilot custom agent.
- **Skills** answer "how I should do something". They are reusable,
  worker-independent know-how, procedures, and guardrails that workers can use.
- **Tasks** answer "what needs to be done". They define actionable,
  worker-neutral objectives that different workers can perform.
- **Workflows** answer "what I do and in which order". They coordinate workers
  through ordered phases. Miez renders that coordination; GitHub Copilot
  executes the work.

Together, these artifacts form a team.

This model reflects how real-world teams are structured. You can have one team
or many teams, with workers that have specific skill sets, execute tasks, and
collaborate through workflows.

You may already have tried approaches such as spec-driven development or fully
automated agent workflows, but something was still missing. Spec-driven
approaches often come with a heavy CLI and more management than is actually
necessary. With a proper domain setup, spec-driven development can simply mean
shipping the right workers, skills, and tasks. The approach is built from
Markdown artifacts, not from yet another complex CLI that users have to learn.
The CLI should support installation, validation, and configuration without
becoming another system to manage. Automated workflows can also be too rigid
for use cases where you need to add or remove agents and workflows while
keeping track of how the whole system works.

This is where miez-cli comes in. Once you have defined your artifacts, a
handful of simple commands helps you manage your team without forcing it to
remain static. Miez can bootstrap new teams and also acts as a small, APM-like
installer for teams. You can install multiple teams and switch between them
without managing their artifacts manually. You can configure workers
dynamically, assign additional skills, change models, activate a workflow, and
adjust which workers participate in it.

All of this could be built directly with GitHub Copilot artifacts. In the end,
Miez resolves its artifacts into GitHub Copilot artifacts. But thinking only in
terms of prompts, instructions, rules, skills, or agents is not helpful; it is
confusing. The Miez domain brings clarity by letting you think in a real-world
model, while the CLI keeps the most important artifacts configurable instead of
static.

## This repository

This repository is a Miez team authoring package. It contains the source files
that Miez validates and turns into a portable GitHub Copilot team.

## Quick start

1. Edit `miez.yaml` and the files below `workers/`, `skills/`, `tasks/`, and
   `workflows/`.
2. Run `miez team build .` to validate the package and generate
   `miez.generated.yaml`.
3. Review and commit the authored files together with the generated catalog.
4. Install the published team with `miez team install <github-url>`.

The generated catalog is derived output; never edit it by hand. A failed build
leaves the previous catalog unchanged.

## Package layout

```text
miez.yaml                         team metadata and model/MCP declarations
workers/<worker-id>.md            worker personas and capabilities
skills/<skill-id>/SKILL.md        reusable skills assigned to workers
tasks/<task-id>.md                reusable task instructions rendered as prompts
workflows/<workflow-id>.md        ordered worker phases
miez.generated.yaml               generated operational catalog
```

The `.github/` directory contains authoring support for GitHub Copilot. It is
not scanned as an operational worker, skill, or workflow and is not added to
`miez.generated.yaml`.

## Authoring model

### Workers

A worker renders as a persistent GitHub Copilot custom agent. Its id comes from
the file name: `workers/architect.md` becomes `architect`.

Workers can assign skills with `skills: [skill-id]`, reference declared MCP
servers with `tools: [mcp-id]`, and select a model. Keep durable identity,
motivation, goals, beliefs, judgment, and boundaries in the worker. Put
repeatable procedures in skills and workflow order in workflows.

### Skills

A skill is reusable and independent of a worker persona. Its id comes from its
directory name, `skills/<skill-id>/SKILL.md`; do not declare `id` in its
frontmatter. Skills use `user-invocable: true` and
`disable-model-invocation: false` and should contain only the sections needed
by their capability.

### Tasks

A task is a single actionable instruction that a worker can perform using its
assigned skills. Its id comes from the file name, and its frontmatter contains
a description. Current tasks use `What is the goal`, `What to do`, and
`What not to do` to make the intent and boundaries clear.

### Workflows

A workflow's id comes from its file name. Its frontmatter declares a `name` and
ordered `phases`; each phase lists existing worker ids. Workflow membership and
handoff order belong only in the workflow.

## CLI controls

After installation, the active workspace can select a workflow and adjust
worker participation, skill assignments, or models without changing the team
source:

```bash
miez team list
miez team use <team-id>
miez team bootstrap <team-id>
miez team build [path]
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

- `.github/instructions/miez-artifact-authoring.instructions.md` defines the
  artifact boundaries.
- `.github/prompts/` contains prompts for creating and verifying workers,
  skills, tasks, and workflows.
- `.github/prompts/commit.prompt.md` helps create focused commits.

Read the relevant helper before changing the team structure. The helpers may
edit source files, but `miez.generated.yaml` must always be regenerated with
`miez team build`.