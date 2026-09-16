---
description: "Add a reusable skill to this miez team package."
argument-hint: What repeatable capability should this skill describe?
---

Add exactly one skill at `skills/<skill-id>/SKILL.md`, with `id` matching the
folder name. That is the only required frontmatter.

A skill is deliberately less structured than a worker: there is no mandatory
section list. Give it whatever shape the capability actually needs — a short
procedure, a checklist, a decision table, worked examples. Supporting files may
live beside it in `assets/` or `references/`; the build ignores everything that
is not `SKILL.md`.

Keep it reusable: no worker personality, no tone, no workflow order. The skill
should still make sense if a different worker picks it up. Remember that miez
strips the frontmatter when rendering, so the body must stand alone.

Before writing, check `skills/` for an existing skill that already covers this;
extending one is better than adding a near-duplicate. Assign the skill to a
worker only if I ask, and do it in that worker's `skills:` frontmatter.

Finish by running `miez team build .` and reporting the skill id and result.
