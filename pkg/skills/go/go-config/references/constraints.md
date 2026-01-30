# Constraints

- Include environment variable parsing with caarlos0/env/v11
- Include validation with custom Validate() method or validator/v10
- Include configuration file loading (YAML/JSON)
- Include feature flags implementation
- Include secrets handling (defer security details to go-sec)
- Include config redaction for logging
- Include configuration hierarchy (defaults → file → env → flags)
- Exclude hardcoding secrets in code
- Exclude committing secrets to version control
- Exclude using global config objects (pass explicitly)
- Exclude parsing environment variables directly with os.Getenv
- Exclude mixing config loading with business logic
- Include injectable getenv for testing
- Always validate config after loading