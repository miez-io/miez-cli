---
status: accepted
date: 2026-09-15
---

# Keep installed modules and the lockfile at repository root

## Context and Problem Statement

The distribution layer needs to keep raw downloaded team source distinct from
local miez state and from generated Copilot files. It also needs a committed
record of resolved refs and hashes. The implementation places raw source in
`miez_modules/` and the lockfile in `miez.lock.yaml`, both at repository root,
while `.miez/` remains the home for local CLI state.

This decision satisfies `team-distribution/source-modules-are-isolated-from-rendered-output`
and `team-distribution/lockfile-records-reproducible-installs`.

## Considered Options

* Put modules and lock data under `.miez/`
* Put modules in a visible repository-root dependency directory and keep a
  root lockfile, mirroring the APM-style distribution shape
* Reuse generated `.github/` output as the installed source record

## Decision Outcome

Chosen option: **Use repository-root `miez_modules/` and `miez.lock.yaml`**,
because the raw source behaves like a dependency cache, while the lockfile is
a portable committed installation record and neither should be conflated with
local runtime/configuration state.

### Consequences

* Good, because source, generated output, local state, and reproducibility
  metadata have distinct ownership and lifecycle boundaries.
* Good, because audit can hash raw installed source without network access and
  the lockfile can be reviewed with normal repository changes.
* Bad, because the repository gains visible root-level dependency artifacts and
  must ignore the modules directory explicitly.
