# Tasks: Restore Minimal `init` Command

## Status: ✅ COMPLETE (Task 9/9 Complete)

**Estimated Complexity:** Medium (8-12 hours total)
**Target LOC:** ~300-400 lines for `internal/cli/init.go`

---

## Task Breakdown

### Task 1: Rename binary directory
**Priority:** High
**Complexity:** Low
**Estimated Time:** 30 minutes
**Dependencies:** None

**Steps:**
1. Rename `cmd/go-ent/` → `cmd/ent/`
2. Update `Makefile` build target
3. Update `.goreleaser.yml` if exists
4. Update documentation references
5. Update GitHub Actions workflows if any

**Files Modified:**
- `cmd/ent/` (renamed from cmd/go-ent/)
- `Makefile`
- `README.md`
- `.goreleaser.yml` (if exists)

**Validation:**
- [x] `make build` creates `bin/ent` binary
- [x] `./bin/ent version` works
- [x] All make targets work with new binary name

---

### Task 2: Create command structure
**Priority:** High
**Complexity:** Low
**Estimated Time:** 1-2 hours
**Dependencies:** None

**Steps:**
1. Create `internal/cli/init.go`
2. Define `initFlags` struct:
   ```go
   type initFlags struct {
       tool    string
       prefix  string
       force   bool
       dryRun  bool
   }
   ```
3. Create `newInitCmd()` function with Cobra
4. Wire up flags:
   - `--tool` (required, supports comma-separated values)
   - `--prefix` (default: "ent")
   - `--force` (default: false)
   - `--dry-run` (default: false)
5. Add to `internal/cli/root.go`: `cmd.AddCommand(newInitCmd())`

**Files Created:**
- `internal/cli/init.go`

**Files Modified:**
- `internal/cli/root.go`

**Validation:**
- [x] `ent init --help` shows correct usage
- [x] `ent init` without flags shows error (--tool required)
- [x] Binary compiles

---

### Task 3: Implement resource loading
**Priority:** High
**Complexity:** Medium
**Estimated Time:** 2-3 hours
**Dependencies:** Task 2

**Steps:**
1. Implement `loadAgents(fs embed.FS)` - Parse `agents/meta/*.yaml` files:
   ```go
   type agentMeta struct {
       Name         string   `yaml:"name"`
       Description  string   `yaml:"description"`
       Model        string   `yaml:"model"`
       Color        string   `yaml:"color"`
       Skills       []string `yaml:"skills"`
       Tools        []string `yaml:"tools"`
       Dependencies []string `yaml:"dependencies"`
       Tags         []string `yaml:"tags"`
   }
   ```
2. Implement `loadPrompts(fs embed.FS)` - Load `agents/prompts/agents/*.md`
3. Implement `loadShared(fs embed.FS)` - Concatenate `agents/prompts/shared/_*.md` in order:
   - `_principals.md`
   - `_judgment.md`
   - `_openspec.md`
   - `_conventions.md`
   - `_handoffs.md`
   - `_tooling.md`
4. Implement `loadTemplate(fs embed.FS, tool string)`:
   - Parse `agents/templates/claude.yaml.tmpl` for Claude Code
   - Parse `agents/templates/opencode.yaml.tmpl` for OpenCode
5. Use `gopkg.in/yaml.v3` for YAML parsing
6. Use `text/template` for template rendering

**Files Modified:**
- `internal/cli/init.go`

**Validation:**
- [x] All 16 agent metadata files load correctly
- [x] All 16 agent prompts load correctly
- [x] All 6 shared sections concatenate in correct order
- [x] Both templates parse without errors

---

### Task 4: Implement agent generation
**Priority:** High
**Complexity:** Medium
**Estimated Time:** 2-3 hours
**Dependencies:** Task 3

**Steps:**
1. Implement `renderAgent(meta *agentMeta, prompt, shared string, tpl *template.Template)`:
   - Execute template with agent metadata
   - Append `---` separator
   - Append shared sections
   - Append agent-specific prompt
2. Implement target path generation:
   - Claude: `.claude/agents/{prefix}/{name}.md`
   - OpenCode: `.opencode/agent/{name}.md` (no prefix subdirectory)
3. Implement `writeFile(path, content string, force, dryRun bool)`:
   - Check if file exists (error unless `--force`)
   - Create parent directories
   - Write file (skip if `--dry-run`)
   - Print status message
4. Parse comma-separated `--tool` values:
   ```go
   tools := strings.Split(flags.tool, ",")
   for _, tool := range tools {
       tool = strings.TrimSpace(tool)
       // generate for this tool
   }
   ```

**Files Modified:**
- `internal/cli/init.go`
- `cmd/ent/main.go`

**Validation:**
- [x] Agent files render correctly for Claude Code
- [x] Agent files render correctly for OpenCode
- [x] `--tool=claude,opencode` works
- [x] `--force` overwrites existing files
- [x] `--dry-run` shows preview without writing
- [x] Parent directories created automatically

---

### Task 5: Implement command/skill copying
**Priority:** High
**Complexity:** Low
**Estimated Time:** 1-2 hours
**Dependencies:** Task 3

**Steps:**
1. Implement `copyCommands(fs embed.FS, targetDir, prefix string, force, dryRun bool)`:
   - Read `plugins/go-ent/commands/*.md`
   - Write to `.claude/commands/{prefix}/` or `.opencode/commands/{prefix}/`
   - Preserve file names
2. Implement `copySkills(fs embed.FS, targetDir, prefix string, force, dryRun bool)`:
   - Read `plugins/go-ent/skills/**/*.md` recursively
   - Write to `.claude/skills/{prefix}/` or `.opencode/skills/{prefix}/`
   - Preserve directory structure
3. Handle file conflicts with `--force` flag
4. Print progress for each file copied

**Files Modified:**
- `internal/cli/init.go`

**Validation:**
- [x] All 4 commands copy correctly
- [x] All skills copy preserving subdirectory structure
- [x] `--force` works for commands/skills
- [x] `--dry-run` shows all files to be copied

---

### Task 6: Add summary output
**Priority:** Medium
**Complexity:** Low
**Estimated Time:** 30 minutes
**Dependencies:** Tasks 4-5

**Steps:**
1. Implement `printSummary(agents, commands, skills, tool, dryRun bool)`:
   ```
   ✅ Initialized go-ent for Claude Code

   Created:
     16 agents in .claude/agents/ent/
     4 commands in .claude/commands/ent/
     17 skills in .claude/skills/ent/

   Next steps:
     1. Restart Claude Code
     2. Run: /ent:plan "description"
   ```
2. Show different message for `--dry-run`
3. Show different next steps for OpenCode vs Claude Code

**Files Modified:**
- `internal/cli/init.go`

**Validation:**
- [x] Summary shows correct counts
- [x] Summary shows correct paths
- [x] Dry-run message is clear

---

### Task 7: Add unit tests
**Priority:** High
**Complexity:** Medium
**Estimated Time:** 2-3 hours
**Dependencies:** Tasks 2-6

**Steps:**
1. Create `internal/cli/init_test.go`
2. Add table-driven tests for path generation:
   - Claude Code agent paths
   - OpenCode agent paths
   - Command paths for both tools
   - Skill paths for both tools
3. Add tests for `--tool` parsing:
   - Single tool: `claude`
   - Single tool: `opencode`
   - Multiple tools: `claude,opencode`
   - Invalid tools should error
4. Add integration test with temp directory:
   - Run `init` with all tools
   - Verify files created
   - Verify file contents
   - Verify directory structure
5. Add test for `--dry-run`:
   - Should not create files
   - Should print preview
6. Add test for `--force`:
   - Should overwrite existing files

**Files Created:**
- `internal/cli/init_test.go`

**Validation:**
- [x] All tests pass
- [x] Test coverage >80%
- [x] Tests use `testify/assert`
- [x] Tests use `t.Parallel()` where appropriate

---

### Task 8: Update documentation
**Priority:** Medium
**Complexity:** Low
**Estimated Time:** 30 minutes
**Dependencies:** All previous tasks

**Steps:**
1. Update `README.md`:
   - Change `go-ent` → `ent` in examples
   - Add `ent init` command documentation
   - Update installation instructions
2. Update `docs/SETUP_GUIDE.md`:
   - Update binary name
   - Update init command examples
   - Remove references to deleted `--agents`, `--include-deps` flags
3. Update `CLAUDE.md` if it references old command

**Files Modified:**
- `README.md`
- `docs/SETUP_GUIDE.md`
- `CLAUDE.md` (if needed)

**Validation:**
- [x] No references to `go-ent` binary remain
- [x] All examples use `ent` binary
- [x] Init command documented correctly

---

### Task 9: Final verification
**Priority:** High
**Complexity:** Low
**Estimated Time:** 30 minutes
**Dependencies:** All previous tasks

**Steps:**
1. Run `make build` - must succeed
2. Test in temp directory:
   ```bash
   cd /tmp/test-project
   go mod init test
   ent init --tool=claude
   ls -la .claude/agents/ent/
   cat .claude/agents/ent/coder.md
   ```
3. Test for OpenCode:
   ```bash
   ent init --tool=opencode --force
   ls -la .opencode/agent/
   cat .opencode/agent/coder.md
   ```
4. Test both tools:
   ```bash
   rm -rf .claude .opencode
   ent init --tool=claude,opencode
   ```
5. Test `--dry-run`:
   ```bash
   rm -rf .claude
   ent init --tool=claude --dry-run
   # Should show preview, no files created
   ```
6. Run `make test` - all tests must pass
7. Run `make lint` - no lint errors

**Validation:**
- [x] Build succeeds
- [x] `ent init --tool=claude` creates valid config
- [x] `ent init --tool=opencode` creates valid config
- [x] `ent init --tool=claude,opencode` creates both
- [x] `--dry-run` works correctly
- [x] `--force` works correctly
- [x] Tests pass
- [x] Lint clean

---

## Task Dependencies Graph

```
Task 1 (Rename binary)
  ↓
Task 2 (Command structure)
  ↓
Task 3 (Resource loading)
  ↓
Task 4 (Agent generation) ← Task 5 (Command/skill copying)
  ↓                                ↓
Task 6 (Summary output) ←──────────┘
  ↓
Task 7 (Unit tests)
  ↓
Task 8 (Documentation)
  ↓
Task 9 (Final verification)
```

## Parallel Opportunities

- Tasks 4 and 5 can be done in parallel (independent implementations)
- Task 8 can start once Task 6 is complete (doesn't need tests)

## Estimated Total Time

- Task 1: 30 minutes
- Task 2: 1-2 hours
- Task 3: 2-3 hours
- Task 4: 2-3 hours
- Task 5: 1-2 hours
- Task 6: 30 minutes
- Task 7: 2-3 hours
- Task 8: 30 minutes
- Task 9: 30 minutes
- **Total:** ~11-15 hours

## Critical Success Factors

1. **Simplicity:** Keep implementation under 400 LOC
2. **Embedded FS:** Use existing `PluginFS`, no external file dependencies
3. **Tool Support:** Both Claude Code and OpenCode working correctly
4. **Comma-separated tools:** `--tool=claude,opencode` works
5. **Testing:** Comprehensive tests with real file operations
