## Ownership & Borrowing
- Prefer borrowing (`&T`) over ownership transfer unless the function needs to own the data
- Use `Clone` sparingly — it's a code smell if overused
- Prefer `&str` over `String` in function parameters
- Use `Cow<'a, str>` when you might or might not need to allocate
- Understand move semantics: non-Copy types are moved, not copied

## Error Handling
```rust
// Define domain errors with thiserror
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("not found: {0}")]
    NotFound(String),
    #[error("validation: {field}: {message}")]
    Validation { field: String, message: String },
    #[error(transparent)]
    Database(#[from] sqlx::Error),
    #[error(transparent)]
    Unexpected(#[from] anyhow::Error),
}

// Use Result<T, E> everywhere — never panic in library code
// Use ? operator for propagation
// Use anyhow for application code, thiserror for libraries
```

## Async Programming
- Use `tokio` as the async runtime (or `async-std` for simpler use cases)
- Prefer `tokio::spawn` for concurrent tasks
- Use `tokio::select!` for racing multiple futures
- Avoid blocking operations in async contexts — use `spawn_blocking`
- Use `async fn` in traits with `async-trait` or native async traits (Rust 1.75+)

## Traits & Generics
- Use trait bounds: `fn process<T: Display + Send>(item: T)`
- Prefer `impl Trait` in argument position for simple cases
- Use associated types for traits with one logical output type
- Use `dyn Trait` for runtime polymorphism (trait objects)
- Implement `From`/`Into` for type conversions instead of custom methods

## Project Structure
```
src/
├── main.rs         # Entry point
├── lib.rs          # Library root
├── config.rs       # Configuration
├── domain/         # Business types and traits
├── service/        # Business logic
├── handler/        # HTTP handlers (axum/actix)
├── repo/           # Database access
└── error.rs        # Error types
```

## Safety & Best Practices
- Use `clippy` with `-W clippy::pedantic` in CI
- Run `cargo fmt` — non-negotiable
- Minimize `unsafe` — document invariants when unavoidable
- Prefer `const` generics and `const fn` where applicable
- Use `#[must_use]` on types/functions where ignoring return is likely a bug
- Use `#[non_exhaustive]` on public enums for future compatibility

## Testing
- Use `#[cfg(test)]` module in the same file for unit tests
- Integration tests in `tests/` directory
- Use `proptest` or `quickcheck` for property-based testing
- Use `mockall` for mocking traits in tests
- `cargo test -- --nocapture` for debug output
