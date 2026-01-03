# ci-pipeline Specification

## Purpose
TBD - created by archiving change add-build-infrastructure. Update Purpose after archive.
## Requirements
### Requirement: Go Build Validation
CI SHALL validate that Go code builds successfully on every push and pull request.

#### Scenario: Build job executes
- **WHEN** code is pushed to main branch or PR is opened
- **THEN** CI workflow triggers `go-cli` job
- **AND** job runs `go build -v ./...`

#### Scenario: Build failure detected
- **WHEN** Go code has build errors
- **THEN** CI job fails
- **AND** error details are displayed in CI logs

#### Scenario: Build success reported
- **WHEN** Go code builds successfully
- **THEN** CI job passes
- **AND** green checkmark is displayed on commit/PR

### Requirement: Go Test Execution
CI SHALL run all Go tests with race detection and coverage reporting.

#### Scenario: Test job executes
- **WHEN** CI workflow runs
- **THEN** tests execute with `go test -race -cover ./...`
- **AND** race detector is enabled

#### Scenario: Test failure detected
- **WHEN** any test fails
- **THEN** CI job fails
- **AND** failing test details are shown in logs

#### Scenario: Test coverage reported
- **WHEN** tests complete successfully
- **THEN** coverage percentage is reported
- **AND** coverage data is available in logs

### Requirement: Go Linting Validation
CI SHALL run golangci-lint to enforce code quality standards.

#### Scenario: Lint job executes
- **WHEN** CI workflow runs
- **THEN** golangci-lint-action runs with version v4
- **AND** linter uses `.golangci.yml` configuration

#### Scenario: Lint failure detected
- **WHEN** code has linting violations
- **THEN** CI job fails
- **AND** linting issues are displayed with file locations

#### Scenario: Lint passes
- **WHEN** code meets linting standards
- **THEN** lint check passes
- **AND** no violations are reported

### Requirement: Go Environment Setup
CI SHALL set up Go 1.23 environment for building and testing.

#### Scenario: Go installation
- **WHEN** CI job starts
- **THEN** Go 1.23 is installed using setup-go action
- **AND** go command is available in PATH

#### Scenario: Go module cache
- **WHEN** dependencies are downloaded
- **THEN** Go module cache is used
- **AND** subsequent runs use cached dependencies

### Requirement: Parallel Job Execution
CI SHALL run Go validation and plugin validation jobs in parallel for faster feedback.

#### Scenario: Jobs run concurrently
- **WHEN** CI workflow triggers
- **THEN** `go-cli` and `validate-plugin` jobs run in parallel
- **AND** neither job waits for the other

#### Scenario: Independent job failures
- **WHEN** one job fails and another succeeds
- **THEN** overall CI status reflects the failure
- **AND** both job results are visible

### Requirement: Plugin Validation
CI SHALL validate plugin structure and configuration files.

#### Scenario: Plugin JSON validation
- **WHEN** plugin files are modified
- **THEN** JSON syntax is validated
- **AND** required plugin files are checked

#### Scenario: Plugin path correction
- **WHEN** checking plugin.json files
- **THEN** correct path `plugins/*/.claude-plugin/plugin.json` is used
- **AND** files are found successfully

### Requirement: Workflow Triggers
CI workflow SHALL trigger on appropriate repository events.

#### Scenario: Push to main triggers CI
- **WHEN** commits are pushed to main branch
- **THEN** CI workflow executes
- **AND** all jobs run

#### Scenario: Pull request triggers CI
- **WHEN** pull request is opened or updated
- **THEN** CI workflow executes
- **AND** results are posted to PR

### Requirement: CI Status Reporting
CI SHALL report build, test, and lint status clearly for developers.

#### Scenario: Status checks on PR
- **WHEN** PR is open with CI running
- **THEN** status checks appear on PR page
- **AND** each job status is visible

#### Scenario: Failed job details
- **WHEN** CI job fails
- **THEN** error logs are accessible
- **AND** failure reason is clear

#### Scenario: Success notification
- **WHEN** all CI jobs pass
- **THEN** green checkmark appears on commit
- **AND** PR is marked as passing checks

