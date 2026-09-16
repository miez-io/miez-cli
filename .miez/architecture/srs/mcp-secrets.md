# SRS: MCP secrets

ISO/IEC/IEEE 29148 software requirements specification using EARS.

**Status:** current implementation contract, captured 2026-09-15.

## 1. Introduction

This capability defines how a team declares MCP tool dependencies and how miez
obtains their environment values. It exists to make tool configuration explicit
without placing operator secrets in portable team artifacts.

## 2. Stakeholders and scope

- **Stakeholders:** repository operators, team authors, CI maintainers, and MCP
  tool providers.
- **Actors:** the operator, enabled workers, the local environment, and the miez
  CLI.
- **In scope:** MCP declarations, worker references, placeholder resolution,
  interactive prompting, non-interactive failure, and local persistence.
- **Out of scope:** starting MCP servers, validating provider credentials with a
  remote service, and encrypting the local store.

## 3. Non-goals

- miez does not expose supplied secret values in command output.
- miez does not write supplied values into portable team or installation files.
- miez does not leave unresolved required values for a downstream process to
  discover later.

## 4. Requirements

### 4.1 Functional

#### `mcp-secrets/worker-tool-dependencies-are-declared-in-team-yaml`

When a worker declares an MCP tool dependency, the system shall require that
the dependency resolves to a tool declared by the owning team.

Acceptance:

- WHEN every worker tool reference resolves to a declared MCP server THEN
  validation accepts the team.
- WHEN a worker references an unknown MCP server THEN validation identifies the
  worker and unresolved id.

#### `mcp-secrets/missing-tool-configuration-prompts-interactively`

When an enabled worker requires an unresolved MCP environment value, the system
shall prompt for the value in an interactive terminal or fail clearly in a
non-interactive or CI session.

Acceptance:

- WHEN the value resolves from local overrides or the process environment THEN
  no prompt is shown.
- WHEN a value is missing in an interactive terminal THEN the operator is
  prompted and sensitive-looking names are masked.
- WHEN a value is missing outside an interactive terminal or under CI THEN the
  operation fails without hanging or installing the team.

#### `mcp-secrets/tool-configuration-values-stay-local-and-uncommitted`

When an operator supplies an MCP configuration value, the system shall store it
only in local workspace configuration and shall exclude it from portable team
source and installation records.

Acceptance:

- WHEN a later operation needs the same variable THEN the local value is
  reused.
- WHEN committed and installed team artifacts are inspected THEN the supplied
  value is absent from them.

## 5. Open questions

- The local MCP store is protected by filesystem permissions but is not an
  encrypted secret manager.
