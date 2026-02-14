---
name: nodejs-nestjs
description: NestJS development with modules, dependency injection, guards, interceptors, and clean architecture
---

# NestJS Development

## Module Organization
```
src/
├── app.module.ts
├── common/              # Shared guards, pipes, interceptors, filters
├── config/              # Configuration module
├── users/
│   ├── users.module.ts
│   ├── users.controller.ts
│   ├── users.service.ts
│   ├── dto/
│   ├── entities/
│   └── users.repository.ts
└── auth/
    ├── auth.module.ts
    ├── auth.guard.ts
    └── strategies/
```

## Controllers & DTOs
```typescript
@Controller('users')
@UseGuards(JwtAuthGuard)
export class UsersController {
    constructor(private readonly usersService: UsersService) {}

    @Post()
    @HttpCode(HttpStatus.CREATED)
    async create(@Body() dto: CreateUserDto): Promise<UserResponseDto> {
        return this.usersService.create(dto);
    }

    @Get(':id')
    async findOne(@Param('id', ParseUUIDPipe) id: string): Promise<UserResponseDto> {
        return this.usersService.findOne(id);
    }
}
```

## Dependency Injection
- Use constructor injection (NestJS standard)
- Use custom providers for complex setup
- Use `@Injectable({ scope: Scope.REQUEST })` sparingly — prefer default singleton
- Use `forRootAsync` for async module configuration

## Exception Handling
```typescript
@Catch()
export class AllExceptionsFilter implements ExceptionFilter {
    catch(exception: unknown, host: ArgumentsHost) {
        const ctx = host.switchToHttp();
        const response = ctx.getResponse<Response>();
        // Map to standard error response...
    }
}
```

## Testing
- Unit test services with mocked dependencies
- Use `Test.createTestingModule` for integration tests
- Use `@nestjs/testing` utilities
- Test guards and interceptors independently

## Best Practices
- One module per domain concept
- Use pipes for validation (`ValidationPipe` globally)
- Use interceptors for response transformation
- Use guards for authorization
- Use events (`@nestjs/event-emitter`) for decoupling
