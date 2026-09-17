# SRS: Copilot rendering

ISO/IEC/IEEE 29148 software requirements specification using EARS.

**Status:** approved target contract for the workflow artifact change, drafted
2026-09-15.

## 1. Introduction

This capability defines how validated team artifacts become GitHub Copilot
prompts, custom agents, skills, and instructions. It exists to give team
authors a stable provider-facing contract while keeping authored artifacts
provider-neutral where possible.

## 2. Stakeholders and scope

- **Stakeholders:** repository operators, team authors, and GitHub Copilot
  users.
- **Actors:** the miez compiler and GitHub Copilot.
- **In scope:** the supported target, worker and skill mappings, rules,
  workflow routing, and stable model-name resolution.
- **Out of scope:** executing Copilot agents, selecting a model through a
  provider API, and rendering unsupported providers.

## 3. Non-goals

- The compiler does not execute prompts or workflows.
- A team's catalog uses stable model ids where it declares a model mapping;
  worker frontmatter may also retain a provider-native selector for the
  supported Copilot target.
- Rules and workflow routing are not treated as the same kind of authored input.

## 4. Requirements

### 4.1 Functional

#### `render-contract/github-copilot-is-the-only-render-target`

When a workspace is rendered, the system shall support the GitHub Copilot target
and shall reject unsupported rendering targets explicitly.

Acceptance:

- WHEN the configured target is GitHub Copilot THEN the team's workers, skills,
  and instructions are rendered for Copilot.
- WHEN another target is configured or requested THEN rendering fails instead
  of silently ignoring the target.

#### `render-contract/authored-folders-map-to-agents-and-prompts`

When a valid team is rendered, the system shall map workers, tasks, skills, and
the selected workflow artifact to their defined Copilot artifact kinds.

Acceptance:

- WHEN a worker is rendered THEN it appears as a Copilot custom agent.
- WHEN a task is rendered THEN it appears as a user-invocable Copilot prompt.
- WHEN a skill or selected workflow is declared THEN its content appears in the
  corresponding Copilot skill or routing artifact.

#### `render-contract/skills-are-referenced-not-inlined`

When a worker has effective skills, the system shall render a generic Skills
section containing a link to each corresponding Copilot skill artifact instead
of copying the skill body into the worker output.

Acceptance:

- WHEN a worker uses one or more skills THEN its generated prompt or agent
  contains one generic Skills section and one resolvable link per effective
  skill.
- WHEN a worker uses a skill THEN the generated worker does not contain a
  second copy of that skill's Markdown body.
- WHEN a skill is removed through a local override THEN its link is absent from
  the worker while the remaining skill links stay valid.

#### `render-contract/worker-frontmatter-is-preserved`

When a worker is rendered for GitHub Copilot, the system shall preserve the
worker's provider frontmatter while excluding miez-owned configuration fields.

Acceptance:

- WHEN a worker contains provider fields such as `name`, `description`,
  `model`, or `reasoning-effort` THEN the generated `.github/agents/<id>.md`
  contains those fields.
- WHEN miez applies a worker model override or team default THEN the generated
  `model` field contains the effective Copilot selector.
- WHEN a worker contains miez-owned `skills`, MCP `tools`, or legacy `kind`
  fields THEN those fields are not copied into the generated agent header.
- WHEN a provider adds another supported frontmatter field THEN miez copies it
  without a renderer allowlist change.

#### `render-contract/rules-are-a-separate-always-on-instruction-from-workflow-routing`

When a team provides general rules and workflow content, the system shall
render the rules independently from the selected workflow instruction.

Acceptance:

- WHEN rules exist and a workflow is selected THEN the rules instruction and
  workflow instruction remain separate managed artifacts.
- WHEN the selected workflow changes THEN only the managed workflow instruction
  is replaced; the rules instruction remains unchanged.

#### `render-contract/selected-workflow-is-always-on`

When a team has an active selected workflow, the system shall render that
workflow, using the path and ordered phase metadata from the generated team
index, as an always-on Copilot instruction.

Acceptance:

- WHEN a team is active THEN exactly one workflow instruction is present and
  applies to all Copilot requests.
- WHEN a workflow is changed THEN the previous workflow instruction is removed
  before the new selected workflow instruction becomes active.

#### `model/stable-identifiers-map-to-copilot-names`

When an agent worker has an effective model identifier, the system shall resolve
that stable identifier to the configured Copilot model name during rendering.

Acceptance:

- WHEN a team extends or overrides its model catalog with a Copilot mapping
  THEN model listing and rendering use the effective mapping.
- WHEN an agent's effective model has no Copilot equivalent THEN rendering
  fails clearly.

#### `model/model-ids-drive-cli-selection`

When the model CLI exposes or completes model choices, the system shall use the
stable ids from the effective model catalog as the selectable values.

Acceptance:

- WHEN an operator lists usable models THEN the output contains each effective
  model id without requiring a second display name.
- WHEN an operator completes the model argument of a worker model selection
  command THEN completion suggests the effective model ids, including team
  additions and overrides.
- WHEN an operator selects a listed model id THEN validation, local state, and
  Copilot rendering use that same id.

## 5. Open questions

- Additional rendering targets may be added later through a separate capability
  decision and contract.
