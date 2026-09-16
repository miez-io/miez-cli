---
name: commit
description: Create small, atomic, logically grouped commits with Conventional Commit subjects.
argument-hint: Describe the change to commit, or leave empty to commit the current work.
---

Create small, atomic, logically grouped commits. A commit describes one
independently understandable change and is easy to review, revert, and
validate.

## Process

1. Inspect `git status`, `git diff`, and `git diff --cached` before
   committing. Do not assume that all staged changes belong together.
2. Stage only the files or hunks for the current intent with explicit paths
   or `git add -p`. Preserve unrelated user changes and unrelated staged
   work.
3. Run the narrowest relevant validation before committing. Do not commit
   known failures without documenting the reason.
4. Before finalizing, review `git diff --cached` and confirm that the staged
   diff contains exactly one logical change.

## Grouping rules

- Use the "AND" test: if the commit subject naturally needs "and", split the
  work into separate commits unless the changes are inseparable for a
  buildable, coherent result.
- Keep implementation and the focused tests that verify the same behavior
  together. Use separate commits for independent tests, documentation,
  configuration, or cleanup.

## Commit message

Use a Conventional Commit subject in imperative mood, with no period and a
concise summary (ideally 72 characters or fewer):

	type(scope): summary

Use `feat`, `fix`, `refactor`, `test`, `docs`, `build`, `ci`, `perf`, or
`chore` as appropriate. For example: `feat: implement employee filtering`.