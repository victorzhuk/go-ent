# Flutter / Maestro Reference

## Install

```bash
# Maestro (Java 17+ required)
curl -Ls "https://get.maestro.dev" | bash
export PATH="$PATH:$HOME/.maestro/bin"
maestro --version

# iOS support (macOS only)
brew tap facebook/fb
brew install idb-companion

# patrol CLI
dart pub global activate patrol_cli
patrol --version

# Verify Flutter environment
flutter doctor -v
flutter devices
```

## Flutter App Setup for Testing

```dart
// Add semantics identifiers for reliable Maestro targeting (Flutter 3.19+)
Semantics(
  identifier: 'login-button',
  child: ElevatedButton(
    onPressed: _onLogin,
    child: Text('Login'),
  ),
)

// Or via ValueKey for patrol
ElevatedButton(
  key: const ValueKey('login-button'),
  onPressed: _onLogin,
  child: Text('Login'),
)
```

## Maestro Flow Examples

### Basic login flow

```yaml
# .maestro/login_flow.yaml
appId: com.example.myapp
---
- launchApp:
    clearState: true
    stopApp: true

- assertVisible: "Login"

- tapOn:
    id: "email-field"
- inputText: "qa@test.com"

- tapOn:
    id: "password-field"
- inputText: "SecurePass123!"

- tapOn:
    id: "login-button"

- assertVisible: "Dashboard"
- takeScreenshot: login-success
```

### Checkout flow

```yaml
# .maestro/checkout_flow.yaml
appId: com.example.shop
---
- launchApp:
    clearState: true

- assertVisible: "Featured Products"
- scrollUntilVisible:
    element: "Premium Widget"
    direction: DOWN
- tapOn: "Premium Widget"
- assertVisible: "Add to Cart"
- tapOn: "Add to Cart"

- tapOn:
    id: "cart-icon"
- assertVisible: "Your Cart"
- assertVisible: "Premium Widget"

- tapOn: "Proceed to Checkout"
- inputText:
    id: "card-number"
    text: "4111111111111111"
- inputText:
    id: "cvv"
    text: "123"
- tapOn: "Place Order"
- assertVisible: "Order Confirmed"
- takeScreenshot: checkout-success
```

### Navigation test

```yaml
# .maestro/navigation_test.yaml
appId: com.example.myapp
---
- launchApp
- assertVisible: "Home"

- tapOn:
    id: "profile-tab"
- assertVisible: "My Profile"

- tapOn:
    id: "settings-tab"
- assertVisible: "Settings"

- pressKey: Back
- assertVisible: "My Profile"

- openLink: "myapp://product/123"
- assertVisible: "Product Details"
```

### Permissions flow

```yaml
# .maestro/permissions_flow.yaml
appId: com.example.myapp
---
- launchApp
- tapOn: "Enable Notifications"
- allowPermission
- assertVisible: "Notifications enabled"

- tapOn: "Take Photo"
- allowPermission
- assertVisible: "Camera ready"
```

### JavaScript integration

```yaml
# .maestro/api_driven_flow.yaml
appId: com.example.myapp
---
- runScript:
    file: setup.js
    env:
      API_URL: ${API_URL}

- launchApp
- assertVisible: "${welcomeMessage}"
```

```javascript
// setup.js
const response = http.get(`${process.env.API_URL}/test-users/create`)
output.userId = response.body.id
output.welcomeMessage = `Welcome, ${response.body.name}`
```

## Maestro CLI Commands

```bash
maestro test .maestro/login_flow.yaml
maestro test .maestro/
maestro test --include-tags smoke .maestro/
maestro test --watch .maestro/login_flow.yaml
maestro record
maestro studio
maestro test --format=junit --output results/ .maestro/
maestro test --shards 3 .maestro/
maestro test --device emulator-5554 .maestro/login_flow.yaml
maestro hierarchy
maestro scroll-up
maestro scroll-down
```

## Patrol — Flutter-Native Integration Tests

```dart
// integration_test/login_test.dart
import 'package:patrol/patrol.dart';

void main() {
  patrolTest(
    'user can log in',
    config: PatrolTesterConfig(
      visibleTimeout: const Duration(seconds: 10),
    ),
    ($) async {
      await $.pumpWidgetAndSettle(MyApp());

      await $(#emailField).enterText('qa@test.com');
      await $(#passwordField).enterText('secret');
      await $('Login').tap();

      await $.waitUntilVisible(find.text('Dashboard'));
      expect(find.text('Welcome'), findsOneWidget);
    },
  );
}

void main() {
  patrolTest('handles notification permission', ($) async {
    await $.pumpWidgetAndSettle(MyApp());
    await $('Enable Notifications').tap();

    await $.native.grantPermissionWhenInUse();
    await $.waitUntilVisible(find.text('Notifications enabled'));
  });
}
```

```bash
patrol test
patrol test integration_test/login_test.dart
patrol test --device "iPhone 15"
patrol test --repeat 3
patrol generate
```

## flutter test — Widget & Unit Tests

```bash
flutter test
flutter test test/widgets/login_form_test.dart
flutter test --coverage
genhtml coverage/lcov.info -o coverage/html
flutter test --update-goldens
flutter test --name "login form"
flutter test -v
flutter test --watch
```

### Widget test example

```dart
// test/widgets/login_form_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:my_app/widgets/login_form.dart';

void main() {
  testWidgets('shows error on empty submit', (WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(home: LoginForm()));

    await tester.tap(find.text('Login'));
    await tester.pumpAndSettle();

    expect(find.text('Email is required'), findsOneWidget);
  });

  testWidgets('calls onLogin with credentials', (WidgetTester tester) async {
    String? capturedEmail;
    await tester.pumpWidget(MaterialApp(
      home: LoginForm(onLogin: (email, pass) => capturedEmail = email),
    ));

    await tester.enterText(find.byKey(Key('email-field')), 'test@test.com');
    await tester.enterText(find.byKey(Key('password-field')), 'secret');
    await tester.tap(find.text('Login'));
    await tester.pumpAndSettle();

    expect(capturedEmail, 'test@test.com');
  });

  testWidgets('login screen matches golden', (WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(home: LoginScreen()));
    await expectLater(
      find.byType(LoginScreen),
      matchesGoldenFile('goldens/login_screen.png'),
    );
  });
}
```

## CI Integration

```yaml
# .github/workflows/flutter-test.yml
- name: Run unit tests
  run: flutter test --coverage

- name: Run Maestro E2E (Android)
  uses: mobile-dev-inc/action-maestro-cloud@v1
  with:
    api-key: ${{ secrets.MAESTRO_CLOUD_API_KEY }}
    app-file: build/app.apk
    workspace: .maestro/

- name: Run Maestro locally
  run: |
    maestro test \
      --format=junit \
      --output test-results/ \
      .maestro/
```

## Agent Patterns

### Discovery

```bash
flutter devices
maestro hierarchy
maestro test .maestro/inspect.yaml
```

### Execution order

```bash
# 1. Fast: no device
flutter test --coverage

# 2. Medium: device required
patrol test

# 3. Slow: E2E flows
maestro test --format=junit --output results/ .maestro/

# 4. Aggregate JUnit results
cat results/*.xml | xq '.testsuites.testsuite[]'
```
