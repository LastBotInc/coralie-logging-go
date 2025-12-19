# Testing

## Running Tests

### Library Tests

From repository root:
```bash
go test ./...
go test -race ./...
```

### Example Module Tests

```bash
cd examples/fyne-audio-monitor
go test ./...
go test -race ./...
```

## Test Coverage

All public APIs and core functionality are unit tested:
- Deduplication behavior
- Drop policies
- File routing
- Hook invocation
- Statistics
- Shutdown and leak detection
- Audio writer (synthetic data, no hardware)

## Test Organization

- Unit tests are colocated with source files (`*_test.go`)
- Example app tests are in the example module
- No tests require real hardware (microphone, etc.)

## Writing Tests

### Example: Testing Deduplication

```go
func TestDedupe_CollapsesConsecutive(t *testing.T) {
    // Setup
    // Test consecutive messages are collapsed
    // Verify summary is emitted
}
```

### Example: Testing File Routing

```go
func TestFileRouting(t *testing.T) {
    // Create temp directory
    // Configure file routing
    // Log messages
    // Verify files contain expected levels
}
```

## Mocking

Example app uses interface-based mocking for logger dependencies:
- Define logger interface
- Implement mock in tests
- Inject mock into components

## Race Detection

Always run tests with `-race` flag to detect data races.

## CI Integration

Tests should pass in CI without requiring:
- Real audio hardware
- Interactive terminals
- Specific file system permissions (use temp dirs)

