---
name: create-workflow
description: Create or revise one miez workflow with ordered worker phases.
argument-hint: Describe the process and the order in which workers act.
---

Create or revise exactly one workflow in `workflows/<workflow-id>.md`.

Use a lowercase kebab-case id from the file name. A workflow describes what
the team does and in which order. Reference only existing workers, and keep
persona content in workers and procedures in skills.

Use this template, replacing the worker ids with workers that already exist:

```markdown
---
name: Workflow Name
phases:
  - id: work
    workers: [worker-id]
---

# Workflow Name

Describe how the team coordinates this workflow.
```

Run `miez team build .` after the edit and report the workflow id, phase order,
and result.
