# Tasks: Add Plugin System

## 1. Plugin Manager
- [ ] Create internal/plugin/manager.go
- [ ] Implement Install(ctx, name) error
- [ ] Implement Uninstall(ctx, name) error
- [ ] Implement List() []PluginInfo
- [ ] Implement Enable/Disable(name) error

## 2. Plugin Manifest
- [ ] Create internal/plugin/manifest.go
- [ ] Define Manifest struct
- [ ] Implement YAML parsing
- [ ] Add manifest validation
- [ ] Define SkillRef, AgentRef, RuleRef types

## 3. Plugin Loader
- [ ] Create internal/plugin/loader.go
- [ ] Implement LoadPlugin(path) (*Plugin, error)
- [ ] Integrate with skill registry
- [ ] Integrate with agent registry

## 4. Plugin Validator
- [ ] Create internal/plugin/validator.go
- [ ] Validate manifest schema
- [ ] Check for conflicts
- [ ] Verify dependencies

## 5. Marketplace Client
- [ ] Create internal/marketplace/client.go
- [ ] Implement Search(ctx, query) ([]PluginInfo, error)
- [ ] Implement Download(ctx, name, version) ([]byte, error)
- [ ] Add HTTP client with retries

## 6. Marketplace Search
- [ ] Create internal/marketplace/search.go
- [ ] Add filtering by category, author
- [ ] Add sorting by downloads, rating

## 7. Marketplace Install
- [ ] Create internal/marketplace/install.go
- [ ] Download and verify plugin
- [ ] Extract to plugin directory
- [ ] Validate before enabling

## 8. Rules Engine
- [ ] Create internal/rules/engine.go
- [ ] Define Rule struct
- [ ] Implement Evaluate(ctx, event) ([]Action, error)
- [ ] Load rules from plugins

## 9. Rule Definition
- [ ] Create internal/rules/rule.go
- [ ] Define rule YAML format
- [ ] Add condition evaluation

## 10. Rule Evaluator
- [ ] Create internal/rules/evaluator.go
- [ ] Implement rule matching
- [ ] Add action execution

## 11. MCP Tools
- [ ] Create internal/mcp/tools/plugin_list.go
- [ ] Create internal/mcp/tools/plugin_install.go
- [ ] Create internal/mcp/tools/plugin_search.go
- [ ] Create internal/mcp/tools/plugin_info.go

## 12. Testing
- [ ] Unit tests for plugin manager
- [ ] Unit tests for marketplace client
- [ ] Integration tests for plugin loading
- [ ] E2E test with sample plugin

## 13. Documentation
- [ ] Plugin development guide
- [ ] Manifest format reference
- [ ] Marketplace usage guide
