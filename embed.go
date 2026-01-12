// Package goent provides embedded plugin resources for the go-ent binary.
// This file must be at the module root to access the plugins/ directory.
package goent

import "embed"

// PluginFS embeds all go-ent plugin resources into the binary.
// This allows distribution as a single executable that can generate
// tool-specific configurations for Claude Code, OpenCode, and Cursor.
//
//go:embed plugins/go-ent/agents/*.md
//go:embed plugins/go-ent/commands/*.md
//go:embed plugins/go-ent/skills/*/*/SKILL.md
//go:embed plugins/go-ent/hooks/hooks.json
//go:embed plugins/go-ent/scripts/run-mcp.sh
var PluginFS embed.FS
