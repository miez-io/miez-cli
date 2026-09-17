# miez-cli

## Domain model

miez-cli organizes an agent team around four distinct artifact roles:

- **Workers** are personas: durable identity, motivation, goals, beliefs,
  judgment, and boundaries. Every worker resolves to a selectable GitHub
  Copilot custom agent.
- **Skills** describe how to solve something: reusable, worker-independent
  know-how, procedures, and guardrails assigned to workers.
- **Tasks** describe what to do: partial reusable todos or worker-neutral work
  objectives that can be applied with different workers and skill sets.
- **Workflows** describe how to chain work: ordered coordination of workers and
  task assignments for more automation. miez renders the coordination; GitHub
  Copilot executes it.

Tasks are first-class reusable objectives. Their requirements and design are
documented in [.miez/architecture/srs/task-artifacts.md](.miez/architecture/srs/task-artifacts.md)
and [.miez/architecture/sdd/sdd-task-artifacts-and-invocation.md](.miez/architecture/sdd/sdd-task-artifacts-and-invocation.md).

miez-cli is a small Go CLI for installing configurable teams of Markdown
workers. Teams contain specialized workers, reusable skills, tasks, and
workflow routing. miez performs deterministic setup, validation, state changes,
distribution (install/update/audit), and compilation; GitHub Copilot remains
responsible for actually running the workflow.

GitHub Copilot is the only render target for now, and installing a team is a
dependency-manager flow with reproducible installs, updates, drift audits,
and private repositories.

## Install and run

Build the CLI locally:

```bash
go build -o bin/miez ./cmd/miez
./bin/miez --help
```

Install a GitHub team into a repository:

```bash
miez team install https://github.com/owner/team-repository --targets copilot
miez check
```

The GitHub URL is required. On the first installation, omit `--targets` to be
prompted for the render target; later installations reuse the workspace's
persisted target configuration. miez ships no embedded team and does not
support a blank, team-less installation.

## Command surface

```text
miez team install|list|use|bootstrap|build|update|outdated|audit
miez workflow use|worker enable|worker disable
miez worker skill add|remove
miez worker model list|set
miez check
```

`miez team install <github-url>` downloads, validates, and activates a team.
One repository can host several teams, each in its own subdirectory, selected
by adding its path to the URL:

```bash
miez team install https://github.com/owner/repository/teams/my-team
miez team install https://github.com/owner/repository/tree/develop/teams/my-team
```

Without a path, the repository or branch must contain exactly one generated
`miez.generated.yaml` beside its package `miez.yaml`.

After multiple teams are installed, switch between them locally without a
network request:

```bash
miez team use my-team
```

`team use` accepts only an installed team id. It cleans and renders the
selected team's managed artifacts while retaining installed team modules.
Remote installation works against private repositories as long as a credential
resolves. The raw downloaded source lands under
`.miez/miez_modules/<team-id>/`, and `.miez/miez.lock.yaml` records the
resolved commit, version, and hash of every deployed file. The previous active team remains usable if download,
validation, compilation, or state persistence fails.

Once a team is installed, keep it current or audit it without reinstalling:

```bash
miez team update                  # update the active team, with confirmation
miez team update my-team --dry-run
miez team update my-team --yes    # apply without prompting
miez team outdated                # report available updates
miez team audit                   # compare files with .miez/miez.lock.yaml
miez team audit --ci              # exit non-zero on drift
```

Exactly one workflow is active for the active team. `miez workflow use
<workflow-id>` replaces the managed Copilot instruction. `miez workflow worker
enable`/`disable` change local worker participation without modifying the
package. Rules remain a separate always-on instruction. Worker skill and model
commands also change only workspace-local state.

## Authentication

GitHub credentials resolve in a fixed priority chain:

1. `GITHUB_MIEZ_PAT_<ORG>`
2. `GITHUB_MIEZ_PAT`
3. `GITHUB_TOKEN`
4. `GH_TOKEN`
5. `gh auth token --hostname github.com`

No resolved credential is written to `miez.yaml`, `.miez/miez.lock.yaml`, or
any file under `.miez/miez_modules/`. A public repository needs no credential.

## Authoring and artifact contract

Team authors work from a package directory containing `miez.yaml` and four
operational artifact folders. `miez.generated.yaml` is generated; do not
hand-edit it:

```text
my-team/
  miez.yaml
  workers/*.md
  skills/<skill-id>/SKILL.md
  tasks/*.md
  workflows/*.md
```

Tasks are rendered as worker-neutral Copilot prompts. The user
selects a worker agent and invokes the task in that agent's context. Workflows
that coordinate multiple workers use provider-supported agent delegation or
handoffs; miez does not rely on nested slash-prompt invocation.

The package manifest owns team metadata and shared declarations:

```yaml
id: my-team
version: 1.0.0
name: My team
description: A concise team description.
author: Team author
default_model: gpt-5
models:
  - id: gpt-5
    copilot: GPT-5 (copilot)
mcp:
  - id: github
    env:
      GITHUB_TOKEN: "${env:GITHUB_TOKEN}"
```

Worker and workflow configuration lives in Markdown frontmatter. The artifact
id is the file name (`workers/architect.md` becomes `architect`), never a
frontmatter field:

```yaml
---
name: Architect
description: Designs system architecture.
model: claude-sonnet-4.5
reasoning-effort: xhigh
skills: [architecture]
tools: [github]
---
Describe the worker persona and behavior here.
```

Workers always render as GitHub Copilot custom agents, so no `kind` field is
required. Miez consumes `skills` and MCP `tools` as package configuration,
applies the effective `model`, and copies the remaining provider frontmatter
fields into `.github/agents/<id>.md`. This preserves fields such as `name`,
`description`, and `reasoning-effort` without requiring miez to maintain an
allowlist of provider headers. A legacy `kind: agent` field is accepted while
older packages are migrated but is omitted from new generated indexes.

Bootstrapped packages include `.vscode/settings.json`, which associates miez
artifact files with VS Code's built-in prompt languages. The `chatagent`
language provides agent-file syntax and highlighting; it does not provide a
miez frontmatter schema or live model completion for arbitrary files under
`workers/`.

The accompanying `.vscode/miez-team.code-snippets` provides completion snippets
for the miez worker fields. Open the folder containing `miez.yaml` as a VS Code
workspace folder, ensure the file's language mode is `Agent`, and reload the
window after bootstrapping a team if the snippets do not appear. Type
`miez-worker` for a complete header or a field name such as `skills` and press
Tab/Enter to insert its snippet.

The `model` field accepts a miez model id or a native Copilot selector such as
`GPT-5.6 Luna (copilot)`. Miez validates that the selector is non-empty; it
does not know which model names a particular VS Code installation exposes.
`miez team build` remains authoritative for skill and MCP ids.

```yaml
---
name: Default
phases:
  - id: implement
    workers: [architect]
---
Describe how the workflow coordinates the workers here.
```

Build and validate the installable catalog before publishing the package:

```bash
miez team bootstrap my-team
miez team build my-team
```

`miez team build` validates package metadata and all artifact frontmatter,
then atomically generates `miez.generated.yaml`. A failed build leaves the previous
generated index unchanged. Once installed, miez uses `miez.generated.yaml` for workflow,
worker, skill, task, model, and MCP catalog information without rediscovering
authoring frontmatter.

`miez team bootstrap` also creates a README, package-local VS Code settings and
snippets under `.vscode/`, and a Copilot authoring support bundle under
`.github/`: a package-contract instruction,
prompts for adding a worker, skill, task, or workflow, a `write-workers` skill
carrying the strict worker template, and a `review-team-package` skill with a
build-error reference. These support files help authors maintain valid team
source but are not operational entries in `miez.generated.yaml` and are not
rendered into the consuming workspace by the current compiler.

An installed package contains both files. The generated catalog begins with a
`Generated by miez team build. DO NOT EDIT.` comment:

```yaml
id: my-team
version: 1.0.0
name: My team
author: Team author
default_model: gpt-5
skills:
  - id: architecture
    path: skills/architecture/SKILL.md
workers:
  - id: architect
    path: workers/architect.md
    model: claude-sonnet-4.5
    skills: [architecture]
tasks:
  - id: analyze-existing-solution
    path: tasks/analyze-existing-solution.md
    description: Analyze an existing solution and document its requirements.
workflows:
  - id: default
    name: Default
    path: workflows/default.md
    phases:
      - id: implement
        workers: [architect]
```

Every worker renders as an agent and may have a model. A worker model overrides
the team's `default_model`, which falls back to a built-in default. Team-
declared models use stable ids and a `copilot` mapping; the compiler resolves
the effective selector and retains the other provider frontmatter fields.

Skills may contain supporting assets and references. The package compiler uses
the canonical `skills/<skill-id>/SKILL.md` file as the skill catalog entry and
renders links to the corresponding `.github/skills/<skill-id>/SKILL.md` files
from each worker instead of copying skill bodies into worker output.

### MCP tools

A worker can declare `tools:` referencing an id in the package's top-level
`mcp:` list. Missing values are resolved from the environment or prompted for
interactively, then stored only in the ignored `.miez/mcp.local.yaml`.

## Repository state

- `.miez/architecture/` is the committed architecture documentation tree.
- `.miez/config.yaml` stores active team, active workflow, and local overrides.
- `.miez/mcp.local.yaml` contains ignored local MCP values.
- A team package `miez.yaml` stores package metadata. Installation does not
  create a separate consuming-workspace alias manifest.
- `.miez/miez.lock.yaml` records resolved team revisions and source hashes.
- `.miez/miez_modules/<team-id>/` contains raw downloaded team source.
- `.miez/changes/` and `.miez/runs/` are optional worker- or skill-owned work
  areas; miez does not create them during initialization.

miez does not execute Copilot workflows, call an LLM, or require an API key.
Local validation works offline; `team outdated` and `team audit` remain
read-only operations.

APM compatibility is contract-based. A repository may be both an APM package
and a miez team, but it must include a generated `miez.generated.yaml` and package
`miez.yaml` to define the miez catalog. miez does not invoke the external APM
executable or infer workflow phases from arbitrary APM primitives.

## Development

```bash
go build ./...
go test ./...
go vet ./...
gofmt -l .
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for repository conventions and
artifact change guidance.
