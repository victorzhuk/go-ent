## Component Patterns
```tsx
// Functional components only — no class components
interface UserCardProps {
    user: User;
    onEdit?: (id: string) => void;
}

export function UserCard({ user, onEdit }: UserCardProps) {
    return (
        <div className="card">
            <h3>{user.name}</h3>
            <p>{user.email}</p>
            {onEdit && <button onClick={() => onEdit(user.id)}>Edit</button>}
        </div>
    );
}
```

## Hooks Best Practices
- `useState`: Keep state minimal and derived values computed
- `useEffect`: Use for synchronization with external systems ONLY
- `useMemo`/`useCallback`: Use when passing to memoized children or expensive computation
- `useRef`: For DOM references and mutable values that don't trigger re-render
- Custom hooks: Extract reusable logic; name with `use` prefix

## State Management
- Local state: `useState` for component-specific state
- Shared state: React Context for small app state, Zustand/Jotai for complex
- Server state: TanStack Query (React Query) for data fetching + caching
- URL state: Use router params/search params as state

## Performance
- Use React DevTools Profiler to identify unnecessary renders
- Memoize expensive components with `React.memo`
- Use `useDeferredValue` for non-urgent updates
- Use `useTransition` for expensive state transitions
- Lazy load routes/components with `React.lazy` + `Suspense`
- Virtualize long lists with `react-window` or `tanstack-virtual`

## File Structure
```
src/
├── components/       # Shared UI components
│   ├── Button/
│   │   ├── Button.tsx
│   │   ├── Button.test.tsx
│   │   └── index.ts
├── features/         # Feature-based modules
│   └── users/
│       ├── components/
│       ├── hooks/
│       ├── api/
│       └── types.ts
├── hooks/            # Shared custom hooks
├── lib/              # Utilities, API client
└── types/            # Shared types
```

## Testing
```tsx
import { render, screen, userEvent } from '@testing-library/react';

test('calls onEdit when button clicked', async () => {
    const onEdit = vi.fn();
    render(<UserCard user={mockUser} onEdit={onEdit} />);
    await userEvent.click(screen.getByRole('button', { name: /edit/i }));
    expect(onEdit).toHaveBeenCalledWith(mockUser.id);
});
```

## Best Practices
- Lift state up only as far as needed
- Keep components small and focused (< 100 lines)
- Use TypeScript strict mode
- Prefer controlled components for forms
- Use error boundaries for graceful error handling
- Never mutate state directly — always return new objects/arrays
