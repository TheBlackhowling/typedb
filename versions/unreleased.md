# Unreleased Changes

## Added

- **Logging Support** - Added comprehensive logging interface for typedb operations
  - New `Logger` interface with `Debug`, `Info`, `Warn`, and `Error` methods
  - Global logger support via `SetLogger()` and `GetLogger()` functions
  - Per-instance logger support via `WithLogger()` option when opening connections
  - `NewDBWithLogger()` function to create DB instances with specific loggers
  - Logging integrated throughout query execution, transactions, and connection management
  - Default no-op logger when no logger is provided (zero overhead when logging is disabled)
  - Log messages include relevant context (queries, arguments, errors, connection details)
