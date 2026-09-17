---
description: "Add an agent worker to this miez team package, using the strict persona template."
argument-hint: What should this worker own, and which skills should it use?
---

Add exactly one worker to `workers/<worker-id>.md`, following the
[write-workers](../skills/write-workers/SKILL.md) skill and its template.

First read `miez.yaml` and the existing `workers/` and `skills/` files so the
new worker does not overlap an existing one. If the requested scope duplicates
a worker that already exists, say so instead of adding a near-copy.

Every worker renders as a Copilot custom agent. Assign only skills that already
exist; if the worker needs a procedure that does not exist yet, propose the
skill rather than inlining the steps into the persona. Preserve supported
Copilot header fields such as `model` and `reasoning-effort` when they are
useful; miez consumes `skills` and MCP `tools` during compilation.

Do not add the worker to any workflow unless I ask for it.

Finish by running `miez team build .` and reporting the worker id, assigned
skills, and the build result.
