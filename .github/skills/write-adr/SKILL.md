---
name: write-adr
description: Write an architecture decision record using MADR 4 and Nygard ADR when a real technical choice has alternatives. Use when recording why one option won, or when asked to write an ADR or MADR.
license: MIT
compatibility: GitHub Copilot and other Agent Skills clients
metadata:
  version: "1.0"
  standards: Nygard ADR, MADR 4.0
---

# Write an ADR

An ADR records **why one option won**. It is not a design dump and not
a requirement.

Read [references/madr.md](references/madr.md) before writing. Use
[assets/adr-template.md](assets/adr-template.md) (MADR 4.0 minimal).

## Where to write

Do not assume a path.

1. Use the path given by the invoking agent, prompt, or user.
2. Else search for existing ADRs (`ADR`, `Decision Record`, MADR
   headings, or `NNNN-title.md`). Add the next file in that directory
   using the local naming scheme.
3. Else ask where decision records live.

If an architecture description exists, add one index line to its
decisions section. Do not paste the ADR into that file.

## When this skill applies

Write only when two or more real options were considered.

Skip when there is no fork. Schema fields and query shape are design
(`write-sdd`).

## Procedure

1. State the problem. Link the SRS IDs that force the choice if they
   exist.
2. List the options that were actually considered.
3. Choose one. Say why.
4. Write consequences: what becomes easier, what becomes harder.
5. Status is `accepted` for a new decision unless the user says otherwise.

## Do not

- Write an ADR for work that had no alternative
- Repeat the SDD
- Invent requirements
