## TypeScript Configuration
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "outDir": "dist",
    "declaration": true,
    "sourceMap": true,
    "noUncheckedIndexedAccess": true
  }
}
```

## Express.js Patterns
```typescript
const app = express();
app.use(express.json({ limit: '10kb' }));
app.use(helmet());
app.use(cors(corsOptions));
app.use(requestId());
app.use(requestLogger());

// Error handling middleware (always last)
app.use((err: AppError, req: Request, res: Response, next: NextFunction) => {
    logger.error({ err, requestId: req.id }, 'Request failed');
    res.status(err.statusCode ?? 500).json({
        error: { code: err.code, message: err.message }
    });
});
```

## Fastify (Preferred for Performance)
```typescript
const server = Fastify({
    logger: { level: 'info' },
    ajv: { customOptions: { allErrors: true } },
});

server.register(fastifyCors);
server.register(fastifyHelmet);
server.register(userRoutes, { prefix: '/api/v1/users' });
```

## Async Error Handling
- Always wrap async route handlers or use express-async-errors
- Use `Promise.allSettled` when partial failures are acceptable
- Use `AbortController` for cancellable operations
- Never swallow errors silently — log and rethrow or handle

## Database (Prisma / Drizzle)
- Use Prisma for rapid development with strong typing
- Use Drizzle for performance-critical or edge-deployed apps
- Always use transactions for multi-step mutations
- Use connection pooling (PgBouncer or built-in)

## Production Checklist
- Use `cluster` module or PM2 for multi-process
- Set `NODE_ENV=production`
- Configure health check endpoints
- Use structured logging (pino/winston)
- Implement graceful shutdown
- Set proper timeouts on HTTP server and database connections
- Rate limit API endpoints
