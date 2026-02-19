## Modern Python (3.12+)
- Use type hints everywhere: `def greet(name: str) -> str:`
- Use `match` statements for structural pattern matching
- Use f-strings for formatting: `f"Hello {name}"`
- Use `dataclasses` or `pydantic` for data structures
- Use `pathlib.Path` instead of `os.path`
- Use `enum.StrEnum` for string enumerations

## Project Structure
```
project/
├── src/
│   └── mypackage/
│       ├── __init__.py
│       ├── domain/
│       ├── service/
│       ├── api/
│       └── repo/
├── tests/
│   ├── unit/
│   ├── integration/
│   └── conftest.py
├── pyproject.toml      # Single config file (PEP 518)
├── uv.lock             # Lock file (if using uv)
└── Dockerfile
```

## Tooling
- **uv**: Fast package manager and resolver (preferred over pip)
- **ruff**: Linting and formatting (replaces black, isort, flake8)
- **mypy**: Static type checking with `--strict`
- **pytest**: Testing framework

## Type Hints
```python
from typing import TypeAlias, Protocol

UserID: TypeAlias = str

class Repository(Protocol):
    async def get(self, id: UserID) -> User | None: ...
    async def save(self, user: User) -> None: ...
```

## Async Patterns
```python
import asyncio

async def process_batch(items: list[Item]) -> list[Result]:
    tasks = [process_item(item) for item in items]
    return await asyncio.gather(*tasks, return_exceptions=True)

# Use async context managers for resources
async with aiohttp.ClientSession() as session:
    async with session.get(url) as resp:
        data = await resp.json()
```

## Error Handling
- Define custom exception hierarchy inheriting from base app exception
- Use `try/except` with specific exceptions — never bare `except:`
- Use `contextlib.suppress` for expected exceptions
- Log exceptions with traceback: `logger.exception("Failed")`
- Use `raise ... from e` for exception chaining

## Testing with pytest
```python
@pytest.fixture
def user_service(mock_repo):
    return UserService(repo=mock_repo)

def test_create_user(user_service):
    user = user_service.create(name="Alice", email="a@b.com")
    assert user.name == "Alice"
    assert user.email == "a@b.com"

@pytest.mark.parametrize("input,expected", [
    ("hello", "HELLO"),
    ("", ""),
])
def test_uppercase(input, expected):
    assert to_upper(input) == expected
```

## Best Practices
- Use virtual environments (venv/uv) — never install globally
- Pin dependencies in lock files
- Use `logging` module — configure at entry point, not in libraries
- Use `__all__` to control public API of modules
- Prefer composition over inheritance
