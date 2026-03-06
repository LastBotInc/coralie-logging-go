# Threat Analysis: [FEATURE NAME]

## Assets

| Asset | Description | Sensitivity |
|-------|-------------|-------------|
| Log data | Application log messages | Medium |
| Config | Logger configuration | Low |
| Tokens | Sink authentication tokens (e.g., BetterStack) | High |
| Files | Log files on disk | Medium |

## Threats

| ID | Threat | Asset | Likelihood | Impact | Risk |
|----|--------|-------|-----------|--------|------|
| T1 | Secrets logged in plaintext | Log data | Medium | High | High |
| T2 | Unbounded resource consumption | Memory | Medium | High | High |
| T3 | Race condition data corruption | Log data | Low | Medium | Low |
| T4 | | | | | |

## Mitigations

| Threat | Mitigation | Status |
|--------|-----------|--------|
| T1 | Caller responsibility; document "never log secrets" | In place |
| T2 | Bounded queue + drop policy | In place |
| T3 | Single-writer agent goroutine + race detector in CI | In place |

## Residual Risks

| Risk | Accepted Because |
|------|-----------------|
| | |
