# Constraints

- Include error wrapping with fmt.Errorf and %w verb
- Include custom error types with Error() and optional Is()/Unwrap() methods
- Include sentinel errors using errors.New()
- Include error chain inspection with errors.Is() and errors.As()
- Include domain-specific error patterns
- Include error testing patterns
- Exclude wrapping errors without adding context
- Exclude double logging (log and wrap)
- Exclude checking errors twice (handle once)
- Exclude using error messages for control flow (use error types)
- Exclude returning error without context in outer layers
- Exclude creating custom errors when sentinel errors suffice