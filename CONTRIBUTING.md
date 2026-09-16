# Contributing to miez-cli

## Prerequisites

- Go 1.26 or newer
- Git
- `gh` CLI (optional) if you want the last tier of the credential chain to
  work locally
- `golangci-lint` for the optional lint pass (config at `.golangci.yml`;
  not currently run in CI for this repository)

## Repository shape

`cmd/miez` contains the small executable entry point. Deterministic behavior
lives under `internal/`, one package per concern:

- `model` -- Team/Worker/Workflow/Skill/MCPServer schema and workspace state shape.
- `artifacts` -- generated `miez.generated.yaml` loading, Markdown parsing, directory copy.
- `build` -- package `miez.yaml` and Markdown frontmatter compilation into `miez.generated.yaml`.
- `workspace` -- `.miez/config.yaml` local per-repository state.
- `manifest` -- typed team-package and workspace `miez.yaml` metadata.
- `lockfile` -- the `.miez/miez.lock.yaml` reproducible-install record.
- `modules` -- `.miez/miez_modules/<team-id>/` path helpers.
- `credentials` -- the GitHub credential-resolution chain.
- `ghclient` -- the GitHub API client (ref resolution, tarball download).
- `secrets` -- MCP env placeholder resolution and the local secrets store.
- `compile` / `validate` -- rendering to Copilot and team artifact checks.
- `team` -- the `Service` that composes all of the above into
  build/install/update/outdated/audit/bootstrap.
- `cli` -- the `cobra` command tree; owns prompting and output formatting
  and delegates every domain operation to `team.Service`.

miez ships no embedded team. `internal/cli/testdata/spec-driven-team` is a
fixture used only by the CLI package's own tests, standing in for a real
`--github` repository via a fake GitHub API server
(`internal/cli/fixture_test.go`).

Repository-local miez state is intentionally separate from authored
artifacts and from the distribution mechanism:

- `.miez/architecture/` is the living architecture documentation tree; its
  `srs/` directory contains one SRS per capability.
- `.miez/config.yaml` stores active team, workflow, target, and local override
  state.
- `.miez/mcp.local.yaml` holds ignored, locally supplied MCP secret values.
- `.miez/miez.lock.yaml` is the committed reproducible-install record.
- `.miez/miez_modules/<team-id>/` contains raw downloaded team source,
  gitignored.
- `.miez/changes/` and `.miez/runs/` are optional work areas owned by the
  installed workers or skills; miez does not create them generically.

Do not put runtime handoffs in a change record or copy a second specification
tree into one. Contracts reference stable requirement IDs in
`.miez/architecture/srs/` and its index.

## Development

Run the CLI from the repository root:

```bash
go run ./cmd/miez --help
go run ./cmd/miez team build ./path/to/team
go run ./cmd/miez team install https://github.com/owner/team-repository --targets copilot
go run ./cmd/miez check
```

Build it with:

```bash
go build -o bin/miez ./cmd/miez
./bin/miez --help
```

The application roots are exactly `team`, `workflow`, `worker`, and `check`.
`team` owns `install`, `list`, `use`, `bootstrap`, `build`, `update`,
`outdated`, and `audit`; `team install` accepts a GitHub URL while `team use`
accepts only an installed team id. Workflow execution remains Copilot-native
and miez never requires an API key or provider CLI.

## Tests and checks

Use focused package tests while iterating, then run the complete gate:

```bash
go build ./...
go test ./...
go vet ./...
gofmt -l .   # must print nothing
```

Network-facing packages (`ghclient`, `credentials`, `team`) are tested
against fake `httptest.Server`s and injected environment/CLI lookups --
never against real GitHub -- so the suite runs offline and deterministically.
When adding a test that exercises `team.Service`, inject `GH`, `Credentials`,
`Interactive`, and `PromptSecret` rather than relying on the real network,
process environment, or a real terminal.

CLI tests use temporary repositories and captured I/O. Do not add generated
Copilot output or test state to the repository root. Test behavior through
the public command or package boundary rather than implementation details.

## Artifact changes

When adding a worker, skill, or workflow:

1. Add or update its Markdown frontmatter and body below the owning team directory.
2. Update the owning package's `miez.yaml` for team-level metadata or declarations.
3. Run `miez team build <team-path>` and the focused compiler tests.
4. Verify the rendered Copilot output (`.github/agents/`, `.github/prompts/`,
   `.github/skills/`) if the artifact is target-facing.

The compiler is the source of generated Copilot output. Do not hand-edit
managed files under `.github/` while developing artifact changes.

## Remote teams and APM

Remote teams are installed with `miez team install <github-url>`. A team package
contains authored `miez.yaml` plus `workers/`, `skills/`, `tasks/`, and `workflows/`; its
generated `miez.generated.yaml` is committed with the package and is the install-time
catalog. Already-installed teams are activated with `miez team use <team-id>`;
that command does not download remote content. APM's `apm.yml` and `.apm/`
directories describe portable agent primitives, but they do not replace the
generated miez catalog. miez does not invoke the external APM CLI.

## Authoring conventions

- GitHub Copilot is the only render target. Models carry a single `copilot:`
  field on `ModelOption`; unknown fields are rejected, not silently ignored.
- Installed team source lives at `.miez/miez_modules/<team-id>/`, and the
  lockfile lives at `.miez/miez.lock.yaml`.
- A team's raw source is never trusted at face value: every install records
  a `.miez/miez.lock.yaml` entry, and `miez team audit` can detect drift
  against it offline.
- The stable model id is the only label miez lists and completes; there is
  no separate display-name field on model entries.
