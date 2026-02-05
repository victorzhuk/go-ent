# Cobra

Command-line applications with cobra.

## Quick Reference

| Pattern                 | Purpose                              | Example                                                     |
|-------------------------|--------------------------------------|-------------------------------------------------------------|
| `RunE`                  | Error-returning command handler      | `RunE: func(cmd *cobra.Command, args []string) error`       |
| `PersistentPreRunE`     | Setup before all commands            | `PersistentPreRunE: func(cmd *cobra.Command, args []string) error` |
| `PersistentFlags`       | Flags available to all subcommands   | `rootCmd.PersistentFlags().String("config", "", "...")`     |
| `Flags`                 | Local flags (command-specific)       | `serveCmd.Flags().Int("port", 8080, "...")`                 |
| `MarkFlagRequired`      | Make flag mandatory                  | `cmd.MarkFlagRequired("output")`                            |
| `GenBashCompletion`     | Generate shell completion            | `rootCmd.GenBashCompletionFile("completion.bash")`          |
| `ValidArgsFunction`     | Dynamic completion for args          | `ValidArgsFunction: completeFiles`                          |
| `SilenceUsage`          | Don't show usage on command error    | `cmd.SilenceUsage = true`                                   |

```go
import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application",
    RunE: func(cmd *cobra.Command, args []string) error {
        return fmt.Errorf("specify a subcommand")
    },
}

func init() {
    rootCmd.PersistentFlags().StringP("config", "c", "", "config file")
    rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## PersistentPreRunE

Setup that runs before every command (logger, config, validation).

```go
var (
    cfgFile string
    verbose bool
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Setup logger
        level := slog.LevelInfo
        if verbose {
            level = slog.LevelDebug
        }
        logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
        slog.SetDefault(logger)

        // Load config
        if cfgFile != "" {
            if err := loadConfig(cfgFile); err != nil {
                return fmt.Errorf("load config: %w", err)
            }
        }

        // Validate prerequisites
        if err := checkDependencies(); err != nil {
            return fmt.Errorf("dependency check: %w", err)
        }

        return nil
    },
}

func init() {
    rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
```

## RunE Pattern

Prefer `RunE` over `Run` for error handling. Return errors instead of `log.Fatal`.

```go
var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the server",
    RunE: func(cmd *cobra.Command, args []string) error {
        port, _ := cmd.Flags().GetInt("port")

        cfg, err := loadServerConfig()
        if err != nil {
            return fmt.Errorf("load config: %w", err)
        }

        srv, err := newServer(port, cfg)
        if err != nil {
            return fmt.Errorf("create server: %w", err)
        }

        ctx := cmd.Context()
        if err := srv.Start(ctx); err != nil {
            return fmt.Errorf("start server: %w", err)
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)
    serveCmd.Flags().IntP("port", "p", 8080, "server port")
    serveCmd.SilenceUsage = true // Don't show usage on runtime errors
}
```

## Flag Binding

Bind flags to variables or Viper config.

```go
var (
    port    int
    timeout time.Duration
    config  string
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the server",
    RunE: func(cmd *cobra.Command, args []string) error {
        slog.Info("starting server", "port", port, "timeout", timeout)
        return startServer(port, timeout)
    },
}

func init() {
    // Persistent flags (all subcommands)
    rootCmd.PersistentFlags().StringVar(&config, "config", "", "config file")

    // Local flags (serve command only)
    serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "server port")
    serveCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "request timeout")
    serveCmd.MarkFlagRequired("port")

    // Viper binding (if using viper)
    viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
    viper.BindPFlag("server.timeout", serveCmd.Flags().Lookup("timeout"))
}
```

## Shell Completion

Generate completions for bash/zsh/fish/powershell.

```go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate shell completion",
    Args:  cobra.ExactArgs(1),
    ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
    RunE: func(cmd *cobra.Command, args []string) error {
        switch args[0] {
        case "bash":
            return rootCmd.GenBashCompletion(os.Stdout)
        case "zsh":
            return rootCmd.GenZshCompletion(os.Stdout)
        case "fish":
            return rootCmd.GenFishCompletion(os.Stdout, true)
        case "powershell":
            return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
        }
        return fmt.Errorf("unsupported shell: %s", args[0])
    },
}

// Dynamic completion for file arguments
var exportCmd = &cobra.Command{
    Use:   "export [file]",
    Short: "Export data to file",
    Args:  cobra.ExactArgs(1),
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return []string{"json", "yaml", "csv"}, cobra.ShellCompDirectiveFilterFileExt
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        return exportData(args[0])
    },
}
```

Installation:
```bash
# Bash
myapp completion bash > /etc/bash_completion.d/myapp

# Zsh
myapp completion zsh > "${fpath[1]}/_myapp"

# Fish
myapp completion fish > ~/.config/fish/completions/myapp.fish

# PowerShell
myapp completion powershell | Out-String | Invoke-Expression
```

## Common Mistakes

| Mistake                      | Problem                              | Solution                                    |
|------------------------------|--------------------------------------|---------------------------------------------|
| Using `Run` instead of `RunE`| Can't return errors properly         | Use `RunE` and return errors                |
| `log.Fatal` in command       | Prevents graceful cleanup            | Return error, let `Execute()` handle exit   |
| Missing `PersistentPreRunE`  | Setup duplicated in every command    | Use `PersistentPreRunE` for shared setup    |
| Not setting `SilenceUsage`   | Usage shown on runtime errors        | Set `SilenceUsage = true` after validation  |
| Ignoring `Execute()` error   | Silent failures                      | Check error: `if err := root.Execute(); err != nil` |
| Global state in `init()`     | Hard to test                         | Use dependency injection in `RunE`          |

## See Also

- [Configuration](./configuration.md) - Loading config from files and environment
- [Functional Options](./functional-options.md) - Clean API design for library configuration
- [Interactive](./interactive.md) - Building interactive CLI experiences
