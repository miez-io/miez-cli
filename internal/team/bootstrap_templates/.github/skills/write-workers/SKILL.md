---
name: write-workers
description: "Author a miez worker persona from the strict template. Use when creating a worker, rewriting a weak or generic persona, deciding between kind command and agent, or splitting procedures out of a worker into skills."
---

# Write Workers

A worker is the one miez artifact with a strict shape. Its Markdown body is
rendered directly into a Copilot prompt or agent with the frontmatter stripped,
so the body alone must carry the persona.

Always start from [assets/worker-template.md](./assets/worker-template.md) and
keep its section order.

## Choose the kind first

| | `kind: command` | `kind: agent` |
|---|---|---|
| Rendered as | `.github/prompts/<id>.prompt.md` | `.github/agents/<id>.md` |
| Invocation | manually, per task | persistent, delegated |
| `model:` | forbidden | optional, defaults to team `default_model` |

Pick `agent` only when the worker benefits from its own model or from being
delegated to autonomously. Otherwise pick `command`.

## Procedure

1. Name the durable role. Start from what this worker *judges and protects*,
   not from a technology list or a workflow step.
2. Choose a kebab-case id and create `workers/<worker-id>.md`.
3. Copy the template structure and fill every section with content that would
   survive a change of project, workflow, or toolchain.
4. Write the body so it reads correctly with no frontmatter, since miez strips
   it. Open with the role, not a placeholder sentence.
5. Move any repeatable procedure, checklist, or format into a skill and assign
   it with `skills: [id]`. Never paste a skill body into the persona.
6. Leave workflow membership alone unless explicitly requested.
7. Run `miez team build .` and fix anything it reports.

## Persona quality bar

- Identity, motivation, goals, beliefs, and boundaries are all filled in with
  specifics — no leftover bracket placeholders.
- The worker stays useful if its assigned skills change.
- Boundaries name what this worker does *not* own, including at least one
  scope boundary with another worker.
- Nothing in the body describes phases, handoffs, or "after X, do Y".

## Do not

- Do not add workflow phases, handoff order, or a `Collaboration` section.
- Do not set `model` on a `kind: command` worker.
- Do not reference a skill or MCP tool that does not exist yet.
- Do not place the template or any other non-worker `.md` under `workers/`;
  the build treats every file there as a worker.