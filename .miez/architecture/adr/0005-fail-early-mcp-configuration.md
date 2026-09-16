---
status: accepted
date: 2026-09-15
---

# Fail early when required MCP configuration is unresolved

## Context and Problem Statement

Workers may depend on MCP servers whose environment values are represented by
placeholders in `miez.generated.yaml`. A missing value can either be left unresolved for
the downstream tool to discover later, or be resolved before installation.
The implementation resolves environment/override values first, prompts only in
an interactive terminal, stores supplied values locally, and fails before
installation when a required value remains unavailable.

This decision satisfies `mcp-secrets/missing-tool-configuration-prompts-interactively`
and `mcp-secrets/tool-configuration-values-stay-local-and-uncommitted`.

## Considered Options

* Preserve unresolved placeholders and let the MCP process fail later
* Require environment variables only and never prompt
* Resolve early, prompt interactively when possible, and fail clearly in
  non-interactive or CI execution

## Decision Outcome

Chosen option: **Resolve early with an interactive prompt and a clear
non-interactive failure**, because installation should not appear successful
while its enabled workers have unusable tool configuration, and a CI process
must never hang waiting for input.

### Consequences

* Good, because configuration errors are reported before source replacement or
  generated output activation.
* Good, because interactive operators can supply values once and reuse them
  from the ignored local store.
* Good, because names containing `token` or `key` are masked in terminal
  prompts.
* Bad, because installation can require operator input even when the team
  source itself is otherwise valid.
* Bad, because the local store is intentionally untracked workspace state and
  must be provisioned separately on another machine.
