# Design Documentation Standards

## When to Write a Design Document

- New public API surface (new package, new exported types)
- Changes to core behavior (queue, dedupe, shutdown semantics)
- New sink implementations
- Breaking changes to configuration

## Design Document Structure

Every design doc must contain:

| Section | Content |
|---------|---------|
| Summary | One-paragraph description of the change |
| Motivation | Why this change is needed |
| Design | Technical approach with diagrams/pseudocode |
| API changes | New/modified exported types and functions |
| Configuration | New/modified config fields |
| Testing | How the change will be tested |
| Migration | Impact on existing consumers |

## File Location

- System-level designs: `docs/design/system/`
- Feature designs: `docs/design/features/`
- Use templates from `docs/templates/design/features/`

## Review Checklist

- [ ] Design addresses all requirements
- [ ] API is consistent with existing patterns
- [ ] Performance impact assessed
- [ ] No breaking changes without major version bump
- [ ] Test plan covers edge cases
- [ ] Documentation update plan included
