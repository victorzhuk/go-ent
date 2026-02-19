## Modern PHP (8.3+)
- Use strict types: `declare(strict_types=1);` in every file
- Use typed properties, return types, and union types
- Use enums: `enum Status: string { case Active = 'active'; }`
- Use readonly classes and properties
- Use named arguments for clarity
- Use match expressions instead of switch
- Use fibers for async-like patterns
- Use `#[Override]` attribute (PHP 8.3)

## PSR Standards
- **PSR-4**: Autoloading with Composer
- **PSR-7**: HTTP message interfaces
- **PSR-11**: Container interface
- **PSR-12**: Extended coding style (use PHP-CS-Fixer or PHP_CodeSniffer)
- **PSR-15**: HTTP server request handlers and middleware
- **PSR-18**: HTTP client interface

## Project Structure
```
src/
├── Domain/          # Entities, value objects, interfaces
├── Application/     # Use cases, DTOs, services
├── Infrastructure/  # Database, external APIs, framework
├── Http/            # Controllers, middleware, requests
└── Console/         # CLI commands
tests/
├── Unit/
├── Integration/
└── Feature/
composer.json
phpstan.neon
phpunit.xml
```

## Error Handling
```php
// Custom exception hierarchy
class DomainException extends \RuntimeException {}
class NotFoundException extends DomainException {}
class ValidationException extends DomainException {
    public function __construct(
        private readonly array $errors,
        string $message = 'Validation failed',
    ) {
        parent::__construct($message);
    }
}
```

## Tooling
- **Composer**: Dependency management (always use `composer.lock`)
- **PHPStan**: Static analysis at level 8+
- **PHP-CS-Fixer**: Code style enforcement
- **PHPUnit**: Testing framework
- **Pest**: Modern testing (alternative to PHPUnit)
- **Rector**: Automated refactoring and upgrades

## Best Practices
- Use dependency injection — never `new` dependencies in constructors
- Use interfaces for external dependencies
- Validate input at the boundary (controllers/commands)
- Use DTOs for data transfer between layers
- Avoid static methods for business logic
- Use immutable value objects for domain concepts
