---
description: "Add a worker to this miez team package, using the strict persona template."
argument-hint: What should this worker own, and should it be a command or an agent?
---

Add exactly one worker to `workers/<worker-id>.md`, following the
[write-workers](../skills/write-workers/SKILL.md) skill and its template.

First read `miez.yaml` and the existing `workers/` and `skills/` files so the
new worker does not overlap an existing one. If the requested scope duplicates
a worker that already exists, say so instead of adding a near-copy.

Decide `kind` deliberately: `command` for a manually invoked prompt, `agent`
for a persistent, delegated worker that may carry its own `model`. Assign only
skills that already exist; if the worker needs a procedure that does not exist
yet, propose the skill rather than inlining the steps into the persona.

Do not add the worker to any workflow unless I ask for it.

Finish by running `miez team build .` and reporting the worker id, kind,
assigned skills, and the build result.
