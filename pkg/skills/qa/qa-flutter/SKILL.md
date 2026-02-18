---
name: qa-flutter
description: Autonomous Flutter/mobile testing with Maestro (YAML flows), patrol (Flutter-native), flutter test, and flutter drive. Use for Flutter UI tests, mobile E2E, widget testing, and cross-platform iOS/Android automation.
triggers:
  - flutter test
  - mobile test
  - maestro
  - patrol
  - android test
  - ios test
  - widget test
---

## Role

Mobile QA specialist using Maestro as the primary E2E tool — YAML flows that run on both Android and iOS without rewriting, handle OS-level dialogs, and read like plain English. Fall back to patrol for tests that require Flutter-native introspection or OS permission interactions. Use `flutter test` for fast unit and widget test cycles.

## Instructions

### When to use this skill vs others

- Use qa-flutter for any mobile-specific flow: widget tests, Maestro E2E flows, patrol integration tests, and Flutter performance profiling.
- Delegate to qa-performance when the goal is API load testing that the mobile app triggers.
- Delegate to qa-visual when comparing Flutter golden files against a reference platform (web/desktop).

### Response format

Provide a Maestro YAML flow file as the primary artifact for E2E scenarios. Provide a Dart test file for widget or unit test scenarios. Include the CLI command to run. For CI, include the GitHub Actions step.

### Edge cases

New project: run `maestro record` to generate a YAML flow interactively from taps on a connected device.

No real device: use an emulator (`flutter emulators --launch <id>`) and verify with `flutter devices` before running Maestro.

iOS testing: requires `idb-companion` on macOS. Install via `brew install idb-companion`.

Performance assertions: always use release or profile builds — debug builds include overhead that skews results.

CI integration: use Maestro Cloud action for cloud-based device farms; run `flutter test --coverage` locally for unit/widget tests.

## Examples

### Example 1

**Input**: Test the login flow in a Flutter app with email and password fields

**Output**:

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

```bash
maestro test .maestro/login_flow.yaml
```

### Example 2

**Input**: Write a widget test for a login form that shows an error on empty submit

**Output**:

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
}
```

```bash
flutter test test/widgets/login_form_test.dart
```

## Quick Reference

See [references/maestro-reference.md](references/maestro-reference.md) for complete Maestro flow syntax, CLI commands, patrol integration test examples, flutter test commands, CI configuration, and agent patterns.

**Install:**
```bash
curl -Ls "https://get.maestro.dev" | bash
dart pub global activate patrol_cli
flutter doctor -v && flutter devices
```

**Execution order (fastest to slowest):**
1. `flutter test --coverage` — unit + widget (no device)
2. `patrol test` — Flutter integration (device required)
3. `maestro test .maestro/` — E2E flows (device required)

**Key anti-patterns:**
- Use `Semantics(identifier:)` not Widget keys for Maestro targeting
- Never use `sleep()` in YAML — Maestro auto-waits
- Always `clearState: true` on `launchApp` to avoid stale state
- Use release/profile builds for any performance assertions
