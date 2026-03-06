# Requirements: [FEATURE NAME]

## Functional Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR1 | | Must | Proposed |
| FR2 | | Must | Proposed |
| FR3 | | Should | Proposed |

## Non-Functional Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| NFR1 | No per-call allocations in hot path | Must | Proposed |
| NFR2 | Pass race detector | Must | Proposed |
| NFR3 | | Should | Proposed |

## Constraints

- Must be backward-compatible with existing API (or justify breaking change)
- Must work without CGO for core library
- Must not add external dependencies without justification

## Dependencies

- Depends on: (list other features/packages)
- Depended on by: (list consumers)

## Acceptance Criteria

- [ ] All functional requirements implemented
- [ ] All tests pass: `go test -race ./...`
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
