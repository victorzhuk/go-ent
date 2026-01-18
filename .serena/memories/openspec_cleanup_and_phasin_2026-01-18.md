# OpenSpec Cleanup and Execution Engine v2 Phasing - January 18, 2026

## ✅ Actions Completed

### 1. Archived 9 Stale Proposals
**Reason**: All had 0% progress for >30 days, creating registry overhead

**Archived Changes**:
1. add-context-memory (24 tasks, 0%) - Lacks clear dependencies
2. add-dynamic-mcp-discovery (69 tasks, 0%) - Complex MCP routing, not prioritized  
3. add-spec-anchoring (47 tasks, 0%) - Spec evolution features, future work
4. align-flow-architecture (166 tasks, 0%) - Large architectural change, needs planning
5. fix-opencode-agent-metadata-parsing (0 tasks) - Empty proposal, no tasks defined
6. integrate-driver-into-adapters (9 tasks, 0%) - Driver integration, lower priority
7. reorganize-plugin-source-layout (11 tasks, 0%) - Source layout changes, not critical
8. split-marketplace-plugins-by-type (16 tasks, 0%) - Marketplace features, future work
9. upgrade-template-engine (12 tasks, 0%) - Template improvements, nice-to-have

**Archive Structure**:
```
openspec/changes/archive/
└── 2026-01-18-{proposal-name}/
    ├── proposal.md
    ├── tasks.md  
    ├── specs/ (if present)
    └── archive.md (summary)
```

**Registry Impact**:
- **Before**: 16 active changes, 2 archived
- **After**: 7 active changes, 11 archived
- **Reduction**: 56% fewer active changes to manage

### 2. Rebuilt execution-engine-v2 with Phased Approach

**Problem**: 100 tasks in single monolithic block, 0% progress, blocking critical work

**Solution**: Broken into 4 manageable phases with clear dependencies and timelines

#### New Phase Structure:

**Phase 1: Foundation (28 tasks)** ⭐ READY TO START
- Unit tests for existing v2 features (sandbox, code-mode)
- **No dependencies** - can start immediately
- **Effort**: 10-12 hours
- **Outcome**: Validates existing implementation

**Phase 2: Context Management (28 tasks)** 🚀 AFTER PHASE 1
- Context summarization and limit handling
- **Depends on**: Phase 1 completion
- **Effort**: 12-15 hours  
- **Outcome**: Core v2 functionality working

**Phase 3: State Persistence (36 tasks)** ⚠️ DEPENDENT
- Execution state persistence and interrupt/resume
- **Depends on**: Phase 2 + add-boltdb-state-system*
- **Effort**: 12-15 hours
- **Outcome**: Full fault tolerance (*may use file fallback)

**Phase 4: Integration (16 tasks)** 🎯 FINAL PHASE
- End-to-end testing and validation
- **Depends on**: All previous phases
- **Effort**: 6-8 hours
- **Outcome**: Production-ready v2 features

#### Key Improvements:

**Risk Mitigation**:
- Phase 1 has no dependencies - immediate progress possible
- Phase 3 has file-based fallback if add-boltdb-state-system remains blocked
- Phases 1+2+4 can deliver 72% of value without Phase 3

**Clear Dependencies**:
- Each phase has explicit prerequisites
- Cross-phase dependency graph documented
- Fallback strategies defined

**Realistic Timeline**:
- **Week 1**: Phase 1 (Foundation)
- **Week 2**: Phase 2 (Context Management)  
- **Week 3**: Phase 3 (State Persistence)
- **Week 4**: Phase 4 (Integration)

## 📊 Impact Analysis

### Registry Health
- **Active Changes**: Reduced from 16 to 7 (56% reduction)
- **Focus**: Critical path changes now visible
- **Management Overhead**: Significantly reduced

### Dependency Graph
- **Critical Path**: Clear progression from Foundation → Integration
- **Blockers**: Identified and mitigated with fallbacks
- **Parallel Work**: Possible within phases

### Development Velocity
- **Immediate Wins**: Phase 1 can start today (28 tasks)
- **Incremental Value**: Each phase delivers working functionality
- **Risk Reduction**: Smaller, manageable task blocks

## 🎯 Next Steps

### Immediate (Today)
1. **Start execution-engine-v2 Phase 1** - Unit tests for existing features
2. **Begin add-acp-agent-mode** - Now unblocked and ready

### This Week  
1. **Complete Phase 1** - Foundation unit tests
2. **Start Phase 2** - Context management implementation

### Next Week
1. **Complete Phase 2** - Context summarization working
2. **Assess Phase 3 blockers** - Determine if file fallback needed

## 📈 Success Metrics

### Short-term (1 week)
- [ ] execution-engine-v2 Phase 1: 28 tasks completed
- [ ] add-acp-agent-mode: Progress started
- [ ] Active changes: 7 (maintained focus)

### Medium-term (4 weeks)  
- [ ] execution-engine-v2: All 4 phases complete (108 tasks)
- [ ] add-acp-agent-mode: 60 tasks complete
- [ ] Critical path: execution-engine-v2 → add-acp-agent-modes unblocked

### Long-term (impact)
- **ACP Agent Mode**: Enables OpenCode worker orchestration
- **Context Management**: Handles long-running executions
- **State Persistence**: Enables fault-tolerant workflows
- **Integration**: Production-ready v2 execution engine

**Overall Result**: Clean, focused registry with clear execution path for critical infrastructure completion.