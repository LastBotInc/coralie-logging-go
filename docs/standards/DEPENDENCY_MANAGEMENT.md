# Dependency Management

## Principles

- Minimal external dependencies -- this is a library consumed by others
- Every new dependency adds transitive weight to consumers
- Prefer stdlib solutions over third-party packages

## Current Dependencies

| Module | Purpose | Direct/Indirect |
|--------|---------|-----------------|
| `golang.org/x/sys` | TTY detection, terminal capabilities | indirect |

## Commands

```bash
# Add a dependency
go get github.com/example/pkg@latest

# Tidy (remove unused, add missing)
go mod tidy

# Verify checksums
go mod verify

# View dependency graph
go mod graph

# Check for available updates
go list -m -u all
```

## Rules

| Rule | Details |
|------|---------|
| Justify additions | Document why stdlib is insufficient before adding |
| Pin versions | Use exact versions or minimum version selection |
| No replace directives | In library go.mod (only allowed in example modules for local dev) |
| License check | New deps must have compatible license (MIT, BSD, Apache 2.0) |
| Security audit | Check for known vulnerabilities with `govulncheck` |

## Versioning

- Library uses semver tags: `vX.Y.Z`
- `make version` -- show current version
- `make bump-patch` / `make bump-minor` / `make bump-major` -- tag and push
- Go module proxy caches tags; once pushed, a tag should not be moved
