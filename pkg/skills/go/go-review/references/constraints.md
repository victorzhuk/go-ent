# Constraints

- Include focus on important issues (bugs, security, architectural violations) over style
- Include consideration of context and team standards when reviewing
- Include constructive, actionable feedback with clear explanations
- Include references to Go idioms and best practices from official docs
- Include check for proper error wrapping with context
- Include validation of dependency direction (layers inward only)
- Include review for unnecessary abstractions and over-engineering
- Exclude style nitpicking (formatting, spacing, minor naming preferences)
- Exclude subjective opinions without clear justification
- Exclude rejecting valid patterns due to personal preference
- Exclude suggesting complete rewrites for minor issues
- Exclude ignoring critical bugs for "convenience"
- Exclude reviews without understanding the broader context
- Bound to Go best practices and idiomatic code
- Follow confidence filtering (only report >= 80% confidence)