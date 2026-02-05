# Interactive CLI

Interactive prompts with huh (replaces archived survey).

## Quick Reference

| Component | Purpose | Code |
|-----------|---------|------|
| **Input** | Text input | `huh.NewInput().Title("Name?").Value(&s)` |
| **Confirm** | Yes/No question | `huh.NewConfirm().Title("Continue?").Value(&b)` |
| **Select** | Single choice | `huh.NewSelect[T]().Options(opts...).Value(&v)` |
| **MultiSelect** | Multiple choices | `huh.NewMultiSelect[T]().Options(opts...).Value(&vs)` |
| **Text** | Multiline input | `huh.NewText().Title("Comment?").Value(&s)` |
| **Form** | Multiple fields | `huh.NewForm(group1, group2).Run()` |
| **Validation** | Field validation | `.Validate(func(s string) error {...})` |
| **Theme** | Custom styling | `.WithTheme(huh.ThemeDracula())` |

```go
import "github.com/charmbracelet/huh"

var name string
huh.NewInput().
    Title("What's your name?").
    Placeholder("John Doe").
    Value(&name).
    RunWithContext(ctx)

var confirmed bool
huh.NewConfirm().
    Title("Continue?").
    Affirmative("Yes").
    Negative("No").
    Value(&confirmed).
    RunWithContext(ctx)
```

## Forms

Multi-step forms with grouped inputs and navigation.

```go
type Config struct {
    Name     string
    Email    string
    Region   string
    Features []string
}

func promptConfig(ctx context.Context) (*Config, error) {
    var cfg Config

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Name").
                Description("Your full name").
                Value(&cfg.Name).
                Validate(func(s string) error {
                    if len(s) < 2 {
                        return fmt.Errorf("name too short")
                    }
                    return nil
                }),

            huh.NewInput().
                Title("Email").
                Placeholder("you@example.com").
                Value(&cfg.Email).
                Validate(validateEmail),
        ),

        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Region").
                Options(
                    huh.NewOption("US East", "us-east"),
                    huh.NewOption("EU West", "eu-west"),
                    huh.NewOption("Asia Pacific", "ap-south"),
                ).
                Value(&cfg.Region),

            huh.NewMultiSelect[string]().
                Title("Features").
                Description("Select features to enable").
                Options(
                    huh.NewOption("Auth", "auth"),
                    huh.NewOption("Cache", "cache"),
                    huh.NewOption("Metrics", "metrics"),
                ).
                Value(&cfg.Features),
        ),
    ).WithTheme(huh.ThemeBase())

    if err := form.RunWithContext(ctx); err != nil {
        return nil, fmt.Errorf("prompt: %w", err)
    }

    return &cfg, nil
}
```

**Navigation**: Tab/Shift+Tab between fields, Enter to advance groups, Ctrl+C to cancel.

## Select and MultiSelect

Single and multiple choice selection with type-safe options.

```go
type Environment string

const (
    EnvDev  Environment = "dev"
    EnvProd Environment = "prod"
)

func selectEnv(ctx context.Context) (Environment, error) {
    var env Environment

    err := huh.NewSelect[Environment]().
        Title("Environment").
        Options(
            huh.NewOption("Development", EnvDev),
            huh.NewOption("Production", EnvProd),
        ).
        Value(&env).
        RunWithContext(ctx)

    return env, err
}

func selectFeatures(ctx context.Context) ([]string, error) {
    var features []string

    err := huh.NewMultiSelect[string]().
        Title("Features").
        Description("Space to toggle, Enter to confirm").
        Options(
            huh.NewOption("Auth", "auth"),
            huh.NewOption("Cache", "cache"),
            huh.NewOption("Metrics", "metrics"),
        ).
        Limit(2). // Max 2 selections
        Value(&features).
        RunWithContext(ctx)

    return features, err
}
```

## Themes

Custom styling for consistent branding.

```go
func customTheme() *huh.Theme {
    theme := huh.ThemeBase()

    // Customize colors
    theme.Focused.Title = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF5733")).
        Bold(true)

    theme.Focused.SelectedOption = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#33FF57"))

    theme.Blurred.Title = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#888888"))

    return theme
}

func styledPrompt(ctx context.Context) error {
    var name string

    return huh.NewInput().
        Title("Name").
        Value(&name).
        WithTheme(customTheme()).
        RunWithContext(ctx)
}
```

**Built-in themes**: `ThemeBase()`, `ThemeDracula()`, `ThemeCatppuccin()`, `ThemeCharm()`.

## Validation

Real-time input validation with custom error messages.

```go
func validateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return fmt.Errorf("invalid email format")
    }
    return nil
}

func validatePort(port string) error {
    p, err := strconv.Atoi(port)
    if err != nil {
        return fmt.Errorf("must be a number")
    }
    if p < 1 || p > 65535 {
        return fmt.Errorf("port must be 1-65535")
    }
    return nil
}

func promptWithValidation(ctx context.Context) error {
    var (
        email string
        port  string
    )

    return huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Email").
                Value(&email).
                Validate(validateEmail),

            huh.NewInput().
                Title("Port").
                Placeholder("8080").
                Value(&port).
                Validate(validatePort),
        ),
    ).RunWithContext(ctx)
}
```

**Validation triggers**: On field blur (Tab/Enter) and form submission.

## Common Mistakes

| Mistake | Problem | Fix |
|---------|---------|-----|
| `form.Run()` without context | Blocks indefinitely, no cancellation | Use `RunWithContext(ctx)` |
| Generic error messages | "Invalid input" unhelpful | Specific: "Port must be 1-65535" |
| No keyboard shortcuts shown | Users unaware of Space/Tab/Enter | Add `.Description("Space to toggle")` |
| Missing validation | Silent failures, bad data | Add `.Validate(func)` to critical fields |
| Too many form groups | Overwhelming UX | Max 3-4 groups, combine related fields |
| Blocking main goroutine | Hangs on signal interrupt | Wrap in goroutine, use context cancellation |

## See Also

- [Cobra](./cobra.md) - CLI framework integration
- [Configuration](./configuration.md) - Env-based config
- [Input Validation](../11-security/input-validation.md) - Validation patterns
