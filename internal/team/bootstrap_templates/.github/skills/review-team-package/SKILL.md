---
name: review-team-package
description: "Review a miez team package before publishing. Use when miez team build fails, when checking worker, skill, task, and workflow boundaries, or when verifying that miez.generated.yaml still matches the authored source."
---

# Review Team Package

A read-only pass over the package. It reports findings and the smallest repair;
it does not restructure artifacts on its own.

## Procedure

1. Run `miez team build .`. If it fails, map the message with
   [references/build-errors.md](./references/build-errors.md), fix the cause,
   and re-run until it passes.
2. Confirm `workers/`, `skills/`, `tasks/`, and `workflows/` exist and that
   nothing parked under `workers/`, `tasks/`, or `workflows/` is anything other
   than a real artifact.
3. Confirm every skill lives at `skills/<skill-id>/SKILL.md` with an `id`
   matching its folder.
4. Confirm every `skills:` and `tools:` reference on a worker resolves, and
   that every workflow phase lists existing worker ids.
5. Confirm every worker uses `kind: agent` and every agent model resolves in
   the merged catalog.
6. Read each worker body as if the frontmatter were gone, which is how Copilot
   receives it. Flag placeholder text, missing persona sections, and anything
   that reads like workflow choreography.
7. Read each skill body for worker-specific persona leaking in.
8. Confirm `miez.generated.yaml` is regenerated, not hand-edited, and that its
   metadata still matches `miez.yaml`.

## Boundary findings to report

| Symptom | Correct home |
|---|---|
| Worker lists phases, handoffs, or "then pass to…" | the workflow |
| Worker carries a reusable checklist or format | a skill |
| Skill describes a personality or tone | the worker |
| Same procedure duplicated across workers | one shared skill |
| Worker added to a workflow nobody asked for | remove it |

## Output

Report `clean`, `drift`, or `blocked`. For each finding give the file, the rule
or reference that is violated, and the smallest repair. Do not rewrite files
during the review.
