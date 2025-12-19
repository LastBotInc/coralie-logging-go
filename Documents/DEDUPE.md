# Deduplication

## Overview

Deduplication collapses consecutive identical log messages to reduce noise and improve readability.

## Behavior

When enabled, deduplication:
1. Tracks the last logged message (level + interface + formatted message)
2. Suppresses consecutive identical messages
3. Emits a summary line when:
   - A different message is logged
   - Shutdown occurs
   - The message changes in level or interface

## Example

Without deduplication:
```
[INFO] Application: Processing request
[INFO] Application: Processing request
[INFO] Application: Processing request
[INFO] Application: Processing request
[INFO] Application: Processing request
```

With deduplication:
```
[INFO] Application: Processing request
[INFO] Application: last message repeated 4 more times
```

## Configuration

```go
cfg.Dedupe.Enabled = true
cfg.Dedupe.SummaryFormat = "last message repeated %d more times"
```

## Rules

- Only consecutive messages are collapsed
- Messages must match: level, interface, and formatted message
- Summary is routed to the same level/interface as the original
- Summary is flushed on shutdown

## Disabling

Set `Dedupe.Enabled = false` to disable deduplication.

