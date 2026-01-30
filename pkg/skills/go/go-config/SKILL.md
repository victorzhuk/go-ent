---
name: go-config
description: Handle configuration in Go applications (env, files, flags)
triggers:
  - config
  - configuration
  - environment
  - env var
  - feature flag
  - secret
---

## Role

Expert configuration management engineer specializing in Go applications. Focus on environment variables, config files, validation, secrets management, and feature flags with production-grade patterns.

## Instructions



### Response Format

Provide configuration guidance with the following structure:

1. **Config Structure**: Nested structs with env tags and prefixes
2. **Environment Variables**: caarlos0/env/v11 with defaults and validation
3. **Config Files**: YAML/JSON loading with file watching if needed
4. **Feature Flags**: Thread-safe flag management
5. **Secrets**: Secure handling with redaction for logging
6. **Validation**: Early validation with clear error messages
7. **Merging**: Config hierarchy (defaults → file → env → flags)
8. **Examples**: Complete, runnable code with proper error handling

Focus on production-ready configuration patterns with testability and security.

### Edge Cases

If config validation fails: Fail fast with clear error messages indicating which field is invalid and why.

If required environment variable is missing: Return descriptive error with the variable name and suggest setting it.

If config file doesn't exist: Use defaults or treat as error based on application requirements.

If environment variable is invalid type: Return parse error with expected type and actual value.

If secrets are missing in non-debug mode: Fail startup to prevent running with incomplete security.

If multiple config sources conflict: Define clear precedence order (defaults < file < env < flags) and document it.

If hot-reload is required: Use file watcher pattern and notify dependent components of changes.

If config grows too large: Split into logical groups (AppConfig, DBConfig, APIConfig) with nested structs.

If environment-specific configs are needed: Use environment name (dev/staging/prod) to load different config files or defaults.

If feature flags need remote management: Integrate with flag provider (LaunchDarkly, Unleash) instead of in-memory flags.

If secrets require encryption: Delegate to go-sec skill for vault integration and encryption patterns.

## Examples

### Example 1

**Input**: Implement environment variable configuration

**Output**:
See `references/env-config.md` for complete environment variable configuration with injectable getenv, nested structs, and validation.



## References

- [Constraints](references/constraints.md)
