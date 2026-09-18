---
name: create-task
description: Create or revise one miez task with valid frontmatter and clear instructions.
argument-hint: Describe the task's goal, what the worker should do, and what it should avoid.
---

Create or revise exactly one task in `tasks/<task-id>.md`.

Use a lowercase kebab-case id from the file name. A task describes what needs
to be done: one actionable step that a worker can perform using its assigned
skills. Keep worker identity in workers, reusable procedures in skills, and
multi-worker order in workflows.

Use this template:

```markdown
---
description: One-sentence task description.
---

## What is the goal

Describe the outcome this task should achieve.

## What to do

- [Instruction or step]
- [Instruction or step]

## What not to do

- [Constraint, exclusion, or failure mode to avoid]
```

Write the task as direct instructions to the worker. Use paragraphs and bullet
lists where they make the task clearer; do not add headings that do not serve
this task.

The task id is derived from the file name; do not add an `id` frontmatter
field. Do not invoke another prompt from the task body.

Run `miez team build .` after the edit and report the task id and result.
