---
name: verify-artifacts
description: Verify miez artifacts against their domain boundaries and suggest improvements, if necessary.
argument-hint: Review the team's workers, skills, tasks, and workflows.
---

Verify miez artifacts to make sure they fit the miez domain standard. Suggest improvements if they do not.

Review these package directories:

- `workers/`
- `skills/`
- `tasks/`
- `workflows/`

Use the [Miez artifact authoring instruction](../instructions/miez-artifact-authoring.instructions.md)
as the source of truth for each artifact's domain boundary.

For each artifact:

- Check its boundary, frontmatter, references, and scope.
- Identify violations, overlap, and inconsistencies.
- Suggest the smallest useful improvement, including the artifact path and
	reason, only if necessary.
- Report "No issue" when it passes review.

Do not edit artifacts during verification unless the user explicitly asks for
the suggested improvements to be applied.
