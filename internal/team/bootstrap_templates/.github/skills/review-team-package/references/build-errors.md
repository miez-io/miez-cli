# miez team build errors

Common `miez team build .` failures, their cause, and the fix.

## Package metadata

| Message | Cause | Fix |
|---|---|---|
| `team package id "…" is not kebab-case` | `miez.yaml` `id` breaks `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` | rename to kebab-case |
| `team package version is required` | missing `version` | add e.g. `version: 1.0.0` |
| `team package name is required` | missing `name` | add a display name |
| `model "…" must set a copilot name` | a `models:` entry has no `copilot:` | add the Copilot model string |
| `default_model "…" is not supported` | `default_model` is not in the merged catalog | use a built-in id or declare it under `models:` |
| `generated miez.generated.yaml metadata does not match miez.yaml` | generated file edited by hand or stale | re-run the build; never hand-edit it |

## Structure

| Message | Cause | Fix |
|---|---|---|
| `required artifact directory … is missing` | `workers/`, `skills/`, `tasks/`, or `workflows/` absent | create it |
| `team package must contain at least one workflow Markdown file` | no workflow | add `workflows/<id>.md` |
| `skill … must use skills/<skill-id>/SKILL.md` | skill nested too deep or too shallow | move to exactly that path |
| `artifact path … must not be a symlink` | symlinked artifact | replace with a real file |

## Frontmatter

| Message | Cause | Fix |
|---|---|---|
| `missing YAML frontmatter` / `unterminated YAML frontmatter` | no `---` block or unclosed | wrap frontmatter in `---` fences |
| `frontmatter field "id" is not allowed` | leftover `id` from an older package | delete the line; the id comes from the file path |
| `file name implies id "…", want "…"` | generated catalog disagrees with the file layout | re-run `miez team build .`; never hand-edit the generated file |
| `invalid frontmatter: field … not found` | unknown key on a worker or workflow — both are strict | remove the key; only documented fields are allowed |
| `duplicate worker id` / `duplicate skill id` / `duplicate workflow id` | two artifacts share an id | make ids unique |
| `kind "…" is not supported` | invalid, missing, or prompt-only worker kind | use `kind: agent` |

## References

| Message | Cause | Fix |
|---|---|---|
| `model "…" is not supported; choose one of: …` | unknown model id | pick a listed id or declare it in `miez.yaml` |
| `worker "…" references unknown mcp id "…"` | `tools:` entry not in `miez.yaml` `mcp:` | declare the server or drop the reference |
| `worker "…" references undeclared skill "…"` | `skills:` entry has no `skills/<id>/SKILL.md` | create the skill or drop the reference |
| `unknown worker "…"` under a phase | workflow phase names a missing worker | fix the id or create the worker |
| `at least one phase is required` | workflow has no `phases:` | add a phase |
| `at least one worker is required` | phase has an empty `workers:` | list at least one worker |
| `duplicate phase id` | repeated phase id in one workflow | make phase ids unique |

## Built-in model ids

`gpt-5` (default), `gpt-5.1`, `claude-sonnet-4.5`, `claude-opus-4.5`,
`gemini-3-pro`.

A team extends or overrides these in `miez.yaml`:

```yaml
models:
  - id: gpt-5.2
    copilot: GPT-5.2 (copilot)
```
