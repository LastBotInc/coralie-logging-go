# Task Guide

## Task Workflow

```
task/templates/  -->  task/active/  -->  task/done/  -->  task/archive/
   (copy)            (work here)       (completed)      (long-term)
```

## Creating a Task

1. Copy the appropriate template from `task/templates/`
2. Rename with descriptive name: `task/active/add-json-sink.md`
3. Fill in all sections
4. Begin implementation

## Task Templates

| Template | Use For |
|----------|---------|
| `task/templates/GENERIC.md` | General tasks |
| `task/templates/golang/DESIGN.md` | Go design tasks |
| `task/templates/golang/IMPLEMENTATION.md` | Go implementation tasks |

## Task Lifecycle

| Stage | Location | Meaning |
|-------|----------|---------|
| Active | `task/active/` | Currently being worked on |
| Done | `task/done/` | Completed, ready for review |
| Archive | `task/archive/` | Historical reference |

## Task Naming Convention

```
YYYY-MM-DD-short-description.md
```

Example: `2026-03-06-add-json-sink.md`

## Completion Checklist

Before moving to `task/done/`:

- [ ] All acceptance criteria met
- [ ] Tests pass: `go test -race ./...`
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Code reviewed
