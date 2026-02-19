## Project Structure
```
lib/
├── main.dart
├── app/
│   ├── app.dart           # MaterialApp configuration
│   └── router.dart        # GoRouter configuration
├── features/
│   └── users/
│       ├── data/          # Repositories, data sources
│       ├── domain/        # Models, use cases
│       └── presentation/  # Widgets, providers/blocs
├── core/
│   ├── theme/
│   ├── widgets/           # Shared widgets
│   ├── utils/
│   └── constants.dart
└── l10n/                  # Localization
```

## Widget Best Practices
```dart
// Prefer const constructors
class UserCard extends StatelessWidget {
  const UserCard({super.key, required this.user, this.onTap});
  final User user;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        title: Text(user.name),
        subtitle: Text(user.email),
        onTap: onTap,
      ),
    );
  }
}
```

## State Management (Riverpod - Recommended)
```dart
// Provider definition
final userProvider = FutureProvider.autoDispose.family<User, String>((ref, id) {
  return ref.watch(userRepositoryProvider).getById(id);
});

// Usage in widget
class UserScreen extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userAsync = ref.watch(userProvider(userId));
    return userAsync.when(
      data: (user) => UserCard(user: user),
      loading: () => const CircularProgressIndicator(),
      error: (e, _) => ErrorWidget(e.toString()),
    );
  }
}
```

## Navigation (GoRouter)
```dart
final router = GoRouter(
  routes: [
    GoRoute(path: '/', builder: (_, __) => const HomeScreen()),
    GoRoute(
      path: '/users/:id',
      builder: (_, state) => UserScreen(id: state.pathParameters['id']!),
    ),
  ],
  redirect: (context, state) {
    if (!isAuthenticated) return '/login';
    return null;
  },
);
```

## Performance
- Use `const` constructors everywhere possible
- Use `ListView.builder` for long lists (not `ListView` with children)
- Avoid `setState` in deeply nested widgets — lift state up or use state management
- Use `RepaintBoundary` for expensive paint operations
- Profile with Flutter DevTools — focus on build, layout, paint phases
- Use `compute()` for heavy synchronous work

## Testing
```dart
testWidgets('displays user name', (tester) async {
  await tester.pumpWidget(
    MaterialApp(home: UserCard(user: testUser)),
  );
  expect(find.text('Alice'), findsOneWidget);
});

test('UserRepository returns user', () async {
  final repo = UserRepository(mockClient);
  final user = await repo.getById('1');
  expect(user.name, equals('Alice'));
});
```

## Platform-Specific
- Use `Platform.isAndroid`/`Platform.isIOS` for platform checks
- Use `MethodChannel` for native platform communication
- Use `kIsWeb` for web-specific logic
- Test on both iOS and Android — behavior differs
- Use adaptive widgets (`Switch.adaptive`, etc.) for platform-native feel
