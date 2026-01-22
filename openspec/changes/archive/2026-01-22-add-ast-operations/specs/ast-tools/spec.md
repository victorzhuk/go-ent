## ADDED Requirements

### Requirement: AST Parsing

The system SHALL parse Go source files into AST for structural analysis.

#### Scenario: Parse single file
- **WHEN** `go_ent_ast_parse` is called with `file: "internal/handler/user.go"`
- **THEN** the file is parsed into AST
- **AND** functions, types, and imports are returned in structured format

#### Scenario: Parse with position info
- **WHEN** `go_ent_ast_parse` is called with `include_positions: true`
- **THEN** line and column numbers are included for each node

#### Scenario: Parse error handling
- **WHEN** file contains syntax errors
- **THEN** parse errors are returned with line numbers
- **AND** partial AST is available if possible

### Requirement: Structural Queries

The system SHALL query AST to find code elements by pattern.

#### Scenario: Find functions by name pattern
- **WHEN** `go_ent_ast_query` is called with `type: "function"`, `pattern: "Handle*"`
- **THEN** all functions matching the pattern are returned
- **AND** results include file path and line number

#### Scenario: Find interface implementations
- **WHEN** `go_ent_ast_query` is called with `type: "implements"`, `interface: "io.Reader"`
- **THEN** all types implementing the interface are returned

#### Scenario: Find by signature
- **WHEN** `go_ent_ast_query` is called with `type: "function"`, `signature: "(context.Context, string) error"`
- **THEN** functions matching the signature pattern are returned

#### Scenario: Query scope
- **WHEN** `go_ent_ast_query` is called with `scope: "internal/..."`
- **THEN** only files in internal/ and subdirectories are searched

### Requirement: Symbol Rename

The system SHALL provide type-aware symbol renaming across files.

#### Scenario: Rename function
- **WHEN** `go_ent_ast_rename` is called with `symbol: "CreateUser"`, `new_name: "NewUser"`, `scope: "./..."`
- **THEN** the function is renamed in definition and all call sites
- **AND** changes are applied atomically

#### Scenario: Rename with preview
- **WHEN** `go_ent_ast_rename` is called with `dry_run: true`
- **THEN** proposed changes are returned without applying
- **AND** each change shows file, line, and diff

#### Scenario: Conflict detection
- **WHEN** new name would conflict with existing symbol
- **THEN** error is returned explaining the conflict
- **AND** no changes are applied

### Requirement: Find References

The system SHALL find all references to a symbol.

#### Scenario: Find all references
- **WHEN** `go_ent_ast_refs` is called with `symbol: "UserService"`, `file: "internal/service/user.go"`, `line: 15`
- **THEN** all references to the symbol are returned
- **AND** references are categorized (definition, read, write)

#### Scenario: Include test files
- **WHEN** `go_ent_ast_refs` is called with `include_tests: true`
- **THEN** references in *_test.go files are included

### Requirement: Extract Function

The system SHALL extract code blocks into new functions.

#### Scenario: Extract to function
- **WHEN** `go_ent_ast_extract` is called with `file: "handler.go"`, `start_line: 20`, `end_line: 35`, `name: "validateInput"`
- **THEN** the code block is extracted to a new function
- **AND** parameters are inferred from used variables
- **AND** return values are inferred from assignments

#### Scenario: Extract with explicit signature
- **WHEN** `go_ent_ast_extract` is called with explicit `params` and `returns`
- **THEN** the specified signature is used

### Requirement: Code Generation from AST

The system SHALL generate Go code from AST templates.

#### Scenario: Generate interface implementation
- **WHEN** `go_ent_ast_generate` is called with `type: "impl"`, `interface: "io.ReadWriter"`, `struct: "MyBuffer"`
- **THEN** stub methods are generated for the interface
- **AND** method bodies contain `panic("not implemented")` placeholder

#### Scenario: Generate test scaffold
- **WHEN** `go_ent_ast_generate` is called with `type: "test"`, `function: "CreateUser"`
- **THEN** test function scaffold is generated
- **AND** table-driven test structure is used

### Requirement: Package-Level Analysis

The system SHALL analyze package dependencies and structure.

#### Scenario: Get package info
- **WHEN** `go_ent_ast_parse` is called with `package: "internal/service"`
- **THEN** all files in the package are parsed
- **AND** exported symbols are listed
- **AND** import graph is included

#### Scenario: Detect circular dependencies
- **WHEN** package import creates a cycle
- **THEN** the cycle is detected and reported
