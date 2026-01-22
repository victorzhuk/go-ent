# Skill Lint CI/CD Integration

This guide explains how to integrate the `skill lint` command into your CI/CD pipeline using GitHub Actions.

## Overview

The `skill lint` command validates skill and command files for common issues and can auto-fix many problems automatically. The CI workflow ensures all skills meet quality standards before merging.

## Workflow File

The GitHub Actions workflow is located at `.github/workflows/skill-lint.yml`.

### Triggers

The workflow runs automatically on:
- **Push to main**: Validates all skill files
- **Pull requests to main**: Validates changed skill files
- **Manual dispatch**: Triggers auto-fix workflow with user input

The workflow only runs when skill/command files or the workflow file itself changes.

### Jobs

The workflow has two jobs:

#### 1. Validation-Only Job

Runs on every push and pull request to validate skill files without making changes.

```yaml
validate:
  runs-on: ubuntu-latest
  steps:
    - Checkout code
    - Set up Go
    - Build go-ent
    - Lint skills (validation only)
    - Upload lint results as artifact
```

**Exit codes:**
- `0`: No errors or warnings (all files valid)
- `1`: Validation errors found (workflow fails)

#### 2. Auto-Fix Job

Runs only on manual dispatch with `fix=true`. Automatically fixes common issues.

```yaml
autofix:
  runs-on: ubuntu-latest
  steps:
    - Checkout code
    - Set up Go
    - Build go-ent
    - Run skill lint with --fix flag
    - Upload lint results as artifact
    - Commit auto-fixed changes (if any)
    - Comment on PR with results
```

**Exit codes:**
- `0`: All issues fixed successfully
- `1`: Errors remain that cannot be auto-fixed

## Exit Codes Reference

| Code | Meaning | CI Behavior |
|------|---------|-------------|
| `0` | Success (no issues) | Workflow passes |
| `1` | Validation errors | Workflow fails |
| `2` | Invalid arguments | Workflow fails immediately |
| `3` | File not found | Workflow fails immediately |

## Dry Run Mode

The `--dry-run` flag lets you preview what would be fixed without modifying any files.

```bash
# Preview fixes without making changes
ent skill lint --dry-run

# Preview with JSON output
ent skill lint --dry-run --json
```

### Dry Run Output

Dry run mode shows:
- What changes would be applied
- Color-coded diff of changes (green = added, red = removed)
- List of affected rules

Example output:

```
🔍 go-code (dry-run):
  • normalized YAML frontmatter (sorted keys, formatted indentation)
  • added 2 trigger suggestion(s) based on skill name and location

  Diff:
  - version: "1.0.0"
  - name: go-code
  + name: go-code
  + triggers:
  +   - patterns:
  +       - "go.*"
  +     weight: 0.7
  +   - patterns:
  +       - "code.*"
  +     weight: 0.7
  + description: Go coding patterns
  - description: Go coding patterns
```

### When to Use Dry Run

- **Before auto-fixing**: Preview what changes will be made
- **CI pipelines**: Check what would change without side effects
- **Code review**: Show reviewers what the fix will do
- **Safety**: Verify changes are correct before applying

### Dry Run vs Auto-Fix

| Feature | --dry-run | --fix |
|----------|-----------|--------|
| Modifies files | No | Yes |
| Shows changes | Yes | Yes |
| Color diff | Yes | No |
| Exit codes | 0 if no issues, 1 if issues found | 0 if fixes succeed, 1 if fail |

## Usage Scenarios

### Scenario 1: Validation-Only (PR Checks)

Use this for pull request checks. It validates without making changes.

**When to use:**
- Every PR that changes skill/command files
- Protecting main branch from invalid skills
- Automated quality gates

**Benefits:**
- No side effects on the codebase
- Quick feedback to developers
- Blocks merging of invalid skills

**Example:**

You create a PR with a new skill. The workflow:

1. Runs automatically when you push
2. Validates the new skill file
3. Fails if issues are found, blocking merge
4. Passes if valid, allowing merge

### Scenario 2: Auto-Fix (Manual Trigger)

Use this for bulk fixing or maintenance. It automatically fixes issues.

**When to use:**
- Fixing multiple skills after updating standards
- Fixing formatting issues across the codebase
- Maintenance tasks requiring auto-fix

**Benefits:**
- Automatic fixes for common issues
- Commits fixes back to the branch
- Reduces manual work

**Example:**

You update the skill format standards and need to update existing skills:

1. Navigate to Actions tab in GitHub
2. Select "Skill Lint" workflow
3. Click "Run workflow"
4. Set `fix` to `true`
5. Click "Run workflow"

The workflow:
1. Runs `skill lint --fix`
2. Fixes common issues automatically
3. Commits fixes with `[skip ci]` tag
4. Comments on PR with results

## JSON Output Format

When using `--json` flag, the output includes:

```json
{
  "errors": ["file.md: missing required field: name"],
  "warnings": ["file.md: description is too short"],
  "fixed": ["file.md: normalized frontmatter format"]
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `errors` | string[] | Validation errors (workflow fails if present) |
| `warnings` | string[] | Non-blocking warnings |
| `fixed` | string[] | Issues that were auto-fixed |

## Artifacts

Each workflow run uploads lint results as artifacts:

- **Validation job**: `lint-results-{sha}.json`
- **Auto-fix job**: `lint-results-autofix-{sha}.json`

Artifacts are retained for 30 days and can be downloaded from the workflow run page.

## Customization

### Change Triggers

Modify the `on` section in the workflow file to change when it runs:

```yaml
on:
  push:
    branches: [main, develop]  # Add more branches
    paths:
      - 'plugins/**/*.md'       # Change path filters
  pull_request:
    branches: [main]
```

### Change Go Version

Modify the Go version:

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.24'  # Update to newer version
```

### Add More Validation Steps

Add additional checks to the workflow:

```yaml
- name: Run custom checks
  run: |
    # Add your custom checks here
    echo "Running additional validation..."
```

### Disable Auto-Fix Job

If you don't need auto-fix, remove the `autofix` job and `workflow_dispatch` trigger.

## Troubleshooting

### Workflow Fails But Local Lint Passes

**Issue**: CI fails but `skill lint` passes locally.

**Possible causes:**
1. Different Go version
2. Different file paths in CI
3. Environment differences

**Solutions:**
1. Check the workflow log for exact error
2. Download the lint results artifact
3. Compare local vs CI environment

### Auto-Fix Doesn't Commit Changes

**Issue**: Auto-fix job runs but doesn't commit changes.

**Possible causes:**
1. No changes needed (files already valid)
2. Git configuration issues
3. Permission issues

**Solutions:**
1. Check the workflow log for "No changes to commit"
2. Verify GitHub token has write permissions
3. Check the lint results artifact

### Workflow Runs Too Slow

**Issue**: Workflow takes too long to complete.

**Possible causes:**
1. Building go-ent from source every time
2. Too many files to lint

**Solutions:**
1. Cache Go dependencies
2. Build go-ent in a separate workflow and cache
3. Use path filters to only lint changed files

## Best Practices

1. **Always run on PRs**: Use validation-only job to catch issues before merge
2. **Use manual trigger for auto-fix**: Avoid accidental auto-fixes on branches
3. **Review auto-fix commits**: Even auto-fixed changes should be reviewed
4. **Keep exit codes in mind**: Design your pipeline around the exit code meanings
5. **Use artifacts for debugging**: Download lint results when investigating failures

## Example: Complete CI Setup

Here's a complete setup combining validation, auto-fix, and other checks:

```yaml
name: Complete CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  # Validation-only (runs on every PR)
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        run: make build

      - name: Skill lint
        run: ./bin/go-ent skill lint --json plugins/

      - name: Go tests
        run: make test

      - name: Go lint
        run: make lint

  # Auto-fix (manual trigger)
  autofix:
    runs-on: ubuntu-latest
    if: github.event_name == 'workflow_dispatch'
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        run: make build

      - name: Auto-fix skills
        run: ./bin/go-ent skill lint --fix --json plugins/

      - name: Commit changes
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add plugins/
          git commit -m "chore: auto-fix skill linting issues [skip ci]" || true
          git push || true
```

## Related Documentation

- [Skill Authoring Guide](SKILL-AUTHORING.md) - Creating and maintaining skills
- [CLI Examples](CLI_EXAMPLES.md) - Using the skill lint command locally
- [Development Guide](DEVELOPMENT.md) - Developing go-ent itself
