# Tasks: Add AST-Based Code Operations

## Dependencies
- None (uses Go stdlib)

## 1. Core Infrastructure

- [x] 1.1 Create `internal/ast/parser.go` - Parse Go files to AST ✓ 2026-01-13
- [x] 1.2 Create `internal/ast/query.go` - Query AST for patterns ✓ 2026-01-13
- [x] 1.3 Create `internal/ast/transform.go` - Transform AST nodes ✓ 2026-01-13
- [x] 1.4 Create `internal/ast/printer.go` - Print AST back to Go code ✓ 2026-01-13

## 2. Symbol Operations

- [x] 2.1 Create `internal/ast/symbols.go` - Symbol table construction ✓ 2026-01-13
- [x] 2.2 Implement find-all-references ✓ 2026-01-13
- [x] 2.3 Implement go-to-definition ✓ 2026-01-15
- [x] 2.4 Implement type-aware rename ✓ 2026-01-15

## 3. MCP Tools

- [x] 3.1 Implement `go_ent_ast_parse` - Parse file and return structure ✓ 2026-01-15
- [x] 3.2 Implement `go_ent_ast_query` - Find functions/types/interfaces by pattern ✓ 2026-01-15
- [x] 3.3 Implement `go_ent_ast_rename` - Safe symbol rename across files ✓ 2026-01-16
- [x] 3.4 Implement `go_ent_ast_refs` - Find all references to symbol ✓ 2026-01-15
- [x] 3.5 Implement `go_ent_ast_extract` - Extract code to new function ✓ 2026-01-15

## 4. Structural Queries

- [x] 4.1 Query by function signature pattern ✓ 2026-01-15
- [x] 4.2 Query by interface implementation ✓ 2026-01-15
- [x] 4.3 Query by struct field type ✓ 2026-01-16
- [x] 4.4 Query by import dependency ✓ 2026-01-16

## 5. Code Generation

- [x] 5.1 Create AST template system ✓ 2026-01-16
- [x] 5.2 Generate interface implementations ✓ 2026-01-16
- [ ] 5.3 Generate test scaffolds from function signatures

## 6. Testing

- [x] 6.1 Unit tests for AST parsing ✓ 2026-01-15
- [ ] 6.2 Test rename across multiple files
- [ ] 6.3 Test edge cases (shadowing, embedding, generics)
