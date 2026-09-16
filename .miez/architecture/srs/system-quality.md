# SRS: system quality and safety

ISO/IEC/IEEE 29148 software requirements specification using EARS and
ISO/IEC/25010 quality categories.

**Status:** current implementation contract, captured 2026-09-15.

## 1. Introduction

This capability collects quality requirements that apply across team loading,
distribution, rendering, and workspace mutation. It exists so reliability,
determinism, and safety are specified once instead of being duplicated in each
functional capability.

## 2. Stakeholders and scope

- **Stakeholders:** all miez operators, CI maintainers, team authors, and
  repository owners.
- **Actors:** the miez process, the local filesystem, and remote GitHub input.
- **In scope:** transaction behavior, repeatability, archive safety, path safety,
  and protection against unmanaged replacement.
- **Out of scope:** provider service-level agreements, machine performance
  guarantees, and secret-manager encryption.

## 3. Non-goals

- These requirements do not prescribe a particular Go implementation pattern.
- These requirements do not replace the functional acceptance criteria in the
  capability SRS documents.

## 4. Quality requirements

#### `reliability/mutations-are-transactional`

While a team installation or workspace mutation is in progress, the system
shall preserve the previous usable state until the new source, generated
output, and local configuration can be committed.

Acceptance:

- WHEN compilation, source copying, workspace saving, or installation-record
  persistence fails THEN the previous active team and managed output remain
  usable.
- WHEN a mutation succeeds THEN its active-team state, installed source,
  generated output, and installation record describe the same team.

#### `determinism/local-rendering-is-repeatable`

When the same team source and workspace state are rendered repeatedly, the
system shall produce equivalent managed paths and content ordering.

Acceptance:

- WHEN rules, skills, workflows, and overrides are unchanged THEN repeated
  rendering produces byte-equivalent generated artifacts and sorted metadata.
- WHEN a local override changes THEN the resulting differences are limited to
  the affected effective output and compilation metadata.

#### `security/remote-input-is-bounded-and-path-safe`

When remote or authored files are read, copied, or rendered, the system shall
reject unsafe paths, symlinks, unsupported file kinds, unmanaged overwrites,
and inputs that exceed configured archive limits.

Acceptance:

- WHEN an archive contains traversal, unsupported entries, or a configured
  size/count violation THEN installation fails before activation.
- WHEN a managed path traverses a symlink or would overwrite an unmanaged file
  THEN rendering fails without replacing that file.

## 5. Open questions

- The repository has no formal performance budget for unusually large teams.
