---
status: accepted
date: 2026-09-15
---

# Use GitHub Copilot as the only render target

## Context and Problem Statement

miez-cli is deliberately scoped to a narrow, current contract. The compiler
must decide whether target-specific output is generated for multiple
providers or whether one provider is a deliberate boundary. The current code
defines `copilot` as its only target and rejects other target names.

This decision satisfies `render-contract/github-copilot-is-the-only-render-target`.

## Considered Options

* Support a multi-target provider abstraction, including Cursor and VS Code
* Support GitHub Copilot only for this version
* Accept arbitrary target names and ignore targets the compiler does not know

## Decision Outcome

Chosen option: **Support GitHub Copilot only for this version**, because the
current product scope is the Copilot artifact contract and explicit rejection
is safer than silently producing incomplete output.

### Consequences

* Good, because worker, skill, rule, workflow, and model rendering have one
  concrete provider contract.
* Good, because unsupported targets fail visibly during initialization,
  checking, and compilation.
* Bad, because teams authored for Cursor or VS Code cannot be rendered by this
  version without a future target implementation or migration.
* Bad, because the model catalog stores one `copilot` mapping rather than the
  provider mappings used by the earlier multi-target shape.
