---
description: "Create or restructure a workflow in this miez team package, wiring existing workers into ordered phases."
argument-hint: What process should the workflow describe, and in what order?
---

Create or update one workflow in `workflows/<workflow-id>.md`.

Frontmatter is strict and must declare a non-empty `name` and at least
one phase. The workflow id is the file name; do not declare `id`. Each phase
needs a kebab-case `id` unique within the workflow and a
non-empty `workers:` list naming workers that already exist under `workers/`.

```yaml
---
name: Default
phases:
  - id: design
    workers: [architect]
  - id: implement
    workers: [implementer, reviewer]
---
```

Phases are ordered, and a phase may hold several workers when they act in the
same step. Use the Markdown body to explain how the phases hand off — that body
becomes the always-on workflow instruction, so write it for whoever executes
the process.

Only the workflow decides membership and order; never move that choreography
into a worker persona. If the process needs a worker that does not exist yet,
propose it instead of inventing an id here.

Finish by running `miez team build .` and reporting the workflow id, its phase
order, and the build result.
