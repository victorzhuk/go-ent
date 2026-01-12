# Tasks: Add Plugin System

## 1. Plugin Manager
- [x] Create internal/plugin/manager.go
- [x] Implement Install(ctx, name) error
- [x] Implement Uninstall(ctx, name) error
- [x] Implement List() []PluginInfo
- [x] Implement Enable/Disable(name) error

## 2. Plugin Manifest
- [x] Create internal/plugin/manifest.go
- [x] Define Manifest struct
- [x] Implement YAML parsing
- [x] Add manifest validation
- [x] Define SkillRef, AgentRef, RuleRef types

## 3. Plugin Loader
- [x] Create internal/plugin/loader.go
- [x] Implement LoadPlugin(path) (*Plugin, error)
- [x] Integrate with skill registry
- [x] Integrate with agent registry

## 4. Plugin Validator
- [x] Create internal/plugin/validator.go
- [x] Validate manifest schema
- [x] Check for conflicts
- [x] Verify dependencies

## 5. Marketplace Client
- [x] Create internal/marketplace/client.go
- [x] Implement Search(ctx, query) ([]PluginInfo, error)
- [x] Implement Download(ctx, name, version) ([]byte, error)
- [x] Add HTTP client with retries

## 6. Marketplace Search
- [x] Create internal/marketplace/search.go
- [x] Add filtering by category, author
- [x] Add sorting by downloads, rating

## 7. Marketplace Install
- [x] Create internal/marketplace/install.go
- [x] Download and verify plugin
- [x] Extract to plugin directory
- [x] Validate before enabling

## 8. Rules Engine
- [x] Create internal/rules/engine.go
- [x] Define Rule struct
- [x] Implement Evaluate(ctx, event) ([]Action, error)
- [x] Load rules from plugins

## 9. Rule Definition
- [x] Create internal/rules/rule.go
- [x] Define rule YAML format
- [x] Add condition evaluation

## 10. Rule Evaluator
- [x] Create internal/rules/evaluator.go
- [x] Implement rule matching
- [x] Add action execution

## 11. MCP Tools
- [x] Create internal/mcp/tools/plugin_list.go
- [x] Create internal/mcp/tools/plugin_install.go
- [x] Create internal/mcp/tools/plugin_search.go
- [x] Create internal/mcp/tools/plugin_info.go

## 12. Testing
- [ ] Unit tests for plugin manager
- [ ] Unit tests for marketplace client
- [ ] Integration tests for plugin loading
- [ ] E2E test with sample plugin

## 13. Documentation
- [ ] Plugin development guide
- [ ] Manifest format reference
- [ ] Marketplace usage guide

## 14. Fix Critical Issues (Security + Broken Functionality)
- [x] Fix path traversal vulnerability in install.go:72
- [x] Fix Install() to extract archive instead of writing as manifest
- [x] Implement unloadPlugin() to properly unregister skills/agents
- [x] Add UnregisterSkill/UnregisterAgent to Registry interface

## 15. Fix High Priority Issues
- [x] Add URL encoding in marketplace client.go:47-58
- [x] Fix resource leak: close file handles immediately in install.go:85-89
- [ ] Fix RegisterAgent no-op in server.go:80-82 (error or implement)

## 16. Fix Medium Priority Issues
- [ ] Add logging for failed plugin initialization in manager.go:57-60
- [ ] Cache regex compilation in evaluator.go:194
- [ ] Implement executeRejectAction in evaluator.go:76
- [ ] Implement executeModifyAction in evaluator.go:80
