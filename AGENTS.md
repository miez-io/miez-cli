# Introduction
miez-cli should become a solution that is not yet on the market.

We have on the market APM, Spec Kit, OpenSpec and custom agentic workflows. All approaches lack something. I want to fill the gap.

# Competitors and why the suck
Spec Kit & OpenSpec: Too much manual control and overhead, no automation, huge CLI the user needs to learn
    - SpecKit mentions something in the docs that workflows didnt work for them
    - SpecKit now just few weeks ago release now workflows again: https://github.com/Fission-AI/OpenSpec/blob/main/docs/opsx.md
    - SpecKit workflows is way too heavy and too much unecessary overhead in my opinion
APM: Installeable artifacts, custom agentic workflows, but static. I cannot switch between artifacts or whole agent team, I can't easily configure my workflow otherwise, its a fixed workflow.

# Vision and top requirements

Here is my vision for miez-cli:

- It must be simple
- We focus on delivering a structure, almost pure markdown files, that allows us to ship teams with specialised workers and skills
- Workers can be invoked manually or integrated into defined workflows
- Around the markdown files, we will be a small and minimalistic Go CLI, that allows us to do some deterministic work
- If the CLI becomes more than 5 commands it already failed. It must stay simple with 3-5 commands + some options
- CLI command groups (4 roots, matching the constraint above):
    - team install <github-url> = initialize the repository on first install, then download, validate, and activate a team from a GitHub URL
    - team list/use/bootstrap/build = list installed teams, switch the active team by local id without downloading, scaffold a new team's authoring files, or compile authoring metadata/frontmatter into an installable `miez.generated.yaml`
    - workflow use = pick which of the team's workflows is active; exactly one workflow is always active for the active team and is installed as an always-on GitHub Copilot instruction
    - workflow worker enable/disable <worker-name> = adjust which workers participate in the active workflow without editing the team's files
    - worker skill add/remove <worker-name> <skill-name> = assign or remove a skill from a worker
    - worker model list/set <worker-name> <model-name> = list the models a `kind: agent` worker can use, and switch one, without caring how vscode or cursor spell that model's name internally
    - check = validate that a team's artifacts are miez-CLI compatible
- So the CLI is only for managing Team -> Worker -> Workflow Artifacts and their generated catalog. Artifact bodies are Markdown; package metadata is `miez.yaml`; `miez.generated.yaml` is generated, not manually authored.
- Artifacts can be downloaded from a GitHub repo or initialized locally where miez is installed. A team package is authored with `miez.yaml`, `workflows/`, `workers/`, and `skills/`; `miez team build` generates the installable `miez.generated.yaml`. A downloaded package is placed under `.miez/miez_modules/<team-id>/`, with its lock record at `.miez/miez.lock.yaml`. One repo or branch can host several teams side by side, each selected by a path in the GitHub URL; a repo/branch with no path must resolve to exactly one generated `miez.generated.yaml`.
- Project is fully open source
- README explains only what the project is, how to install and work with it. Some additional Info
- CONTRIBUTING file is for detailed local development instructions and how to collaborate
- CLI is build in Go language
- How the artifacts are build is totally up to the users. The can build whatever artifacts they need. 
    - For example building a spec driven development is task of the user, not a framework like SpecKit. Because not always spec driven development is required. So if we have a team that uses spec-driven development, we would build artifacts that exactly describes this and ships with prompts and scripts that will do all kind of work.
- Workspace-local CLI state, including the active team, active workflow, and local worker overrides, is stored under `.miez`. Team package source remains in its package files and installed module; it is not workspace state.

# Detailed requirements

## CLI commands

### team install

- Installation requires a GitHub team URL. On the first installation, the target is prompted when it is not supplied by a flag; later installations reuse the workspace target. Interrupting a target or MCP prompt (Ctrl+C) must exit cleanly instead of hanging
- GitHub Copilot is the only supported render target; target selection is retained as a future extension point.
- Installation downloads and activates the URL-selected team. There is no predefined/built-in team shipped in the miez binary anymore, and no "blank" start.

### team

- Allows sub-command `list`, to list all installed teams
- Allows sub-command `install <github-url>` to download, validate, install, and activate a remote team. The URL may point at the repo root (default branch), a `tree/<branch>`, a path inside the repo, or both (`tree/<branch>/<path>`).
- Allows sub-command `use <team-id>` to select the active team from installed local modules. It does not contact GitHub or install source; it removes the previous managed artifacts and renders the selected team's artifacts.
- Allows sub-command `bootstrap <team-id>` to scaffold a new team's authoring files (`miez.yaml` plus `workflows/`, `workers/`, and `skills/`, each with one frontmatter template). It does not generate `miez.generated.yaml`; run `team build` after editing the artifacts. Its output is plain, target-agnostic Markdown and package metadata.
- Allows sub-command `build [path]` to validate package metadata and all artifact frontmatter, then deterministically generate an installable `miez.generated.yaml`. A failed build leaves the previous generated index unchanged.

### workflow

- Allows sub-command `use <workflow-name>` for setting the active workflow. The selected workflow is installed as an always-on GitHub Copilot instruction. Workflow identity, path, and ordered phases come from the generated `miez.generated.yaml`, which was compiled from workflow frontmatter.
    - Switching between workflows means also installing new workflow instruction and removing old

#### worker

- Worker is also a sub command of workflows
- Allows sub-command `enable/disable <worker-name>` for removing workers from the workflow
    - A workflow Markdown file declares ordered phases and participating workers in frontmatter; `team build` compiles that metadata into `miez.generated.yaml`
    - If a worker is enabled/disabled this is recorded as local repo state, not a change to the team's own files, so switching teams or updating the team doesn't lose it
    - Removing workers from a workflow must ensure the workflow still works

### worker

- The standalone worker command is for assigning skills and capabilities to the worker, and for `kind: agent` workers, for choosing the underlying model
- Allows sub-command `skill add <worker-name> <skill-name>` for assigning a new skill to the worker
- Allows sub-command `skill remove <worker-name> <skill-name>` for removing a skill from the worker
- A worker's effective skills are rendered as links to the canonical Copilot skill files, with a generic Skills section directing Copilot to read and follow them. This lets the user customize worker capabilities without duplicating skill bodies in every worker artifact.
- A worker's `kind` is either `command` (a manually-invoked prompt/command) or `agent` (a persistent custom agent with its own model). Only `kind: agent` workers have a model
- Allows sub-command `model list` to show which model ids are usable right now (the active team's resolved catalog, or just the built-in ids if no team is active)
- Allows sub-command `model set <worker-name> <model-name>` to switch a `kind: agent` worker to a different model at any time, without editing the team's files. Both arguments are shell-tab-completable
- Workers and teams reference miez's stable model ids, never provider-specific names. The Copilot mapping is resolved when the generated output is rendered, and a missing Copilot mapping fails the build.
- miez ships a small built-in list of model ids, but a team is never limited to it: a team's own config can add ids the moment a provider ships them, or correct/override a built-in id's mapping, so extending the model list never requires a new miez release

### check

- A simple command that tests artifacts if they are correct. It checks if the config schema is correct and not broken. If this test passes, we can certainly be sure that the artifacts are compatible with the CLI

## Artifacts

- Artifacts are organised by teams
- The hierarchy is Teams -> Workers -> Workflows
- The team consists of specialised workers that can be used to build workflows - like a real team, has workers and a manager could use these workers to establish a workflow/process, we use the same paradigm here
- The artifacts must be a combination of determenistic config files that the CLI can work with and markdown file templates that allow for dynamic injections and configuration by the CLI
- How these artifacts are than really build, is up to the users
- The miez bootstrap command ensures users can bootstrap artifacts that are miez compatible
- The miez check command ensures the user can run a self-check if the config schema etc. is correct, and if it is, the user can be sure that its miez cli compatible and it will work
- The authored package contract is `miez.yaml` plus Markdown artifacts. `miez team build` generates one `miez.generated.yaml` per team, containing:
    - `id`, `version`, `name`, `description`
    - `skills`: each with an `id` and generated source path
    - `workers`: each with an `id`, a `kind` (`command` or `agent`), a generated path to its Markdown file, optional `skills`, and for `kind: agent` an optional `model` and MCP tools
    - `workflows`: each with an `id`, `name`, generated source path, and ordered `phases`, each phase listing the worker ids that participate in it
    - optional `default_model` and `models` (a team-declared list of model ids/mappings, extending or overriding the built-in ones)
    - optional MCP declarations and package metadata copied from `miez.yaml`

## Example team (spec-driven-team)

- Decided: miez ships no built-in team in the binary anymore. What used to be "the one predefined team" is now `spec-driven-team`, a normal external GitHub team hosted in the separate `miez-scrum-team` repo - installed like any other via `team install <github-url>`, nothing special about it in the CLI
    - Its authoring files still double as the reference example for what `team bootstrap` scaffolds, and for the shape `miez.generated.yaml` is expected to have
- This team includes a spec-driven approach similar to OpenSpec
- It has specialised workers built as commands/prompts
    - spec-engineer -> for planning and writing detailed specs
        - creates a folder .miez/changes/<change-name>/spec/ for working notes
        - creates .miez/changes/<change-name>/contract.md which defines the change contract and references the living SRS, including operations and intent when/then scenarios
    - architect -> Analyses the code base critically. For each spec, creates technical design decision, asks the user for decisions, document the decisions, documents the technical implementation appraoch with important techncail details
        - create .miez/changes/<change-name>/design.md
        - currently the one worker set up as `kind: agent` in this team (with its own model), since it benefits most from a stronger/dedicated reasoning model - the rest are `kind: command`
    - plan -> Breaks the architects work into clear tasks
        - create .miez/changes/<change-name>/tasks.md
    - user-story -> Creates Jira compatible user-stories out of the tasks.md
    - implement -> Take one task from the tasks.md and start implementing it. If finished
    - reviewer -> Reviews only the implemented changes introduced by tasks and verifies if it meets the architecture/srs/ requirements and design.md
    - finish -> Finishes the task job. Marks the task as done so the state is complete and there is no state drift that leads another worker start the same task again.
- If worklfow is disable, each worker can be invoked one by one. 
- The team defines a configure workflow that would simply chain these workers after another. The user just gives his idea and the agent runs spec to finish.

### Kick-off prompt for v3

I created a v3 repo and copied this AGENTS.md Only update this one from now on. The problem I have is, that the download function does not work well and requires much work. So I actually want to use apm for that. Now there are two ways. Either use apm as it is and built on top our custom cli OR port a minimal version of apm to a Go CLI. I prefer the Go port. So I also added apm repository here in this workspace, which should be the port reference. So the domain model should remain as I considered it. We think in teams, workers, workflows and skills. Team is the root config yaml. Workers can be either commands or agents. Worker can get assigned skills, tools and specific models to run. Workers can be configured into workflows. All this configurable by the CLI as it is today. Now for managing these team installations, versioning, reproducbiality, etc. we want to rely on what apm already built. Here are the specific requirements: we want to port command apm install <github-url> (public but also private repos possible), apm update <github-url>. Both will work as command miez team use <github-url> and miez team update <github-url>. Additionally, miez allows for named teams. So I can configure key value pair like <team-name>: github-url so I can use the team name for miez team use and miez team update commands. We also port apm audit (miez team audit). We want the same functionality from apm, this includes also a modules folder and a lock folder. Also, if one team comes with mcp tools, same like on APM the CLI will ask for additionaly config like API keys, if the MCP requires it. So again, miez-cli is an minimal Go APM port for managing the installation of artifacts. Miez additional CLI commands is for managing and configuring teams dynamically.Also please refer to miez-scrum-team repo as an example module. I want to enforce this project structure. rules folder results into e.g github copilot instructions. workers either result into agents or prompts. skills result into skills, workflows into rules (e.g copilot instructions) that are always on. For now we only support copilot as target. Now please analyse my prompt, my AGENTS.md and the apm workspace, than write detailed spec deliverables according to the skill miez-specs and use the spec-engineer agent. 