## App Router Structure
```
app/
├── layout.tsx          # Root layout
├── page.tsx            # Home page
├── loading.tsx         # Loading UI
├── error.tsx           # Error boundary
├── not-found.tsx       # 404 page
├── api/                # Route Handlers
│   └── users/route.ts
├── users/
│   ├── page.tsx        # /users
│   ├── [id]/
│   │   ├── page.tsx    # /users/:id
│   │   └── edit/page.tsx
│   └── layout.tsx
└── (auth)/             # Route groups
    ├── login/page.tsx
    └── register/page.tsx
```

## Server Components (Default)
```tsx
// This is a Server Component — runs on the server
export default async function UsersPage() {
    const users = await getUsers(); // Direct DB/API call, no useEffect
    return <UserList users={users} />;
}
```

## Client Components
```tsx
'use client'; // Explicit opt-in

export function Counter() {
    const [count, setCount] = useState(0);
    return <button onClick={() => setCount(c => c + 1)}>{count}</button>;
}
```

## Server Actions
```tsx
'use server';

export async function createUser(formData: FormData) {
    const data = schema.parse(Object.fromEntries(formData));
    await db.users.create({ data });
    revalidatePath('/users');
    redirect('/users');
}
```

## Data Fetching & Caching
- Server Components fetch data directly (no API routes needed)
- Use `unstable_cache` for caching expensive computations
- Use `revalidatePath`/`revalidateTag` for on-demand revalidation
- Use `generateStaticParams` for static generation of dynamic routes

## Middleware
```tsx
// middleware.ts
export function middleware(request: NextRequest) {
    // Auth checks, redirects, headers
}
export const config = { matcher: ['/dashboard/:path*'] };
```

## Best Practices
- Default to Server Components — use Client Components sparingly
- Colocate data fetching with the component that uses it
- Use loading.tsx and Suspense for streaming
- Use Route Groups for layout organization
- Use Parallel Routes for complex dashboards
- Optimize images with `next/image`
- Use `next/font` for font optimization
