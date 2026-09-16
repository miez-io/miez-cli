---
status: accepted
date: 2026-09-15
---

# Make `team install` the remote bootstrap operation

## Context and Problem Statement

The public command surface currently splits one user goal across two commands:
`miez init --github URL` creates workspace state and installs the first team,
while `miez team use <name-or-url>` requires that workspace state to exist
before it can install or activate a team. This makes a fresh repository unable
to use the command that appears to install a team, and makes `init` an
installation command under a different name.

The initial installation does not create a consuming-workspace `miez.yaml`
with name-to-URL aliases. The downloaded team's package `miez.yaml` is stored
inside its installed module and is not a registry for other teams. A bare name
therefore has no remote resolution source during first installation.

The installation transaction already validates the downloaded team, resolves
MCP configuration, renders Copilot output, writes the lock record, and
activates the team. The decision is where first-time workspace preparation
belongs and how local activation remains distinct from remote installation.

This decision satisfies `team-install/bootstraps-a-missing-workspace`,
`team-install/reuses-existing-workspace`, and `cli/command-surface-is-stable`.

## Considered Options

* Keep `miez init` as a required first step, then use `team use` for installation or activation
* Make `team use <name-or-url>` initialize implicitly while retaining one overloaded command
* Remove the root `init` command, add `team install <github-url>` for remote installation, and reserve `team use <team-id>` for local activation

## Decision Outcome

Chosen option: **Remove the root `init` command and make `team install` own
remote installation and first-time workspace bootstrap**, because it matches
package-manager semantics, makes the first command actionable in a fresh
repository, and keeps remote installation separate from activation of source
that is already installed.

`team install` accepts a GitHub URL only. miez does not provide a remote team
registry, create a consuming-workspace alias map during installation, or infer
a URL from an arbitrary team id. A bare name is therefore rejected before any
remote request.

On the first install, the command prompts for the render target unless
supplied by a flag, creates the workspace state as part of the successful
transaction, resolves required MCP values through the existing early-resolution
flow, and activates the downloaded team. On later installs, it opens the
existing workspace, reuses its persisted target configuration, and replaces the
active installation through the same transaction. A supplied target that
conflicts with the persisted configuration is rejected before mutation.

`team use` remains available only for activating an already-installed local
team by id. It performs no GitHub lookup or download. Successful activation
uses the existing compiler transaction to remove the previously active team's
managed artifacts and add the selected team's artifacts, then persists the new
active team. `team update`, `team list`, `team audit`, and `team outdated` keep
their existing lifecycle roles.

### Consequences

* Good, because a fresh repository has one clear first command and does not
  need a separate initialization ceremony.
* Good, because target prompting, MCP setup, rendering, lockfile persistence,
  and activation stay on the existing install path instead of being split
  between commands.
* Good, because `team use` has one local meaning and cannot accidentally imply
  a network installation that requires prior workspace state.
* Good, because installation has one unambiguous input and does not depend on a
  workspace manifest that the first install does not create.
* Bad, because `init` is a breaking CLI removal and documentation, completion,
  and tests must migrate to `team install`.
* Bad, because operators must retain or repeat the full GitHub URL; named remote
  aliases are not part of this installation flow.
* Bad, because first-install rollback must cover both workspace preparation and
  the existing source/output transaction, including any locally persisted MCP
  values created during the attempt.
* Bad, because operators who currently pass a URL to `team use` must change to
  `team install`.
