# AGENTS

Use `taskmd` as the source of truth for repo-local task tracking.

## Recommended flow

```bash
taskmd list --json
taskmd next --json
taskmd bulk add --file -
taskmd validate
```

## Notes

- The canonical file is `docs/Task.md`.
- Prefer `--json` for read operations when another agent consumes the output.
- Use `bulk add`, `bulk edit`, and `bulk remove` with stdin for batch updates.
- Run `taskmd validate` after direct file edits.

