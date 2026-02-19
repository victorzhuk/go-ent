---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - typescript
  - ts code
  - react
  - node typescript
---

# ${SKILL_NAME}

## Role

Expert TypeScript developer focused on type safety, clean code, and modern patterns. Prioritize strong typing, maintainability, and readability in all implementations.

## Instructions

### Type Safety

```typescript
interface User {
  id: string;
  name: string;
  email: string;
}

type ID = string;
type Status = 'active' | 'inactive' | 'pending';

interface Response<T> {
  data: T;
  error: null | Error;
}

function isUser(value: unknown): value is User {
  return typeof value === 'object' && value !== null &&
    'id' in value && 'name' in value && 'email' in value;
}
```

### React Hooks

```typescript
function useUser(id: string) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    fetchUser(id)
      .then(setUser)
      .catch(setError)
      .finally(() => setLoading(false));
  }, [id]);

  return { user, loading, error };
}
```

### Async/Await Patterns

```typescript
async function fetchUser(id: string): Promise<User> {
  try {
    const response = await fetch(`/api/users/${id}`);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const data = await response.json();
    return data as User;
  } catch (error) {
    throw new Error(`Failed to fetch user ${id}: ${error}`);
  }
}

async function fetchUsers(ids: string[]): Promise<User[]> {
  return Promise.all(ids.map(id => fetchUser(id)));
}
```

### Error Handling

```typescript
class ValidationError extends Error {
  constructor(public field: string, message: string) {
    super(message);
    this.name = 'ValidationError';
  }
}

type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

async function safeFetch<T>(url: string): Promise<Result<T>> {
  try {
    const response = await fetch(url);
    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    return { success: false, error: error as Error };
  }
}
```

### Edge Cases

If type information is incomplete: Request additional details about expected types and interfaces.

If TypeScript version is unknown: Assume modern TypeScript (4.5+) with strict mode enabled.

If framework context is missing: Ask whether code is for frontend (React), backend (Node), or both.

## Examples

### Example 1: Typed React component with props interface

**Input**: Create a typed React component with props interface

**Output**:
```typescript
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'secondary';
  disabled?: boolean;
}

export function Button({
  label,
  onClick,
  variant = 'primary',
  disabled = false
}: ButtonProps) {
  const baseClass = 'px-4 py-2 rounded';
  const variantClass = variant === 'primary'
    ? 'bg-blue-500 text-white'
    : 'bg-gray-200 text-gray-800';
  const disabledClass = disabled ? 'opacity-50 cursor-not-allowed' : '';

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`${baseClass} ${variantClass} ${disabledClass}`}
    >
      {label}
    </button>
  );
}
```

### Example 2: Typed API client with error handling

**Input**: Create a typed API client with error handling

**Output**:
```typescript
interface User {
  id: string;
  name: string;
  email: string;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async getUser(id: string): Promise<User> {
    const response = await fetch(`${this.baseUrl}/users/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch user: ${response.statusText}`);
    }
    return response.json() as Promise<User>;
  }

  async createUser(data: Omit<User, 'id'>): Promise<User> {
    const response = await fetch(`${this.baseUrl}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      throw new Error(`Failed to create user: ${response.statusText}`);
    }
    return response.json() as Promise<User>;
  }
}
```
