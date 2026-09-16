# SRS: team distribution

ISO/IEC/IEEE 29148 software requirements specification using EARS.

**Status:** approved target contract for the workflow artifact change, drafted
2026-09-15.

## 1. Introduction

This capability defines how miez obtains, installs, updates, and identifies
team source from GitHub. It exists to make a team installation reproducible and
usable for both public and authenticated private repositories.

## 2. Stakeholders and scope

- **Stakeholders:** repository operators, team authors, CI maintainers, and
  GitHub repository administrators.
- **Actors:** the operator, GitHub, the local credential sources, and miez.
- **In scope:** source isolation, lock records, GitHub references, credentials,
  install, update, and read-only upstream version checks.
- **Out of scope:** generated Copilot rendering details, drift classification,
  and MCP value prompting.

## 3. Non-goals

- The capability does not support non-GitHub remote hosts.
- It does not execute an external APM binary or infer a miez team from an APM
  package without an explicit team manifest.
- It does not silently update an installed team without operator approval or an
  explicit non-interactive approval.

## 4. Requirements

### 4.1 Functional

#### `team-distribution/source-modules-are-isolated-from-rendered-output`

When a team is installed from a remote reference, the system shall keep its
source separate from generated target output and from an uninstalled authoring
bundle.

Acceptance:

- WHEN a remote installation succeeds THEN the source is available in the
  installed team module area and generated files are in managed target output.
- WHEN a team is bootstrapped for authoring THEN its files remain outside the
  installed module area until explicitly installed.

#### `team-distribution/miez-owned-distribution-state`

When miez initializes or mutates a workspace, the system shall keep installed
team source and its reproducibility record inside one miez-owned workspace
area.

Acceptance:

- WHEN initialization or installation succeeds THEN no new root-level
  `miez_modules/` directory or `miez.lock.yaml` file is created.
- WHEN update, audit, or outdated runs after installation THEN each operation
  reads the same miez-owned source and lock records as installation.

#### `team-distribution/legacy-distribution-state-is-migrated`

When a workspace contains a valid legacy root-level team source or lock record,
the system shall preserve its installed team and integrity data while adopting
the miez-owned distribution location.

Acceptance:

- WHEN a legacy workspace is migrated THEN every installed team and every
  lockfile entry remains available at the new location with equivalent file
  hashes.
- WHEN migration fails THEN the legacy source and lock record remain usable
  and no active team or generated output is lost.

#### `team-distribution/lockfile-records-reproducible-installs`

When a team is installed or updated, the system shall record the source
reference, resolved revision, package version, source file set, and content
integrity data needed to reproduce and inspect that installation.

Acceptance:

- WHEN an installation succeeds THEN the workspace lock record contains the
  selected reference, resolved commit, version, and a hash for every recorded
  source file.
- WHEN the lock record is read without network access THEN the revision and
  source hashes are available for local comparison.

#### `team-distribution/team-use-installs-from-any-visibility-repository`

When a valid public or accessible private GitHub reference is supplied, the
system shall validate, install, and activate the selected team without changing
the previous active installation if the operation fails.

Acceptance:

- WHEN an accessible repository contains a valid team THEN that team becomes
  active and is available for rendering.
- WHEN authentication, download, validation, rendering, or persistence fails
  THEN the previous active source and effective output remain usable.

#### `team-distribution/install-uses-generated-team-index`

When a team is installed, the system shall use its generated `miez.generated.yaml` as
the operational catalog and shall preserve the package `miez.yaml` as part of
the installed source.

Acceptance:

- WHEN an installed package contains a valid generated `miez.generated.yaml` THEN
  workflow, worker, skill, model, and MCP information is available without
  scanning artifact frontmatter.
- WHEN a remote package has no generated `miez.generated.yaml` or its references are
  invalid THEN installation fails before activation.
- WHEN installation succeeds THEN the package `miez.yaml` is present in the
  installed module and is included in source integrity tracking.

#### `team-distribution/workspace-manifest-declares-team-sources`

Where a workspace declares named team sources, the system shall resolve those
names from its root `miez.yaml` while using the lockfile for the installed
revision and integrity state.

Acceptance:

- WHEN a named team source is present in the workspace manifest THEN
  `team use` and `team update` can resolve it without duplicating the URL in a
  command argument.
- WHEN the workspace manifest and lockfile disagree about an installed team
  revision THEN the lockfile remains the source for audit and reproducible
  update decisions.

#### `team-distribution/install-exposes-artifact-catalog`

When a team installation succeeds, the system shall make the installed team's
workflow artifact identifiers available to workflow selection and completion.

Acceptance:

- WHEN `team use` completes THEN the active team's available workflows can be
  selected without another network request.
- WHEN a team is switched THEN completion and selection use only the newly
  active team's workflow artifacts.

#### `team-distribution/credential-resolution-order-mirrors-host-conventions`

When GitHub access requires authentication, the system shall resolve a
credential using a fixed priority order and shall identify the selected source
without exposing the credential value.

Acceptance:

- WHEN multiple supported credential sources exist THEN the highest-priority
  source is selected consistently.
- WHEN no supported credential source exists for a private repository THEN the
  operation fails clearly without attempting an unsupported fallback.

#### `team-distribution/team-update-refreshes-an-installed-team`

When an installed team has a newer revision at its recorded reference, the
system shall present the source changes and apply them only after explicit
confirmation or explicit non-interactive approval.

Acceptance:

- WHEN the recorded revision has not changed THEN update reports that the team
  is current and performs no installation.
- WHEN a dry-run is requested THEN source additions, changes, and removals are
  reported without changing repository state.
- WHEN interactive confirmation is declined THEN the installed team and its
  effective output remain unchanged.

#### `team-distribution/update-uses-active-team-by-default`

Where `team update` receives no team reference and the workspace has an active
team, the system shall plan and apply the update for that active team.

Acceptance:

- WHEN no team reference is supplied and an active team is installed THEN the
  update plan identifies that team and retains the existing confirmation,
  dry-run, and approval behavior.
- WHEN no team reference is supplied and no active team is available THEN the
  command fails clearly before contacting GitHub or changing local state.
- WHEN an explicit team name or GitHub URL is supplied THEN the command keeps
  its existing explicit-reference behavior.

#### `team-distribution/named-team-references`

Where a repository defines a name for a team reference, the system shall accept
the name wherever the corresponding remote reference is accepted.

Acceptance:

- WHEN a registered name is supplied THEN its mapped GitHub reference is used.
- WHEN an unregistered name is supplied and no installed team matches it THEN
  the operation reports an unknown team name.

#### `team-distribution/outdated-reports-available-updates-without-changing-anything`

When an operator checks for outdated installed teams, the system shall compare
their recorded revisions with the latest revisions at the same references
without changing local installation state.

Acceptance:

- WHEN no team is selected THEN every installed team is reported.
- WHEN a newer revision exists THEN the current and newer revisions are shown.
- WHEN the check completes THEN local source, lock records, and generated
  output are unchanged.

### 4.2 Constraints

#### `compatibility/github-is-the-distribution-host`

The system shall accept supported GitHub HTTP(S) references with an optional
branch and optional path to a team and shall reject unsupported hosts or unsafe
URL components.

Acceptance:

- WHEN a supported repository, branch, and team path are supplied THEN the
  corresponding team source is selected.
- WHEN a reference contains credentials, a port, query, fragment, or an
  unsupported host THEN it is rejected before download.

#### `security/credentials-are-not-persisted`

When a GitHub credential is resolved, the system shall use it for remote access
without persisting the credential in repository configuration, installation
records, or downloaded team source.

Acceptance:

- WHEN a credential is resolved THEN diagnostic output names only its source
  tier and not its value.
- WHEN installation completes THEN the credential value is absent from the
  repository manifest, lock record, and installed source.

## 5. Open questions

- The current distribution protocol has no retry or backoff policy for GitHub
  rate limits.
