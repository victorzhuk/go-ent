// Package pkg provides embedded resources for go-ent.
// This is the single source of truth for all embedded content including
// agents, skills, prompts, templates, schemas, commands, hooks, and scripts.
package pkg

import "embed"

// FS embeds all go-ent resources into the binary.
// This allows distribution as a single executable that can generate
// tool-specific configurations for Claude Code, OpenCode, and other tools.
//
//go:embed all:agents all:skills all:templates all:schemas all:commands all:hooks all:scripts
var FS embed.FS
