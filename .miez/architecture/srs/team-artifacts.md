# SRS: team artifacts

ISO/IEC/IEEE 29148 software requirements specification using EARS.

**Status:** approved package contract including task artifacts, drafted
2026-09-16.

## 1. Introduction

This capability defines the authored and generated parts of the portable team
bundle that miez can build, load, validate, and bootstrap. A team contains four
distinct domain artifact roles: workers describe who does the work, skills
describe how to solve a topic, tasks describe what to do, and workflows
describe how to chain workers and tasks. This document covers the package
contract; task behavior and invocation are specified in
[task-artifacts.md](task-artifacts.md).

It exists so team authors can keep configuration beside the Markdown artifact
it describes while operators receive one deterministic team index for
installation and CLI use.

## 2. Stakeholders and scope

- **Stakeholders:** team authors, repository operators, and CI maintainers.
- **Actors:** the operator, a team author, and the miez validator.
- **In scope:** the package metadata manifest, Markdown frontmatter, generated
  team index, worker/skill/task/workflow artifacts, validation, and starter-team
  authoring.
- **Out of scope:** task-specific invocation behavior, provider-specific
  rendering details, remote download mechanics, and execution of workers or
  workflows.

## 3. Non-goals

- The capability does not prescribe how a team author writes prompt content.
- The capability does not infer undocumented behavior from Markdown bodies;
  only declared frontmatter is compiled into the team index.
- The capability does not publish releases or create Git tags.
- The capability does not execute or evaluate the quality of a worker response.

## 4. Requirements

### 4.1 Functional

#### `artifact/team-contract-is-loadable`

When a generated team artifact is loaded, the system shall accept a manifest
that identifies the team and declares its workers, skills, tasks, workflows,
model catalog, and MCP declarations.

Acceptance:

- WHEN a compatible generated team manifest is loaded THEN its declared
  workers, skills, tasks, and workflows are available for validation,
  completion, and rendering.
- WHEN a manually authored manifest omits generated artifact relationships THEN
  the build operation reports the missing generated data instead of silently
  inferring it during installation.
- WHEN the manifest contains unknown fields or malformed data THEN loading
  fails with an actionable error.

#### `artifact/team-index-is-generated`

When a team is built, the system shall generate its installable team index from
the package metadata and valid artifact frontmatter.

Acceptance:

- WHEN build inputs are valid THEN the generated `miez.generated.yaml` contains the team
  metadata, worker paths and configuration, available skills, task paths and
  configuration, workflow paths and configuration, model catalog, and MCP
  declarations.
- WHEN build inputs are invalid THEN the operation fails without replacing a
  previously generated `miez.generated.yaml`.
- WHEN the same inputs are built repeatedly THEN the generated file is
  byte-equivalent and deterministically ordered.

#### `artifact/model-catalog-uses-stable-ids`

When a team declares model catalog entries, the system shall identify each
entry by one stable id and shall not require a duplicated display-name field.

Acceptance:

- WHEN a model catalog entry contains a valid id without a display `name`
  property THEN the team manifest loads successfully.
- WHEN a model catalog entry contains an unsupported display `name` property
  THEN manifest loading reports it as an unknown field.

#### `artifact/markdown-contract-is-enforced`

When a team is built, the system shall require each worker, skill, task, and
workflow Markdown artifact to have valid frontmatter and a path contained
within the team source. When a generated team is installed, the system shall
validate the declared artifact paths without reconstructing configuration from
frontmatter.

Acceptance:

- WHEN a build input has missing or mismatched frontmatter THEN validation
  reports the artifact and its problem.
- WHEN a generated index references a path that escapes the team source or
  resolves through a symlink THEN installation rejects the package.

#### `artifact/frontmatter-declares-cli-configuration`

When a team artifact is built, the system shall derive each worker, task, and
workflow configuration from its frontmatter without requiring a duplicate
manual entry for that artifact in `miez.generated.yaml`.

Acceptance:

- WHEN a worker declares optional miez fields `skills`, `model`, or MCP `tools`
  THEN the generated worker entry contains those values and its source path.
- WHEN a worker contains provider frontmatter fields THEN the build accepts
  those fields without requiring miez to duplicate the provider schema.
- WHEN a workflow declares its id, display name, and ordered worker phases
  THEN the generated workflow entry contains those values and its source path.
- WHEN a task declares its description THEN the generated task entry contains
  those values and its source path without binding the task to a worker.
- WHEN a skill declares its id THEN the generated skill catalog contains its
  fixed source path.
- WHEN an artifact has duplicate or conflicting metadata THEN the build fails
  and identifies the artifact and field.

#### `artifact/worker-frontmatter-is-provider-compatible`

When a worker is authored for the supported Copilot target, the system shall
allow provider frontmatter to evolve independently from miez-owned worker
configuration.

Acceptance:

- WHEN a worker omits `kind` THEN the build succeeds and the generated worker
  catalog omits that obsolete field.
- WHEN a worker contains provider fields such as `name`, `description`,
  `model`, or `reasoning-effort` THEN those fields remain available to the
  renderer.
- WHEN a legacy worker declares `kind: agent` THEN the package remains
  readable during migration, but the field is not generated into new indexes.
- WHEN a legacy worker declares another kind THEN the build fails clearly.

#### `artifact/workflow-artifacts-are-discoverable`

When a team is built or loaded, the system shall expose every valid workflow
declared by the generated team index as a selectable workflow with a stable
identifier.

Acceptance:

- WHEN a generated team index contains one or more valid workflows THEN
  validation exposes their identifiers, source paths, and phase configuration
  in deterministic order.
- WHEN a workflow frontmatter has an invalid identifier, unsupported field, or
  duplicate identifier THEN build fails and identifies the artifact.
- WHEN a team source contains no valid workflow artifacts THEN build fails
  because an active team must always have one selectable workflow.

#### `check/validates-team-artifacts`

When an operator builds or checks a team, the system shall validate its package
  metadata, generated index, identifiers, worker/skill/task/workflow
  references, models, MCP relationships, and referenced files.

Acceptance:

- WHEN all team relationships and artifacts are valid THEN build or check
  succeeds and identifies the team as valid.
- WHEN an identifier, worker, workflow artifact, model, skill, MCP reference,
  generated index, or file is invalid THEN the operation fails and identifies
  the relevant path.

#### `team-authoring/bootstrap-creates-compatible-skeleton`

When an operator bootstraps a team identifier, the system shall create a
package metadata manifest and four artifact directories containing one worker,
one skill, one task, and one workflow template without overwriting an existing
team.

Acceptance:

- WHEN the destination does not contain the team THEN `miez.yaml`,
  `workflows/`, `workers/`, `tasks/`, and `skills/` with one frontmatter
  template each are created in a flat team directory, and no generated
  `miez.generated.yaml` is created before build.
- WHEN the identifier is invalid or the destination already contains a team
  THEN the command fails without overwriting existing authoring files.

#### `team-authoring/bootstrap-includes-guided-authoring-support`

When an operator bootstraps a team, the system shall provide a self-contained
authoring guide and Copilot-compatible helpers for creating and maintaining
valid team artifacts.

Acceptance:

- WHEN bootstrap completes THEN the new team contains `README.md`, an
  always-on authoring instruction, user-invocable worker and skill prompts, a
  facilitator agent, and reusable authoring skills.
- WHEN a team author reads the guide THEN it explains the CLI lifecycle,
  worker-agent contract, provider and miez-owned frontmatter, skill assignment, task
  invocation, workflow membership, and the distinction between authored files
  and the generated team index.

#### `team-authoring/bootstrap-copies-editor-support`

When an operator bootstraps a team, the system shall include the package-local
VS Code settings and snippets needed to author the Markdown artifacts.

Acceptance:

- WHEN bootstrap completes THEN `.vscode/settings.json` and
  `.vscode/miez-team.code-snippets` exist in the new team directory.
- WHEN a worker is opened in VS Code THEN the settings associate it with the
  built-in agent language and the snippets expose the worker frontmatter
  templates.
- WHEN the package is installed or audited THEN these regular support files
  remain part of the source module and lockfile inventory.

#### `team-authoring/support-files-are-not-operational-artifacts`

When a team package contains authoring support files, the system shall keep
those files outside the operational worker, skill, and workflow catalog.

Acceptance:

- WHEN a package containing the support bundle is built THEN the generated
  team index contains no worker, skill, task, or workflow entry derived only
  from a support file.
- WHEN the same package is built repeatedly THEN support-file presence does
  not make the generated team index nondeterministic.

#### `team-authoring/installed-source-retains-support-files`

When a built team package is installed, the system shall retain its authoring
documentation and helpers in the installed source module.

Acceptance:

- WHEN installation succeeds THEN every support file present in the built
  package is present below the installed team's source module with its content
  unchanged.
- WHEN the team is audited THEN the retained support files participate in the
  installed source file inventory and integrity hashes.

#### `team-authoring/build-produces-installable-package`

When an operator builds a team package, the system shall produce a generated
team index containing workers, skills, tasks, and workflows that can be
installed without parsing authoring frontmatter.

Acceptance:

- WHEN build succeeds THEN `miez.generated.yaml` is present beside `miez.yaml`
  and its worker, skill, task, and workflow references resolve to files in the
  package.
- WHEN the generated package is copied to an installation source THEN the
  installed team can be validated using `miez.generated.yaml` alone for its catalog and
  CLI completion data.

## 5. Open questions

- The artifact contract does not require a particular team-authoring toolchain.
- Bootstrap authoring helpers remain package-local after installation; rendering
  them into a consuming workspace is a separate future capability.
