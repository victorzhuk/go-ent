// Package execution provides the execution engine for running agent tasks
// across different runtimes (Claude Code MCP, OpenCode subprocess, CLI).
//
// The execution engine supports multiple execution strategies (single, multi,
// parallel) and includes budget tracking, fallback handling, and JavaScript
// code-mode for dynamic tool composition.
package execution
