# Security Review: [FEATURE NAME]

## Review Metadata

| Field | Value |
|-------|-------|
| Feature | |
| Reviewer | |
| Date | |
| Status | Pending / Approved / Rejected |

## Checklist

### Input Validation
- [ ] All external inputs validated
- [ ] Config values bounds-checked at Init()

### Resource Management
- [ ] Bounded queues/buffers used
- [ ] File handles closed on shutdown
- [ ] Goroutines stopped on shutdown

### Concurrency
- [ ] No shared mutable state without synchronization
- [ ] `go test -race ./...` passes
- [ ] Deadlock-free shutdown sequence

### Secrets
- [ ] No secrets in source code
- [ ] No secrets logged
- [ ] Tokens passed via config/environment

### Dependencies
- [ ] No new dependencies added (or justified)
- [ ] `govulncheck ./...` clean

## Findings

| ID | Severity | Description | Remediation | Status |
|----|----------|-------------|-------------|--------|
| | | | | |

## Decision

- [ ] Approved
- [ ] Approved with conditions
- [ ] Rejected (reason: )
