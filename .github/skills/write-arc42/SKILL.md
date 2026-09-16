---
name: write-arc42
description: Write or patch a system architecture description using the arc42 template and C4 context, container, and deployment views. Use when documenting system shape, or when a process, database, or external system is added, removed, or moved.
license: MIT
compatibility: GitHub Copilot and other Agent Skills clients
metadata:
  version: "1.0"
  standards: ISO/IEC/IEEE 42010, arc42, C4
---

# Write arc42

This is the **city map** of the whole product. One architecture
description, not one file per feature.

Standards: [ISO/IEC/IEEE 42010](https://www.iso.org/standard/74393.html)
(architecture description). Template:
[arc42](https://docs.arc42.org/). Diagrams: [C4](https://c4model.com/)
context, container, deployment.

Read [references/arc42-sections.md](references/arc42-sections.md) before
editing a section. Create a missing document from
[assets/arc42-template.md](assets/arc42-template.md).

## Where to write

Do not assume a path or filename.

1. Use the path given by the invoking agent, prompt, or user.
2. Else search for an existing arc42 document (title or headings
   matching the twelve sections). Patch that file.
3. Else ask where the architecture description should go.

Do not create a second system-level arc42 when one already exists.

## When this skill applies

Run when the map changes: a process, container, database, external
system, or deploy shape is added, removed, or moved.

Skip when the change stays inside an existing container. That is an SDD
(`write-sdd`).

## What belongs here

Who uses the system. Which containers exist. Where they run. System-wide
quality goals. An index of decision records (section 9).

## What does not

Schema, queries, indexes, module folders, class lists, algorithm steps.

## Procedure

1. Read the approved SRS and the current architecture description.
2. If none exists, create one from the template.
3. Patch only the sections this change moves. Leave the rest.
4. Draw C4 in Mermaid. Context and container first. Deployment if
   placement changed. Do not put component internals in this file.
5. Point section 9 at decision records. Do not write the decision here
   (`write-adr`).
6. If the SRS is too weak to place the change on the map, stop and
   return to `write-srs`.
