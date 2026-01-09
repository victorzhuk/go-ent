# Tasks: Add Spec Anchoring and Evolution

## 1. Anchoring Mode System
- [ ] 1.1 Create `internal/spec/anchor.go`
- [ ] 1.2 Define `AnchorMode` type (Free, Review, Strict)
- [ ] 1.3 Implement mode configuration storage
- [ ] 1.4 Add mode enforcement hooks
- [ ] 1.5 Create violation detection logic
- [ ] 1.6 Add mode transition validation

## 2. Spec Versioning
- [ ] 2.1 Create `internal/spec/evolution.go`
- [ ] 2.2 Implement spec version tagging (semantic versioning)
- [ ] 2.3 Add spec history storage
- [ ] 2.4 Implement diff algorithm for specs
- [ ] 2.5 Add merge conflict detection
- [ ] 2.6 Create version comparison utilities

## 3. Code Analysis
- [ ] 3.1 Create `internal/spec/analyzer.go`
- [ ] 3.2 Implement AST parsing for Go code
- [ ] 3.3 Extract API signatures (funcs, methods, types)
- [ ] 3.4 Detect data schema changes (structs, fields)
- [ ] 3.5 Identify breaking vs non-breaking changes
- [ ] 3.6 Add support for OpenAPI/Proto specs

## 4. Code-to-Spec Sync
- [ ] 4.1 Create `internal/spec/sync.go`
- [ ] 4.2 Implement change detection from git diff
- [ ] 4.3 Map code changes to spec requirements
- [ ] 4.4 Generate spec delta proposals
- [ ] 4.5 Add confidence scoring for inferred changes
- [ ] 4.6 Create review workflow for proposed updates

## 5. MCP Tools Implementation
- [ ] 5.1 Implement `spec_anchor_set` tool
- [ ] 5.2 Implement `spec_anchor_status` tool
- [ ] 5.3 Implement `spec_diff` tool
- [ ] 5.4 Implement `spec_sync` tool
- [ ] 5.5 Add formatted output with diffs
- [ ] 5.6 Register tools in `register.go`

## 6. Integration
- [ ] 6.1 Add anchoring hooks to code execution
- [ ] 6.2 Integrate with existing `spec_validate`
- [ ] 6.3 Add anchoring mode to project config
- [ ] 6.4 Create CI script templates
- [ ] 6.5 Add logging for anchor violations

## 7. Testing
- [ ] 7.1 Test Free mode (no enforcement)
- [ ] 7.2 Test Review mode (suggestions)
- [ ] 7.3 Test Strict mode (blocking)
- [ ] 7.4 Test spec diff with various changes
- [ ] 7.5 Test code-to-spec sync accuracy
- [ ] 7.6 Test version merge scenarios
- [ ] 7.7 Integration tests with full workflow

## 8. Documentation
- [ ] 8.1 Document anchoring modes and use cases
- [ ] 8.2 Add workflow examples (free → review → strict)
- [ ] 8.3 Document CI integration
- [ ] 8.4 Create migration guide
- [ ] 8.5 Add architecture diagrams
