package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/cli"
	internalserver "github.com/victorzhuk/go-ent/internal/mcp/server"
	"github.com/victorzhuk/go-ent/internal/version"
)

func main() {
	// Detect CLI mode vs MCP mode
	// CLI mode: has arguments (except help flags) or is a TTY
	// MCP mode: stdin is a pipe (not a TTY)
	if len(os.Args) > 1 {
		// Handle version flag for backward compatibility
		switch os.Args[1] {
		case "version", "--version", "-v":
			v := version.Get()
			fmt.Printf("go-ent %s\n", version.String())
			fmt.Printf("  go: %s\n", v.GoVersion)
			os.Exit(0)
		}

		// Run in CLI mode
		if err := cli.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run in MCP server mode (no arguments, expects stdio communication)
	if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
	logger := setupLogger(getenv("LOG_LEVEL"), getenv("LOG_FORMAT"), stdout)
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	s := internalserver.New()
	transport := &mcp.StdioTransport{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Run(ctx, transport)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("mcp server: %w", err)
		}
		return nil
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.Info("shutting down gracefully", "timeout", "30s")

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	case <-shutdownCtx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

func setupLogger(level, format string, w io.Writer) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}
	return slog.New(handler)
}
