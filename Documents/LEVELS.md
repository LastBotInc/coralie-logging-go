# Log Levels

## Available Levels

coralie-logging-go provides seven log levels, ordered by severity:

1. **Debug** - Detailed diagnostic information
2. **Info** - General informational messages
3. **Success** - Successful operation completion
4. **Warning** - Warning conditions that may need attention
5. **Fail** - Operation failures that are recoverable
6. **Error** - Error conditions requiring attention
7. **Catastrophe** - Critical failures requiring immediate action

## Usage

Each level has a corresponding function:

```go
clog.Debug("Component", "Debug message: %v", data)
clog.Info("Component", "Info message")
clog.Success("Component", "Operation completed")
clog.Warning("Component", "Warning: %s", message)
clog.Fail("Component", "Operation failed: %v", err)
clog.Error("Component", "Error occurred: %v", err)
clog.Catastrophe("Component", "Critical failure: %v", err)
```

## Level Constants

Levels are represented as `clog.Level` constants:
- `clog.LevelDebug`
- `clog.LevelInfo`
- `clog.LevelSuccess`
- `clog.LevelWarning`
- `clog.LevelFail`
- `clog.LevelError`
- `clog.LevelCatastrophe`

## Routing

Levels can be routed to different files or omitted from console output:

```go
cfg.File.PerLevel = map[clog.Level]string{
    clog.LevelError: "errors.log",
    clog.LevelCatastrophe: "critical.log",
}
cfg.Console.OmitLevels = map[clog.Level]bool{
    clog.LevelDebug: true, // Omit debug from console
}
```

## Default Interface

If no interface is specified, the default is `"Application"`.

