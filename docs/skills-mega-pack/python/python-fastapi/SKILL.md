---
name: python-fastapi
description: FastAPI development with Pydantic v2, async SQLAlchemy, dependency injection, and production patterns
---

# FastAPI Development

## Application Structure
```python
from fastapi import FastAPI, Depends, HTTPException, status
from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await init_db()
    yield
    # Shutdown
    await close_db()

app = FastAPI(title="My API", lifespan=lifespan)
```

## Pydantic v2 Models
```python
from pydantic import BaseModel, Field, ConfigDict

class UserCreate(BaseModel):
    model_config = ConfigDict(strict=True)
    name: str = Field(min_length=1, max_length=100)
    email: str = Field(pattern=r'^[\w.-]+@[\w.-]+\.\w+$')

class UserResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: uuid.UUID
    name: str
    email: str
    created_at: datetime
```

## Dependency Injection
```python
async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session() as session:
        yield session

async def get_user_service(db: AsyncSession = Depends(get_db)) -> UserService:
    return UserService(UserRepository(db))

@app.post("/users", status_code=status.HTTP_201_CREATED)
async def create_user(
    data: UserCreate,
    service: UserService = Depends(get_user_service),
) -> UserResponse:
    return await service.create(data)
```

## Error Handling
```python
from fastapi import Request
from fastapi.responses import JSONResponse

class AppException(Exception):
    def __init__(self, status_code: int, code: str, message: str):
        self.status_code = status_code
        self.code = code
        self.message = message

@app.exception_handler(AppException)
async def handle_app_error(request: Request, exc: AppException):
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": {"code": exc.code, "message": exc.message}},
    )
```

## Async SQLAlchemy
- Use `sqlalchemy[asyncio]` with `asyncpg` driver
- Use `async_sessionmaker` for session factory
- Use `select()` statements, not legacy Query API
- Always use `async with` for sessions

## Best Practices
- Use `response_model` to control output serialization
- Use `status_code` parameter for proper HTTP codes
- Use `tags` for OpenAPI grouping
- Add `middleware` for CORS, timing, request ID
- Use `BackgroundTasks` for fire-and-forget operations
- Use `APIRouter` for modular route organization
