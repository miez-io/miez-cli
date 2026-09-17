---
name: write-workers
description: "Author a miez worker persona from the template. Use when creating a worker, rewriting a weak or generic persona, or splitting procedures out of a worker into skills."
---

# Write Workers

A worker is the miez artifact that renders as a persistent Copilot agent. Its
Markdown body is rendered with the provider frontmatter retained, so the body
alone must carry the persona and miez-owned fields must remain configuration.

Always start from [assets/worker-template.md](./assets/worker-template.md) and
keep its section order.

## Use the agent worker form

Every worker renders to `.github/agents/<id>.md`. Workers are persistent
Copilot agents; reusable task objectives belong in `tasks/<task-id>.md` and
render to prompts. The optional legacy `kind: agent` field is not needed.

## Procedure

1. Name the durable role. Start from what this worker *judges and protects*,
   not from a technology list or a workflow step.
2. Choose a kebab-case id and create `workers/<worker-id>.md`.
3. Copy the template structure and fill every section with content that would
   survive a change of project, workflow, or toolchain.
4. Write the body so it reads correctly beneath the retained provider
   frontmatter. Open with the role, not a placeholder sentence.
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
- Do not reference a skill or MCP tool that does not exist yet.
- Do not place the template or any other non-worker `.md` under `workers/`;
  the build treats every file there as a worker.