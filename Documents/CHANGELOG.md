# Changelog

All notable changes to coralie-logging-go will be documented in this file.

## [Unreleased]

### Added
- Initial repository structure
- Documentation stubs
- Package skeletons
- Core logging API with async agent goroutine
- Bounded queue with configurable drop policies (drop_new, drop_old)
- Statistics tracking (drops per level, accepted count, emitted count)
- Init/Shutdown with sync.Once guard and re-initialization support
- All log level functions (Debug, Info, Success, Warning, Fail, Error, Catastrophe)
- Message formatting in agent goroutine using fmt.Appendf
- Comprehensive unit tests including race detection tests

