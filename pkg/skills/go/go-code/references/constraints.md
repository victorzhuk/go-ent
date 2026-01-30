# Constraints

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