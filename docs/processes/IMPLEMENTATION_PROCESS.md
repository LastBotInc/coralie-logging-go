# Implementation Process

## Workflow

| Step | Action | Verification |
|------|--------|-------------|
| 1. Create task | Copy template to `task/active/` | Task file exists |
| 2. Branch | Create feature branch from `main` | `git checkout -b feature/name` |
| 3. Implement | Write code following policies | Code compiles |
| 4. Test | Write and run tests | `go test -race ./...` passes |
| 5. Document | Update all affected docs | Docs match behavior |
| 6. Review | Self-review against checklist | All items checked |
| 7. Complete | Move task to `task/done/` | PR merged |

## Implementation Checklist

- [ ] Code follows Go conventions (gofmt, golint clean)
- [ ] All exported identifiers have GoDoc comments
- [ ] File header comment present
- [ ] Files under 200 lines (split if needed)
- [ ] No per-call heap allocations in hot paths
- [ ] Race-free: `go test -race ./...` passes
- [ ] Tests cover happy path, error path, edge cases
- [ ] Documentation updated (see [DOCUMENTATION.md](../standards/DOCUMENTATION.md))
- [ ] CHANGELOG.md updated with date + bullet list

## Commands Reference

```bash
# Build check
go build ./...

# Run all tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run demo
go run ./cmd/coralie-logging-demo

# Check version
make version

# Release
make bump-patch  # or bump-minor, bump-major
```

## File Organization

| Type | Location |
|------|----------|
| Public API | `pkg/clog/` |
| Audio logging | `pkg/pcmlog/` |
| Internal utilities | `internal/term/`, `internal/timefmt/` |
| Demo CLI | `cmd/coralie-logging-demo/` |
| Example apps | `examples/` |
