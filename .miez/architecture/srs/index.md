# miez-cli SRS index

This directory contains the living software requirements specifications for
miez-cli. The documents are grouped by externally observable capability,
not by Go package, feature branch, or implementation file.

Requirements describe the current product contract or an approved target
contract for a planned feature. Stable requirement IDs are unique across this
directory and remain unchanged when a requirement moves or its wording is
refined.

## Structure rule

A new feature normally updates one or more existing capability SRS documents.
Create a new SRS document only when the feature introduces a separate,
consumer-visible capability that would still make sense if surrounding
behavior were replaced. Use this index to keep the capability set coherent.

Use the other architecture documents for different questions:

- `arc42.md` describes system context, containers, runtime, and deployment.
- `srs/` describes observable behavior and acceptance conditions.
- `sdd/` describes the structure and runtime design of existing pieces.
- `adr/` records why a meaningful architectural alternative was chosen.
- Worker- or skill-owned `.miez/changes/<change-id>/` areas may hold temporary
  proposal, contract, design, and task records for work in progress; miez does
  not create them.

## Capability documents

| Document | Capability boundary | Primary design |
|---|---|---|
| [team-artifacts.md](team-artifacts.md) | Team manifests, Markdown artifacts, validation, and bootstrap. | [Artifact and rendering SDD](../sdd/sdd-artifact-and-rendering.md) |
| [task-artifacts.md](task-artifacts.md) | Reusable task objectives, task prompts, worker/task composition, and workflow task assignments. | [Task artifacts and invocation SDD](../sdd/sdd-task-artifacts-and-invocation.md) |
| [workspace-configuration.md](workspace-configuration.md) | Initialization, active-team state, workflow controls, worker overrides, and CLI behavior. | [Workspace and distribution SDD](../sdd/sdd-workspace-and-distribution.md) |
| [team-distribution.md](team-distribution.md) | GitHub installation, source modules, lockfiles, updates, credentials, and remote references. | [Workspace and distribution SDD](../sdd/sdd-workspace-and-distribution.md) |
| [team-audit.md](team-audit.md) | Offline source drift detection and CI gating. | [Workspace and distribution SDD](../sdd/sdd-workspace-and-distribution.md) |
| [mcp-secrets.md](mcp-secrets.md) | MCP declarations, placeholder resolution, prompting, and local values. | [Workspace and distribution SDD](../sdd/sdd-workspace-and-distribution.md) |
| [copilot-rendering.md](copilot-rendering.md) | Copilot target mapping, rules, workflow instructions, and model-name resolution. | [Artifact and rendering SDD](../sdd/sdd-artifact-and-rendering.md) |
| [system-quality.md](system-quality.md) | Cross-cutting reliability, determinism, and remote/filesystem safety. | [Artifact and rendering SDD](../sdd/sdd-artifact-and-rendering.md) and [Workspace and distribution SDD](../sdd/sdd-workspace-and-distribution.md) |

## Requirement registry

### Team artifacts

- `artifact/team-contract-is-loadable`
- `artifact/team-index-is-generated`
- `artifact/model-catalog-uses-stable-ids`
- `artifact/markdown-contract-is-enforced`
- `artifact/frontmatter-declares-cli-configuration`
- `artifact/workflow-artifacts-are-discoverable`
- `check/validates-team-artifacts`
- `team-authoring/bootstrap-creates-compatible-skeleton`
- `team-authoring/bootstrap-includes-guided-authoring-support`
- `team-authoring/support-files-are-not-operational-artifacts`
- `team-authoring/installed-source-retains-support-files`
- `team-authoring/build-produces-installable-package`

### Task artifacts and worker invocation

- `task/domain-roles-are-distinct`
- `worker/all-workers-render-as-agents`
- `task/is-a-reusable-work-objective`
- `task/is-discoverable`
- `task/catalog-includes-task-relationships`
- `task/renders-as-a-copilot-prompt`
- `task/selected-agent-context-is-preserved`
- `task/user-request-is-task-context`
- `workflow/agent-delegation-is-explicit`
- `task/prompt-composition-is-not-nested`
- `task/workflow-task-assignments-are-valid`
- `task/prompt-identifiers-are-unique`
- `task/rendering-is-deterministic`
- `task/rendering-is-traceable`
- `task/copilot-is-the-only-render-target`
- `task/miez-does-not-execute-tasks`
- `task/agent-invocation-is-select-then-prompt`
- `task/agent-invocation-is-the-composition-boundary`

### Workspace configuration

- `init/remote-team-is-required`
- `workflow/selection-and-routing`
- `workflow/selection-replaces-the-active-instruction`
- `workflow/completion-uses-the-installed-catalog`
- `workflow/worker-toggles-preserve-a-usable-workflow`
- `workspace/active-selection-is-local-state`
- `workspace/initialization-creates-only-required-miez-state`
- `workspace/managed-output-has-no-persistent-compiler-ledger`
- `worker/skill-overrides-are-local`
- `worker/agent-model-overrides-are-local`
- `compatibility/one-active-team-per-workspace`
- `cli/command-surface-is-stable`
- `cli/cancellation-is-clean`

### Team distribution

- `team-distribution/source-modules-are-isolated-from-rendered-output`
- `team-distribution/miez-owned-distribution-state`
- `team-distribution/legacy-distribution-state-is-migrated`
- `team-distribution/lockfile-records-reproducible-installs`
- `team-distribution/team-use-installs-from-any-visibility-repository`
- `team-distribution/install-exposes-artifact-catalog`
- `team-distribution/install-uses-generated-team-index`
- `team-distribution/credential-resolution-order-mirrors-host-conventions`
- `team-distribution/team-update-refreshes-an-installed-team`
- `team-distribution/update-uses-active-team-by-default`
- `team-distribution/named-team-references`
- `team-distribution/outdated-reports-available-updates-without-changing-anything`
- `compatibility/github-is-the-distribution-host`
- `security/credentials-are-not-persisted`

### Team audit

- `team-audit/detects-drift-against-the-lockfile`
- `team-audit/ci-gate-mode`

### MCP secrets

- `mcp-secrets/worker-tool-dependencies-are-declared-in-team-yaml`
- `mcp-secrets/missing-tool-configuration-prompts-interactively`
- `mcp-secrets/tool-configuration-values-stay-local-and-uncommitted`

### Copilot rendering

- `render-contract/github-copilot-is-the-only-render-target`
- `render-contract/authored-folders-map-to-agents-and-prompts`
- `render-contract/skills-are-referenced-not-inlined`
- `render-contract/rules-are-a-separate-always-on-instruction-from-workflow-routing`
- `render-contract/selected-workflow-is-always-on`
- `model/stable-identifiers-map-to-copilot-names`
- `model/model-ids-drive-cli-selection`

### System quality

- `reliability/mutations-are-transactional`
- `determinism/local-rendering-is-repeatable`
- `security/remote-input-is-bounded-and-path-safe`
