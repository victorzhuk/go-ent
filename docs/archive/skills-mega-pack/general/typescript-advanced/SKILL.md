---
name: typescript-advanced
description: Advanced TypeScript patterns, utility types, generics, type guards, and strict configuration
---

# Advanced TypeScript

## Strict Configuration
```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "exactOptionalPropertyTypes": true,
    "noPropertyAccessFromIndexSignature": true,
    "forceConsistentCasingInFileNames": true
  }
}
```

## Utility Types
```typescript
// Extract/Exclude for union manipulation
type Status = 'active' | 'inactive' | 'pending';
type ActiveStatus = Extract<Status, 'active' | 'pending'>; // 'active' | 'pending'

// Branded types for type safety
type UserId = string & { readonly __brand: 'UserId' };
function createUserId(id: string): UserId { return id as UserId; }

// Discriminated unions
type Result<T> =
  | { success: true; data: T }
  | { success: false; error: string };

// Mapped types
type Readonly<T> = { readonly [K in keyof T]: T[K] };
type Optional<T> = { [K in keyof T]?: T[K] };
```

## Generics Patterns
```typescript
// Constrained generics
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// Generic factories
function createService<T extends new (...args: any[]) => any>(
  ServiceClass: T
): InstanceType<T> {
  return new ServiceClass();
}

// Conditional types
type IsArray<T> = T extends any[] ? true : false;
type Unwrap<T> = T extends Promise<infer U> ? U : T;
```

## Type Guards
```typescript
function isUser(value: unknown): value is User {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    'email' in value
  );
}

// Assertion functions
function assertDefined<T>(value: T | null | undefined): asserts value is T {
  if (value == null) throw new Error('Expected defined value');
}
```

## Error Handling with Types
```typescript
// Type-safe error handling
class AppError extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly statusCode: number = 500,
  ) {
    super(message);
  }
}

// Result type pattern
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };
```
