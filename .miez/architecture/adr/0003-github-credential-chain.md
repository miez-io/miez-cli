---
status: accepted
date: 2026-09-15
---

# Use a fixed GitHub credential chain without git credential-helper fallback

## Context and Problem Statement

Remote teams may be public or private. The CLI needs predictable credential
selection that works in local development and CI, does not couple miez to a
user's unrelated Git credential configuration, and never writes the resolved
token into repository artifacts. The implementation follows a fixed chain and
reports only the selected tier under `--verbose`.

This decision satisfies `team-distribution/credential-resolution-order-mirrors-host-conventions`
and `security/credentials-are-not-persisted`.

## Considered Options

* Use Git's credential-helper protocol as the primary or final fallback
* Resolve miez-specific and standard GitHub environment variables, then use
  `gh auth token` as the final supported tier
* Require one explicitly configured miez credential and reject all other
  sources

## Decision Outcome

Chosen option: **Use the fixed environment chain followed by `gh auth token`,
with no git credential-helper fallback**, because it mirrors the useful APM
shape, keeps miez and APM credentials from silently authenticating each other,
and avoids an additional subprocess protocol dependency.

The priority is per-organization `GITHUB_MIEZ_PAT_<ORG>`,
`GITHUB_MIEZ_PAT`, `GITHUB_TOKEN`, `GH_TOKEN`, then `gh auth token
--hostname github.com`.

### Consequences

* Good, because precedence is deterministic and observable without exposing a
  token.
* Good, because public repositories need no credential and private repositories
  can use common CI or `gh` authentication paths.
* Good, because credentials are request-scoped and excluded from the manifest,
  lockfile, and downloaded source.
* Bad, because users who rely only on a configured Git credential helper are not
  supported by this version.
* Bad, because the chain is GitHub-specific and does not generalize to other
  private Git hosts.
