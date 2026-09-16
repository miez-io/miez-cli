# ADR and MADR

## Nygard ADR

Michael Nygard, [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
(2011), defined the short record:

- Title
- Status (proposed, accepted, deprecated, superseded)
- Context
- Decision
- Consequences

Write one decision per file. Keep the files in version control next to
the code.

## MADR 4.0

[Markdown Architectural Decision Records](https://adr.github.io/madr/)
is a Markdown template on top of that idea. Current release:
[MADR 4.0.0](https://github.com/adr/madr/tree/4.0.0).

License: MIT OR CC0-1.0.

Minimal sections (official minimal template):

1. Title
2. Context and Problem Statement
3. Considered Options
4. Decision Outcome
5. Consequences (good / bad)

The [full template](https://github.com/adr/madr/blob/4.0.0/template/adr-template.md)
adds Decision Drivers, Confirmation, Pros and Cons of the Options, and
More Information. Use the full form only when the choice is expensive
or contested.

Suggested filename, from MADR: `NNNN-title-with-dashes.md`. Use the
repository's existing scheme if it already has one.

## What is not an ADR

- A requirement (no alternatives; it is a need)
- A design description (how the chosen option is built)
- A changelog of implementation work

## Sources

- Michael Nygard, Documenting Architecture Decisions, 2011
- [MADR](https://adr.github.io/madr/)
- [adr/madr v4.0.0](https://github.com/adr/madr/tree/4.0.0)
- [adr-tools naming](https://github.com/npryce/adr-tools)
