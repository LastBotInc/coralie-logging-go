# Prompting Guide

## Effective Prompts for This Project

### Before Starting Work

Always read CLAUDE.md first -- it contains project-specific commands, conventions, and policies that the agent must follow.

### Task Prompts

#### Bug Fix
```
Fix: [describe the bug]
- Reproduce: [steps or test case]
- Expected: [correct behavior]
- Actual: [current behavior]
- Files likely involved: [list]
```

#### New Feature
```
Add: [feature description]
- API: [proposed function signatures]
- Config: [new config fields]
- Behavior: [expected behavior]
- Tests: [what to test]
- Docs: [what to update]
```

#### Refactor
```
Refactor: [what to change]
- Goal: [why]
- Constraints: [what must not break]
- Files: [scope]
```

### Key Reminders for Agents

| Rule | Details |
|------|---------|
| Always run tests | `go test -race ./...` before and after changes |
| Update docs | Every behavioral change must update Documentation |
| Check CHANGELOG | Add entry with date + bullet list |
| File size limit | Keep files under 200 lines |
| GoDoc required | Every exported identifier needs a comment |

### Project-Specific Context

- This is a **library** consumed by other Go projects
- Public API stability matters -- no breaking changes without major version bump
- Single-writer agent goroutine pattern -- all sink writes are serialized
- Dedupe is on by default -- be aware when testing log output
- Shutdown must be clean -- drain queue, flush all sinks, close handles
