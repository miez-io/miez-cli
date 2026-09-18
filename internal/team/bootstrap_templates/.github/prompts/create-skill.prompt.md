---
name: create-skill
description: Create or revise one reusable miez skill with valid frontmatter and clear procedures.
argument-hint: Describe the repeatable capability, method, and guardrails.
---

Create or revise exactly one reusable skill at
`skills/<skill-id>/SKILL.md`.

Use a lowercase kebab-case id matching the directory name. Keep the skill
independent of worker personality and workflow order. Do not put a worker's
identity, motivation, or workflow choreography in the skill.

Every skill must use this frontmatter:

```yaml
---
name: skill-id
description: One-sentence reusable capability.
user-invocable: true
disable-model-invocation: false
---
```

Do not add `license`, `compatibility`, or `metadata` fields to the skill
frontmatter.

There is no fixed body template. Choose sections that fit the capability. A
skill should explain how to apply a reusable capability and the guardrails that
matter; it does not own a task's expected output. Add references or assets
beside `SKILL.md` when they make the procedure reusable.

Inspect existing workers before assigning the skill. Assign it explicitly in
worker frontmatter only when the capability is relevant. Do not silently add a
worker to a workflow.

Run `miez team build .` after the edit and report the resulting validation.