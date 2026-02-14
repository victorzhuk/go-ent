---
name: rust-web
description: Rust web development with Axum, SQLx, tower middleware, and production deployment patterns
---

# Rust Web Development

## Axum Framework
```rust
use axum::{Router, routing::get, extract::{Path, State, Json}};

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/users/:id", get(get_user).put(update_user))
        .route("/users", post(create_user))
        .layer(TraceLayer::new_for_http())
        .with_state(AppState::new().await);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8080").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn get_user(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<User>, AppError> {
    let user = state.user_service.get(id).await?;
    Ok(Json(user))
}
```

## SQLx (Compile-Time Verified SQL)
```rust
let user = sqlx::query_as!(User,
    "SELECT id, name, email FROM users WHERE id = $1", id
)
.fetch_optional(&pool)
.await?
.ok_or(AppError::NotFound)?;
```

## Tower Middleware Stack
- `tower-http::trace` for request tracing
- `tower-http::cors` for CORS configuration
- `tower-http::compression` for response compression
- Custom middleware with `tower::Layer` and `tower::Service`
- Rate limiting with `tower::limit`

## State Management
- Use `Arc<AppState>` for shared state
- Pool connections in state (sqlx::PgPool, redis ConnectionManager)
- Use `OnceCell` or `LazyLock` for expensive one-time initialization

## Production Checklist
- Graceful shutdown with `tokio::signal`
- Health endpoints (`/health`, `/ready`)
- Structured logging with `tracing` + `tracing-subscriber`
- OpenTelemetry integration for distributed tracing
- Cargo release profile: `lto = true`, `codegen-units = 1`, `strip = true`
