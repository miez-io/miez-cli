---
description: "Add a reusable task prompt to this miez team package."
argument-hint: What reusable objective should the task capture?
---

Add exactly one task to `tasks/<task-id>.md`.

Keep the task independent of worker identity and skill assignment. A task
describes what should be done: a reusable objective, partial todo, expected
output, constraints, and completion checks. The selected Copilot agent supplies
the persona and assigned skills when the task prompt is invoked.

Use this frontmatter:

```markdown
---
description: One-sentence task objective.
---
```

The task id is derived from the file name; do not add an `id` frontmatter
field. Do not invoke another prompt from the task body.

Finish by running `miez team build .` and reporting the task id and build result.