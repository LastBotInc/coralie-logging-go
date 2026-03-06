# Generic Policies

## Code Quality

- All code must compile without warnings
- All tests must pass before merging
- Race detector must pass: `go test -race ./...`
- Files should stay under 200 lines; split as needed

## Version Control

- Main branch: `main`
- Feature branches: `feature/description`
- Fix branches: `fix/description`
- Commit messages: concise, imperative mood ("Add hook support", not "Added hooks")
- Never force-push to `main`

## Documentation

- Docs must update with every behavioral change (see [DOCUMENTATION.md](../standards/DOCUMENTATION.md))
- If docs are not updated, the work is considered failed
- CHANGELOG.md updated with date + bullet list for every PR

## Releases

- Semantic versioning: `vX.Y.Z`
- Patch: bug fixes, backward-compatible
- Minor: new features, backward-compatible
- Major: breaking changes
- Use `make bump-patch`, `make bump-minor`, `make bump-major`

## Security

- No secrets in code or logs
- No hardcoded tokens or credentials
- Validate all external inputs
- Use bounded resources to prevent exhaustion

## Dependencies

- Minimize external dependencies (this is a library)
- Justify every new dependency
- Compatible licenses only (MIT, BSD, Apache 2.0)
