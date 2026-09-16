# SRS: team audit

ISO/IEC/IEEE 29148 software requirements specification using EARS.

**Status:** current implementation contract, captured 2026-09-15.

## 1. Introduction

This capability verifies whether installed team source still matches its
recorded installation. It exists to support local integrity checks and CI gates
without requiring network access.

## 2. Stakeholders and scope

- **Stakeholders:** repository operators, CI maintainers, and reviewers.
- **Actors:** the operator, a CI process, the local source directory, and the
  lock record.
- **In scope:** modified, missing, extra, clean, and CI-gated audit results.
- **Out of scope:** upstream version discovery, team installation, and repair of
  drift findings.

## 3. Non-goals

- Audit does not update source or rewrite the lock record.
- Audit does not contact GitHub or decide whether an upstream revision exists.

## 4. Requirements

### 4.1 Functional

#### `team-audit/detects-drift-against-the-lockfile`

When an installed team is audited, the system shall compare its deployed source
files with the recorded file set and content hashes and report missing,
modified, and undeclared extra files.

Acceptance:

- WHEN every recorded file matches and no extra file exists THEN the team is
  reported clean.
- WHEN a file is modified, missing, or extra THEN the finding identifies its
  team, status, and relative path.

#### `team-audit/ci-gate-mode`

Where audit CI mode is requested, the system shall return a non-zero result
when any selected team has a drift finding and a zero result when all selected
teams are clean.

Acceptance:

- WHEN CI audit finds drift THEN the process lists every finding and returns a
  non-zero exit code.
- WHEN CI audit finds no drift THEN the process returns exit code zero.

## 5. Open questions

- The human-readable audit format is intentionally smaller than the validation
  report because its findings are file-oriented.
