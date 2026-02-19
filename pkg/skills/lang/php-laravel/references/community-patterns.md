## Architecture
- Follow Laravel conventions — don't fight the framework
- Use Service classes for complex business logic
- Use Form Requests for validation
- Use API Resources for response transformation
- Use Events + Listeners for decoupling
- Use Jobs + Queues for background processing

## Eloquent Best Practices
```php
// Use query scopes
class User extends Model {
    public function scopeActive(Builder $query): Builder {
        return $query->where('status', Status::Active);
    }
}

// Eager loading to prevent N+1
$users = User::with(['posts', 'roles'])->active()->paginate(20);

// Use chunking for large datasets
User::where('status', 'inactive')->chunkById(1000, function ($users) {
    foreach ($users as $user) { $user->delete(); }
});
```

## API Development
```php
// API Resource for consistent responses
class UserResource extends JsonResource {
    public function toArray(Request $request): array {
        return [
            'id' => $this->id,
            'name' => $this->name,
            'email' => $this->email,
            'created_at' => $this->created_at->toISOString(),
        ];
    }
}

// Controller
class UserController extends Controller {
    public function index(): AnonymousResourceCollection {
        return UserResource::collection(
            User::query()->active()->paginate()
        );
    }
}
```

## Testing
```php
class CreateUserTest extends TestCase {
    use RefreshDatabase;

    public function test_creates_user_successfully(): void {
        $response = $this->postJson('/api/users', [
            'name' => 'Alice',
            'email' => 'alice@example.com',
        ]);

        $response->assertCreated()
                 ->assertJsonPath('data.name', 'Alice');
        $this->assertDatabaseHas('users', ['email' => 'alice@example.com']);
    }
}
```

## Production
- Use `php artisan config:cache`, `route:cache`, `view:cache`
- Use Redis for sessions, cache, and queues
- Use Laravel Horizon for queue monitoring
- Set up proper logging channels (daily rotation, stderr for containers)
- Use database transactions for multi-step operations
