# Spec Delta: Execution Engine Bug Fixes

## MODIFIED Requirements

### Requirement: Checkpoint cleanup must enforce max checkpoint limit

The checkpoint cleanup routine SHALL properly enforce the maximum checkpoint limit by deleting the oldest checkpoints when the limit is exceeded.

#### Scenario: Cleanup enforces max checkpoint limit

**Given** an execution engine with `maxCheckpoints` set to 5
**And** there are 10 checkpoints stored
**When** the cleanup routine runs
**Then** the 5 oldest checkpoints should be deleted
**And** only 5 most recent checkpoints remain

#### Scenario: Cleanup handles empty checkpoint directory

**Given** an execution engine with checkpoint cleanup enabled
**And** there are no checkpoints stored
**When** the cleanup routine runs
**Then** no errors occur
**And** the cleanup completes successfully

---

### Requirement: State tracking must not create duplicate states

The execution engine SHALL create exactly one state instance per execution, avoiding duplicate state creation when executing pending tasks.

#### Scenario: Pending state execution creates single state

**Given** a task submitted for execution
**When** the engine processes the pending state
**Then** exactly one ExecutionState instance is created
**And** no duplicate state entries exist in storage

#### Scenario: Resumed execution reuses existing state

**Given** an interrupted execution with existing state
**When** the engine resumes execution
**Then** the existing state is reused
**And** no new state instance is created

---

### Requirement: Process management must support cross-platform builds

The execution engine's subprocess management SHALL compile and run on all supported platforms including Unix-like systems and Windows.

#### Scenario: Build succeeds on Linux

**Given** the execution engine source code
**When** building with `GOOS=linux`
**Then** compilation succeeds without errors
**And** all tests pass

#### Scenario: Build succeeds on Windows

**Given** the execution engine source code
**When** building with `GOOS=windows`
**Then** compilation succeeds without errors
**And** platform-specific syscalls use Windows-compatible APIs

#### Scenario: Process cleanup works on Unix

**Given** an OpenCode subprocess running on Unix
**When** the subprocess times out
**Then** the process group is killed using SIGKILL
**And** all child processes are cleaned up

#### Scenario: Process cleanup works on Windows

**Given** an OpenCode subprocess running on Windows
**When** the subprocess times out
**Then** the process is terminated using Windows APIs
**And** child processes are cleaned up appropriately
