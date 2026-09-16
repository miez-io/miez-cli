---
status: accepted
date: 2026-09-15
---

# Render rules separately from workflow routing

## Context and Problem Statement

A team can provide general guidance in a `rules/` directory and can also
provide a selected workflow Markdown artifact. Both become Copilot
instructions, but they have different ownership and replacement behavior:
general rules remain active while the selected workflow instruction changes.

This decision satisfies `render-contract/rules-are-a-separate-always-on-instruction-from-workflow-routing`.

## Considered Options

* Concatenate rules into the selected workflow instruction
* Render rules as an independent always-on instruction and the selected
  workflow as a separate managed instruction
* Treat `rules/` as worker skills instead of instructions

## Decision Outcome

Chosen option: **Render rules and the selected workflow as two independent
instructions**, because one is persistent team-wide guidance and the other is a
replaceable execution-routing instruction.

### Consequences

* Good, because selecting a different workflow replaces routing without
  removing general team guidance.
* Good, because each concern has a stable generated path and can be tested and
  audited independently.
* Bad, because compilation manages two instruction files and must maintain
  their distinct cleanup behavior.
* Bad, because rule content is concatenated as freeform Markdown; the team
  author remains responsible for avoiding contradictory guidance.
