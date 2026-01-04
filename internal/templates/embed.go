// Package templates contains embedded project templates for scaffolding.
package templates

import "embed"

// TemplateFS contains all embedded template files.
//
//go:embed *.tmpl
//go:embed .gitignore.tmpl .golangci.yml.tmpl
//go:embed build/*.tmpl
//go:embed cmd/server/*.tmpl
//go:embed deploy/*.tmpl
//go:embed internal/app/*.tmpl internal/config/*.tmpl
//go:embed mcp/*.tmpl
//go:embed mcp/build/*.tmpl
//go:embed mcp/cmd/server/*.tmpl
//go:embed mcp/internal/server/*.tmpl
var TemplateFS embed.FS
