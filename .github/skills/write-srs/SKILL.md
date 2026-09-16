---
name: write-srs
description: Write a software requirements specification (SRS) using ISO/IEC/IEEE 29148 and EARS shall-lines. Use when defining, changing, reviewing, or inventorying what a capability must do, or when asked to write requirements, an SRS, or EARS statements.
license: MIT
compatibility: GitHub Copilot and other Agent Skills clients
metadata:
  version: "1.0"
  standards: ISO/IEC/IEEE 29148, EARS, ISO/IEC 25010
---

# Write an SRS

An SRS is the document. A requirement is one EARS shall-line inside it.
This skill teaches the writing rules. It does not own a folder.

Read [references/29148-and-ears.md](references/29148-and-ears.md) when
writing or checking a requirement. Copy structure from
[assets/srs-template.md](assets/srs-template.md).

## Where to write

Do not assume a path.

1. Use the path given by the invoking agent, prompt, or user.
2. Else search the repository for an existing SRS for this capability
   (headings such as `SRS`, `Requirements specification`, or EARS
   `shall` lines). Update that file, or add a sibling in the same directory.
3. Else ask where the first SRS should go.

## Procedure

1. Name the capability, actors, and why it exists.
2. Write non-goals.
3. Write one EARS shall-line per need. Give each a stable ID.
4. Put a pass/fail acceptance check under every requirement.
5. Ask when a need is unclear. Do not invent scope.
6. If the input is a foreign spec or user story, rewrite it into EARS.
   Keep the meaning, not the original shape.

## Structure rule of thumb

Group SRS documents by stable, user-visible capability rather than by feature
branch, package, class, or individual command. Keep one SRS per capability and
use a small index when a product has several capabilities. A feature may update
several existing SRS documents; create a new document only when it introduces
a capability with its own durable contract.

## EARS

| Pattern | Form |
|---|---|
| Ubiquitous | The system shall `<behavior>`. |
| Event | When `<event>`, the system shall `<behavior>`. |
| State | While `<state>`, the system shall `<behavior>`. |
| Optional | Where `<feature>`, the system shall `<behavior>`. |
| Unwanted | If `<unwanted>`, the system shall `<behavior>`. |

## 29148 bar

Necessary, implementation-free, unambiguous, singular, verifiable,
traceable. Full list in the reference.

## Do not

- Put stack, folders, schemas, or commands in a requirement
- Put several needs in one sentence
- Store user stories as a second requirements list
- Write architecture or design in this file
