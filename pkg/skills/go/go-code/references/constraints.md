# Constraints

## General Code Patterns
- Include clean, idiomatic Go code following standard conventions
- Include proper error wrapping with context using `%w` verb
- Include context propagation as first parameter throughout layers
- Include domain entities with zero external dependencies
- Include dependency injection pattern (accept interfaces, return structs)
- Exclude magic numbers (use named constants instead)
- Exclude global mutable state (pass dependencies explicitly)
- Exclude panic in production code (use error handling instead)
- Exclude over-engineering and premature abstractions (YAGNI)
- Exclude AI-style verbose naming and unnecessary comments
- Bound to clean layered architecture: Transport → UseCase → Domain ← Repository
- Follow DI pattern with explicit dependency graphs
- Keep domain layer pure with no external dependencies

## Configuration Patterns
- Include environment variable parsing with caarlos0/env/v11
- Include validation with custom Validate() method or validator/v10
- Include configuration file loading (YAML/JSON)
- Include feature flags implementation
- Include secrets handling (defer security details to go-sec)
- Include config redaction for logging
- Include configuration hierarchy (defaults → file → env → flags)
- Include injectable getenv for testing
- Exclude hardcoding secrets in code
- Exclude committing secrets to version control
- Exclude using global config objects (pass explicitly)
- Exclude parsing environment variables directly with os.Getenv
- Exclude mixing config loading with business logic
- Always validate config after loading