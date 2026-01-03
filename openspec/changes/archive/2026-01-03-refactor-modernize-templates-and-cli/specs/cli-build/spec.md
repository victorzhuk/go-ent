# Spec Delta: CLI Build System

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: CLI Entry Point Pattern
The CLI SHALL follow the run(ctx, getenv, stdout, stderr) error pattern with proper signal handling.

#### Scenario: Main function delegates to run
- **GIVEN** CLI application starts
- **WHEN** `main()` function is called
- **THEN** it SHALL call `run(context.Background(), os.Getenv, os.Stdout, os.Stderr)`
- **AND** it SHALL log error with `slog.Error` if run returns error
- **AND** it SHALL call `os.Exit(1)` only after error logging

#### Scenario: Run function is testable
- **GIVEN** CLI implementation
- **WHEN** examining `run()` signature
- **THEN** signature SHALL be `run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error`
- **AND** function SHALL NOT reference `os.Getenv`, `os.Stdout`, or `os.Stderr` directly
- **AND** function SHALL enable dependency injection for testing

#### Scenario: Signal handling for graceful shutdown
- **GIVEN** CLI application is running
- **WHEN** SIGTERM, SIGINT, or SIGQUIT signal is received
- **THEN** application SHALL log "shutdown signal received"
- **AND** application SHALL initiate graceful shutdown
- **AND** application SHALL use `signal.NotifyContext` for signal handling

#### Scenario: Graceful shutdown timeout
- **GIVEN** shutdown signal is received
- **WHEN** application initiates shutdown
- **THEN** shutdown SHALL use fresh `context.Background()` as parent
- **AND** shutdown context SHALL have 30-second timeout
- **AND** shutdown SHALL complete within timeout or force exit

#### Scenario: No log.Fatal outside main
- **GIVEN** CLI implementation
- **WHEN** examining error handling
- **THEN** `log.Fatal` SHALL only appear in `main()` function (if at all)
- **AND** all other functions SHALL return errors
- **AND** errors SHALL be wrapped with context using `fmt.Errorf("context: %w", err)`

#### Scenario: Structured logging setup
- **GIVEN** CLI application starts
- **WHEN** `run()` function executes
- **THEN** it SHALL call `setupLogger()` to configure structured logger
- **AND** it SHALL set default logger with `slog.SetDefault(logger)`
- **AND** logger SHALL support JSON and text formats
- **AND** logger SHALL support debug, info, warn, error levels

### Requirement: Go Version for CLI
The CLI SHALL be built with Go 1.25.5.

#### Scenario: Root go.mod specifies Go 1.25.5
- **GIVEN** CLI project go.mod
- **WHEN** examining go directive
- **THEN** version SHALL be `go 1.25.5`

#### Scenario: Build succeeds with Go 1.25.5
- **GIVEN** CLI source code
- **WHEN** running `make build`
- **THEN** build SHALL succeed using Go 1.25.5 toolchain
- **AND** binary SHALL be created in `dist/goent`

### Requirement: Build Metadata for CLI
The CLI SHALL include VERSION and VCS_REF in build metadata.

#### Scenario: Makefile defines build metadata
- **GIVEN** CLI Makefile
- **WHEN** examining variable definitions
- **THEN** Makefile SHALL define `VERSION` from git describe or "dev"
- **AND** Makefile SHALL define `VCS_REF` from git rev-parse or "unknown"
- **AND** Makefile SHALL pass metadata via LDFLAGS

#### Scenario: LDFLAGS include metadata
- **GIVEN** CLI Makefile build target
- **WHEN** examining build command
- **THEN** LDFLAGS SHALL include `-X main.version=$(VERSION)`
- **AND** LDFLAGS SHALL include `-X main.vcsRef=$(VCS_REF)`

## REMOVED Requirements

None.

## RENAMED Requirements

None.
