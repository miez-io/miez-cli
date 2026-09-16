---
name: write-sdd
description: Write a software design description (SDD) for one structured piece using IEEE 1016 viewpoints and C4 component views. Use when defining schema, interfaces, runtime steps, or scaling before implementation.
license: MIT
compatibility: GitHub Copilot and other Agent Skills clients
metadata:
  version: "1.0"
  standards: IEEE 1016, C4
---

# Write an SDD

An SDD is the **building blueprint** for one piece. When relevant SRS
documents exist, the SDD builds upon them; it does not replace them.

Read [references/ieee-1016.md](references/ieee-1016.md) when choosing
views. Copy structure from [assets/sdd-template.md](assets/sdd-template.md).

## Where to write

Do not assume a path.

1. Use the path given by the invoking agent, prompt, or user.
2. Else search for an existing SDD for this piece (headings such as
   `SDD`, `Software design description`, or IEEE 1016 view names).
   Update that file, or add a sibling in the same directory.
3. Else ask where the design description should go.

## When this skill applies

Run when the change has structure: a new store or schema, a new
protocol, a new module boundary, or scaling and failure rules for this
piece.

Skip for a behavior change inside an existing module that needs no new
structure. The SRS is enough.

## Procedure

1. Read relevant SRS documents, when available, and the current system map.
2. Read matching decision records. Do not reopen them.
3. Write only the IEEE 1016 views this piece needs. Keep the design aligned
   with the available SRS without repeating its requirements.
4. Draw C4 component or dynamic diagrams only where a list of parts is
   not enough.
5. If a new container appears, the city map is stale. Stop and return
   to `write-arc42`.
6. Do not use the SDD to introduce new observable behavior. Where relevant,
   capture that behavior in the SRS separately.

## Do not

- Repeat EARS shall-lines as design
- Document the whole product
- Choose among alternatives here (`write-adr`)
- Specify function-level code or file-by-file edits
