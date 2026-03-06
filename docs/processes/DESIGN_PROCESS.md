# Design Process

## Overview

This process applies when adding new features, changing public API, or modifying core behavior.

## Steps

| Step | Action | Output |
|------|--------|--------|
| 1. Requirements | Capture what and why | `docs/templates/design/features/REQUIREMENTS.md` |
| 2. Threat analysis | Identify security concerns | `security/THREAT_ANALYSIS.md` |
| 3. Design | Write technical approach | `docs/templates/design/features/DESIGN.md` |
| 4. Review | Validate design against requirements | Approved design doc |
| 5. Handoff | Create implementation task | `task/active/*.md` |

## Design Checklist

- [ ] Requirements documented and understood
- [ ] API consistent with existing `clog.*` patterns
- [ ] Performance impact assessed (no per-call allocations)
- [ ] Backward compatibility verified (or major version bump planned)
- [ ] Thread safety considered (single-writer agent goroutine pattern)
- [ ] Shutdown behavior defined (drain, flush, finalize)
- [ ] Test plan written
- [ ] Documentation update plan included

## Patterns to Follow

- **Sink pattern**: Implement `Sink` interface, register in agent loop
- **Config pattern**: Add field to `Config` struct, set default in `DefaultConfig()`
- **API pattern**: Top-level function in `api.go` forwarding to agent
- **Dedupe-aware**: New sinks receive already-deduped events from agent

## When to Skip

- Bug fixes to existing behavior
- Internal refactors with no API/behavior change
- Documentation-only changes
