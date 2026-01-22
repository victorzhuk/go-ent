# Proposal: Add AST-Based Code Operations

## Why

Current code transformations rely on text-based pattern matching, which is fragile and can introduce subtle bugs. Go's excellent tooling (go/ast, go/parser) enables deterministic, type-aware refactoring that's reliable and safe.

Inspired by:
- [Oh-My-OpenCode](https://github.com/code-yeongyu/oh-my-opencode) - LSP & AST integration for deterministic surgical refactoring
- Go's built-in refactoring tools (gorename, goimports, etc.)

## What Changes

- **AST Parsing**: Parse Go files into AST for structural operations
- **Symbol Operations**: Rename, find references, go-to-definition using AST
- **Structural Queries**: Find functions, types, interfaces by pattern
- **Safe Refactoring**: Type-aware rename, extract function, inline variable
- **Code Generation**: Generate code from AST templates

## Impact

- Affected specs: ast-tools (new capability)
- Affected code: internal/ast/, cmd/mcp/
- Dependencies: None (uses Go stdlib)

## Key Benefits

1. **Reliability**: Deterministic transformations vs fuzzy text matching
2. **Safety**: Type-aware operations prevent broken references
3. **Speed**: AST queries faster than regex for structural patterns
4. **Go-Native**: Leverage Go's excellent tooling ecosystem
