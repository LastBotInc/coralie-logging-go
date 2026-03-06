# Documentation Standards

## Principles

- Every exported type/function must have GoDoc comments
- Every file must start with a brief header comment describing its purpose
- No meta commentary ("As an AI...") -- write as normal engineering docs
- Docs must stay in sync with code at all times

## Required Documentation Updates

Any time you add/change/remove:

| Change | Must Update |
|--------|-------------|
| Public API | GoDoc, README.md, relevant Documents/*.md |
| Config fields | Documents/CONFIGURATION.md, README.md |
| Behavior (drop policy, dedupe, shutdown, audio) | Relevant Documents/*.md |
| Folder layout | README.md, Documents/INDEX.md |

**If docs are not updated, the work is considered failed.**

## File Conventions

- Markdown files: UPPER_SNAKE_CASE.md for standards/guides, lowercase for code docs
- Use tables and bullet points over prose
- Keep files under 200 lines where practical
- Include code examples with triple-backtick Go blocks

## Documentation Locations

| Type | Location |
|------|----------|
| API reference | GoDoc comments in source |
| Architecture | Documents/ARCHITECTURE.md |
| Configuration | Documents/CONFIGURATION.md |
| Changelog | Documents/CHANGELOG.md |
| Claude agent docs | docs/ directory |
