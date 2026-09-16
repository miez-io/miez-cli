# SRS: workspace configuration

ISO/IEC/IEEE 29148 software requirements specification using EARS.

**Status:** approved target contract for the workflow artifact change, drafted
2026-09-15.

## 1. Introduction

This capability defines how an operator initializes a repository and changes
workspace-local team choices without modifying authored team files. It exists
to keep one active team and exactly one active workflow instruction, alongside
local worker, skill, and model configuration.

## 2. Stakeholders and scope

- **Stakeholders:** repository operators, CI users, team authors, and GitHub
  Copilot users.
- **Actors:** the operator and the miez CLI.
- **In scope:** initialization, active-team state, workflow selection and
  instruction replacement, workflow completion, worker capability overrides,
  model overrides, command surface, and cancellation.
- **Out of scope:** remote source retrieval, team schema validation details,
  and Copilot execution.

## 3. Non-goals

- Workspace-local choices do not rewrite a remote team's authored files.
- The workspace does not run workers or make workflow decisions for Copilot.
- The workspace does not support multiple simultaneously active rendered teams.

## 4. Requirements

### 4.1 Functional

#### `init/remote-team-is-required`

When a repository is initialized, the system shall require a GitHub team
reference and a supported rendering target instead of selecting an embedded or
blank team.

Acceptance:

- WHEN initialization completes successfully THEN the workspace records one
  active installed team and the selected target is supported.
- WHEN the team reference or target is missing or unsupported THEN
  initialization fails without leaving a partial workspace.

#### `workflow/selection-and-routing`

When an active team is loaded, the system shall use the generated team index to
maintain exactly one selected workflow artifact as the workspace's active
routing instruction.

Acceptance:

- WHEN a team is activated THEN one valid workflow is selected and its
  instruction is installed for Copilot.
- WHEN no valid workflow can be selected THEN activation fails without
  changing the previous active team or instruction.

#### `workspace/active-selection-is-local-state`

The system shall persist active team and workflow selections as workspace-local
state without modifying an installed team's generated index or package metadata.

Acceptance:

- WHEN an operator switches the active team or workflow THEN the installed
  `miez.generated.yaml` and package `miez.yaml` remain byte-equivalent.
- WHEN a workspace is reopened THEN its active team and workflow selections are
  restored from local state and resolved against the installed team index.

#### `workspace/initialization-creates-only-required-miez-state`

When miez initializes or prepares a workspace, the system shall create only
state required for team distribution, active selections, and configured local
secrets.

Acceptance:

- WHEN initialization succeeds THEN it does not create `.miez/changes/`,
  `.miez/runs/`, `.miez/manifest.json`, or `.miez/compiled.json`.
- WHEN a worker or skill declares a work area THEN that area is created by the
  artifact that owns it rather than by generic workspace initialization.

#### `workspace/managed-output-has-no-persistent-compiler-ledger`

When miez renders Copilot output, the system shall manage replacement and
cleanup without requiring a persistent compiler manifest or compilation
summary in the workspace.

Acceptance:

- WHEN rendering succeeds THEN no new `.miez/manifest.json` or
  `.miez/compiled.json` is written.
- WHEN a team or workflow is replaced THEN previously managed Copilot files
  are removed or restored transactionally without those records.

#### `workflow/selection-replaces-the-active-instruction`

When an operator selects a different workflow for the active team, the system
shall remove the previously managed workflow instruction and install the
selected workflow instruction as one atomic workspace mutation.

Acceptance:

- WHEN a valid workflow is selected THEN the old workflow instruction is absent
  and the selected workflow instruction is present.
- WHEN selection or rendering fails THEN the previous workflow instruction and
  workspace state remain effective.

#### `workflow/worker-toggles-preserve-a-usable-workflow`

When an operator changes participation for a worker in the selected workflow,
the system shall persist the choice locally and shall reject a change that
would leave the workflow with no enabled worker.

Acceptance:

- WHEN a participating worker is disabled THEN the authored Markdown and
  generated team index remain unchanged and the rendered workflow contains the
  remaining workers.
- WHEN disabling the last active worker is requested THEN the operation fails
  and the prior workspace state remains effective.

#### `workflow/completion-uses-the-installed-catalog`

When an operator requests completion for workflow selection, the system shall
suggest the valid workflow identifiers declared by the active installed team's
generated team index.

Acceptance:

- WHEN an active team has multiple workflows THEN completion returns all of
  their identifiers in deterministic order.
- WHEN no active team is installed THEN completion returns no workflow
  candidates and does not access a remote repository.

#### `worker/skill-overrides-are-local`

When an operator adds or removes a worker skill, the system shall persist the
override locally without modifying the authored team.

Acceptance:

- WHEN a valid skill is added or removed THEN the next rendering uses the
  resulting effective skill set.
- WHEN an unknown skill is named THEN the operation fails without changing
  team source files.

#### `worker/agent-model-overrides-are-local`

When an operator changes the model for an agent worker, the system shall
persist the selected stable model identifier locally without editing the team
manifest.

Acceptance:

- WHEN a supported model is selected for an agent worker THEN subsequent
  rendering uses that effective model.
- WHEN a command worker or unknown model is selected THEN the operation fails
  and no override is applied.

#### `compatibility/one-active-team-per-workspace`

The system shall maintain one active team in a workspace while allowing other
installed teams to remain available for later activation.

Acceptance:

- WHEN a locally installed team is activated THEN only that team is the
  workspace's active rendering source.
- WHEN switching teams succeeds THEN the workspace records the new active team
  and the previous team's output is no longer active.

#### `cli/command-surface-is-stable`

The system shall expose the documented root command groups for initialization,
team lifecycle, workflow configuration, worker configuration, and validation.

Acceptance:

- WHEN the CLI help is requested THEN the five root command groups are listed.
- WHEN a valid subcommand is invoked THEN it performs its documented operation
  and returns an appropriate failure code when the operation cannot complete.
- WHEN workflow configuration help is requested THEN workflow selection and
  workflow-id completion are exposed without phase or worker-participation
  toggle commands.
- WHEN team authoring help is requested THEN a team build operation is exposed
  for generating and validating an installable team index.

#### `cli/cancellation-is-clean`

When an interactive operation is canceled while waiting for input, the system
shall stop waiting and return a conventional interrupt result without leaving
partial initialization state.

Acceptance:

- WHEN an initialization prompt is canceled THEN the process exits with the
  interrupt code and does not hang.
- WHEN cancellation occurs during a mutating operation THEN staged state is
  removed or rolled back.

## 5. Open questions

- The workspace currently has no process-level lock for concurrent mutations.
