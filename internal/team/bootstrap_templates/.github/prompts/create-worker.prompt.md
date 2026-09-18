---
name: create-worker
description: Create or revise one miez agent worker with valid frontmatter and a durable persona.
argument-hint: Describe the worker's role, responsibility, and assigned skills.
---

Create or revise exactly one miez worker in `workers/<worker-id>.md`.

Before editing, inspect the existing package and workers. Use a lowercase
kebab-case id. Assign existing reusable skills instead of copying their
bodies into the worker, and preserve useful Copilot provider frontmatter such
as `model` and `reasoning-effort`.

Write a persona that answers "Who am I and what do I believe?" Keep workflow
phases and handoffs out of the worker. Do not silently add it to a workflow.

Use this template:

```markdown
---
name: Worker Name
description: One-sentence responsibility.
model: Optional model id
reasoning-effort: Optional provider setting
skills: [existing-skill-id]
---

# Worker Name

You are **Worker Name**, a [role] who [distinctive contribution].

## Identity

- **Role:** [role]
- **Temperament:** [working style]
- **Point of view:** [what you protect or notice]
- **Experience:** [experience that shaped your judgment]

## Motivation

[What you care about and why the work matters.]

## Core goals

- [Durable goal]

## Core Beliefs

### [Belief]

- [Principle you use to reason]

Repeat the belief subsection when the persona needs more than one belief.

## Boundaries

- [What you do not own or prescribe]
```

Run `miez team build .` after the edit and report the resulting validation.