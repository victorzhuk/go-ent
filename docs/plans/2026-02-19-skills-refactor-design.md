# Skills Refactoring Design — Full Migration to Official Standard

**Date:** 2026-02-19
**Approach:** B — Full Structure Alignment
**Scope:** 70 SKILL.md files across 8 categories
**Goal:** Align `pkg/skills/` with the official Claude Code / Agent Skills standard

---

## Context

go-ent uses a custom skill schema (`triggers:` array, `## Role`, `## Instructions`, `## Examples` with `**Input**`/`**Output**`) that diverges from the official Claude Code schema. Skills activate via pure LLM reasoning against descriptions — no pattern matching or embedding search. This means description quality is the primary lever for skill discovery.

**Agent team findings (2026-02-19):**
- 70 SKILL.md files, ~7,221 lines across 8 categories
- Quality tiers: 15 production (Tier 1), 20 good (Tier 2), 25 functional (Tier 3), 10 stubs (Tier 4)
- Critical: `debug-core` has empty XML placeholder examples (unusable)
- 8 missing validation rules identified
- Community testing: optimised descriptions improve activation from ~20% → ~90%

---

## Section 1 — New SKILL.md Schema

### Frontmatter

```yaml
---
name: go-api
description: >
  This skill should be used when the user asks to design a REST API
  spec-first, generate Go server code from OpenAPI, work with ogen,
  define gRPC services, write protobuf schemas, or integrate HTTP and
  RPC transports in a Go service.
status: production          # production | draft | deprecated | delegated
version: 1                  # increment on breaking changes (optional)
---
```

**Changes from current format:**

| Field | Before | After |
|---|---|---|
| `triggers:` | Required array | Removed — description replaces it |
| `description` | Noun phrase ("Spec-first API design…") | Third-person trigger-specific prose |
| `status:` | Absent | Required: `production`, `draft`, `deprecated`, `delegated` |
| `version:` | Absent | Optional integer |

### Body Structure

```markdown
# Skill Name

## Overview
One paragraph — what this skill covers and its core principle. Imperative prose.

## When to Use
- Designing a new REST endpoint or resource
- Generating Go server stubs from an OpenAPI spec

**Not for:** general HTTP debugging (use debug-core), frontend API consumption.

## Quick Reference
| Task | Tool/Pattern |
|------|-------------|
| REST spec | OpenAPI 3.1 + ogen |
| RPC spec  | protobuf 3 + protoc |

## Implementation

### Spec-First Workflow
[imperative instructions, code blocks, subsections]

## Common Mistakes
- **Mixing business logic into handlers**: move to use-case layer
- **Error codes that don't map to HTTP semantics**: define contracts in spec

## Resources
- [`references/openapi-patterns.md`](references/openapi-patterns.md) — full OpenAPI examples
- [`examples/rest-service/`](examples/rest-service/) — working ogen-generated service
```

**Section changes:**

| Before | After | Reason |
|---|---|---|
| `## Role` | Removed, merged into `## Overview` | Not in official schema |
| `## Instructions` | `## Implementation` + descriptive subsections | Clearer, imperative form |
| `## Examples` (Input/Output) | Moved to `examples/` directory | Progressive disclosure |
| `## Edge Cases` | `## Common Mistakes` | More actionable |
| `## References` | `## Resources` with explicit file paths | Clarity |

**Body style:** All text imperative ("Write a handler", not "You are an expert in…" or "You should…").

### Directory Structure

```
pkg/skills/{category}/{skill-name}/
├── SKILL.md          ← lean: ≤2,000 words
├── references/       ← deep content (Tier 1–2 skills only)
│   └── patterns.md
└── examples/         ← runnable files (Tier 1–2 skills only)
    └── main.go
```

---

## Section 2 — Validation Rules

### Rules Removed

| Rule | Reason |
|---|---|
| `triggers:` required | Replaced by description-driven activation |
| `## Role` required | Not in official schema |
| `## Examples` with `**Input**`/`**Output**` required | Moved to `examples/` directory |

### Rules Kept

| Rule | Notes |
|---|---|
| `name` required, kebab-case, ≤64 chars | Unchanged |
| `## Implementation` ≥200 chars with subsections or code blocks | Was `## Instructions` |

### New Blocking Rules

**Description quality:**
- Max 1024 chars (official schema limit)
- Must contain trigger signal: one of `when`, `use when`, `asks to`, `should be used`
- Warning if contains workflow summary words (`first`, `then`, `step 1`, `→`)

**Status field:**
- Required, must be one of: `production`, `draft`, `deprecated`, `delegated`

**Placeholder detection:**
- Error if content contains empty XML tags (e.g. `<example><input></input></example>`)

**Name matches directory:**
- `pkg/skills/go/go-api/SKILL.md` → `name` must equal `go-api`

**Reference file existence:**
- All `[text](path)` links in `## Resources` section must point to existing files

### New Warning Rules

| Warning | Condition |
|---|---|
| SKILL.md too long | >2,000 words → suggest `references/` |
| `triggers:` still present | Deprecated field |
| `## Role` still present | Deprecated section |
| `status: draft` with >150 lines | Consider promoting to production |
| Description starts with "I " | First-person detected |
| Description starts with "Use this skill" | Second-person — rewrite in third person |

### Required Body Sections (updated)

| Section | Status | Min length |
|---|---|---|
| `## Overview` | Required | 50 chars |
| `## When to Use` | Required | 30 chars |
| `## Implementation` | Required | 200 chars + subsections or code blocks |
| `## Resources` | Optional | — |
| `## Common Mistakes` | Warning if absent on `status: production` skills | — |

---

## Section 3 — Migration Strategy

### Validator Transition Mode

Update `rules.go` first with a `strictMode` flag. Old sections (`## Role`, `triggers:`) become warnings during migration, errors after.

```go
const strictMode = false  // flip to true after all 70 skills migrated

func validate(skill Skill) Result {
    // Always enforced (new rules)
    enforceDescriptionQuality(skill)
    enforceNameMatchesDir(skill)
    enforcePlaceholderFree(skill)
    enforceStatus(skill)

    if strictMode {
        enforceNoRoleSection(skill)
        enforceNoTriggersField(skill)
        enforceRequiredSections(skill) // Overview, When to Use, Implementation
    } else {
        warnIfRoleSection(skill)
        warnIfTriggersPresent(skill)
    }
}
```

### Migration Batches

All skill batches are independent and can run in parallel after Batch 0.

```
Batch 0 (validator)  ──┐
                        ▼
Batch 1 (stubs)  ─┐
Batch 2 (lang)   ─┤  parallel  ──→  Batch 6 (strictMode)
Batch 3 (core)   ─┤
Batch 4 (infra)  ─┤
Batch 5 (go/ent) ─┘
```

| Batch | Target | Skills | Tier | Work |
|---|---|---|---|---|
| 0 | `rules.go`, `parser.go` | — | — | Add new rules, transition mode |
| 1 | `agent/`, `backend/` stubs | 10 | 4 | Frontmatter only (status: draft, description rewrite, remove triggers) |
| 2 | `lang/` | 17 | 3–4 | Frontmatter + section renames + imperative form |
| 3 | `core/` | 22 | 2–3 | Full body restructure + descriptions + references/ for heavy skills |
| 4 | `infra/`, `backend/`, `qa/` | 15 | 2–3 | Full restructure + selective references/ + examples/ |
| 5 | `go/`, `ent/` | 20 | 1–2 | Full migration + references/ + examples/ for all |
| 6 | `rules.go` | — | — | Flip strictMode = true, delete transition code |

### Definition of Done (per skill)

- [ ] `status:` field present and valid
- [ ] `description` passes all validator rules (trigger signal, ≤1024 chars, no workflow summary)
- [ ] `triggers:` removed from frontmatter
- [ ] `## Role` removed, content in `## Overview`
- [ ] `## Instructions` → `## Implementation` (or descriptive subsections)
- [ ] `## Examples` removed from body (moved to `examples/` if worth keeping)
- [ ] All body text imperative form
- [ ] `## When to Use` section present with explicit "Not for:" clause
- [ ] No empty XML placeholder tags
- [ ] All `## Resources` links point to existing files
- [ ] Validator passes with zero errors

---

## Section 4 — Naming Convention & Canonical Template

### Naming Rules by Category

| Category | Pattern | Renames required |
|---|---|---|
| `core/` | `{topic}` — no prefix, no `-core` suffix | `arch-core` → `architecture`, `debug-core` → `debugging`, `security-core` → `security` |
| `go/` | `go-{topic}` | None — already consistent |
| `lang/` | `{lang}-{topic}` | `flutter` → `flutter-core`, `vue` → `vue-core`, `tailwind` → `tailwind-css` |
| `backend/` | `{topic}` | None |
| `infra/` | `{topic}` | None |
| `qa/` | `qa-{layer}` | None — already consistent |
| `agent/` | `{topic}` | None |
| `ent/` | `ent-{topic}` | None — already consistent |

**Total: 6 renames** (3 in `core/`, 3 in `lang/`). Validator enforces pattern per category.

### Canonical Template

Saved to `pkg/templates/standard/SKILL.md`. Used by `go-ent skill new` wizard.

```markdown
---
name: ${SKILL_NAME}
description: >
  This skill should be used when the user asks to ${TRIGGER_PHRASE_1},
  ${TRIGGER_PHRASE_2}, or mentions ${TRIGGER_KEYWORD_1}, ${TRIGGER_KEYWORD_2}.
status: draft
---

# ${SKILL_TITLE}

## Overview
${ONE_PARAGRAPH_WHAT_AND_WHY}

## When to Use
- ${USE_CASE_1}
- ${USE_CASE_2}
- ${USE_CASE_3}

**Not for:** ${ANTI_USE_CASE} — use [${OTHER_SKILL}](../${OTHER_SKILL}/) instead.

## Quick Reference
| Task | Approach |
|------|----------|
| ${TASK_1} | ${APPROACH_1} |
| ${TASK_2} | ${APPROACH_2} |

## Implementation

### ${SUBSECTION_1}
${IMPERATIVE_INSTRUCTIONS}

\`\`\`go
${CODE_EXAMPLE}
\`\`\`

### ${SUBSECTION_2}
${IMPERATIVE_INSTRUCTIONS}

## Common Mistakes
- **${MISTAKE_1}**: ${FIX_1}
- **${MISTAKE_2}**: ${FIX_2}

## Resources
- [`references/patterns.md`](references/patterns.md) — ${DESCRIPTION}
- [`examples/`](examples/) — ${DESCRIPTION}
```

### Template Variables

Documented in `pkg/templates/README.md`:

| Variable | Required | Description |
|---|---|---|
| `${SKILL_NAME}` | Yes | kebab-case, must match directory name |
| `${SKILL_TITLE}` | Yes | Title case display name |
| `${TRIGGER_PHRASE_1}` | Yes | Verb phrase: "design a REST API", "debug a goroutine leak" |
| `${TRIGGER_KEYWORD_1}` | Yes | Noun keyword: "OpenAPI", "ogen", "gRPC" |
| `${ONE_PARAGRAPH_WHAT_AND_WHY}` | Yes | Imperative prose, ≤150 words |
| `${ANTI_USE_CASE}` | Recommended | What this skill is NOT for |
| `${OTHER_SKILL}` | Recommended | Cross-reference for the anti-use case |

### Wizard Update

`go-ent skill new` prompts explicitly for:
1. Trigger phrases (verb form: "design a…", "debug a…")
2. Trigger keywords (noun form: "OpenAPI", "gRPC")
3. Anti-use case + recommended alternative skill

---

## Summary of Changes

| Area | Change |
|---|---|
| Frontmatter | Remove `triggers:`, add `status:` (required), rewrite `description` |
| Body sections | Remove `## Role`, rename `## Instructions` → `## Implementation`, add `## Overview` + `## When to Use` |
| Body style | Imperative form throughout, no second/first person |
| Examples | Move from body to `examples/` directory |
| Deep content | Move from body to `references/` directory (Tier 1–2 skills) |
| Validation | 5 new blocking rules, 6 new warning rules, 3 rules removed |
| Naming | 6 renames, category pattern enforced by validator |
| Template | New canonical template + documented variables + updated wizard |
| Migration | Batch 0 (validator) → Batches 1–5 in parallel → Batch 6 (strict mode) |
