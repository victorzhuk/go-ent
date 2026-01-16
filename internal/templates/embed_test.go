package templates_test

//nolint:gosec // test file with necessary file operations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/templates"
)

func TestEmbeddedFiles(t *testing.T) {
	// Test that we can read root directory
	entries, err := templates.TemplateFS.ReadDir(".")
	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	// Test that dotfiles are embedded
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}

	assert.Contains(t, names, ".gitignore.tmpl", "Should embed .gitignore.tmpl")
	assert.Contains(t, names, ".golangci.yml.tmpl", "Should embed .golangci.yml.tmpl")
	assert.Contains(t, names, "CLAUDE.md.tmpl")
	assert.Contains(t, names, "Makefile.tmpl")
	assert.Contains(t, names, "go.mod.tmpl")

	// Test subdirectories
	assert.Contains(t, names, "build")
	assert.Contains(t, names, "cmd")
	assert.Contains(t, names, "deploy")
	assert.Contains(t, names, "internal")
	assert.Contains(t, names, "mcp")
}

func TestDotfilesReadable(t *testing.T) {
	// Test that we can actually read the dotfiles
	gitignore, err := templates.TemplateFS.ReadFile(".gitignore.tmpl")
	assert.NoError(t, err)
	assert.NotEmpty(t, gitignore)

	golangci, err := templates.TemplateFS.ReadFile(".golangci.yml.tmpl")
	assert.NoError(t, err)
	assert.NotEmpty(t, golangci)
}

func TestAllTemplatesCount(t *testing.T) {
	// Count all .tmpl files recursively
	count := 0
	var countFiles func(string)
	countFiles = func(dir string) {
		entries, err := templates.TemplateFS.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			path := dir + "/" + e.Name()
			if dir == "." {
				path = e.Name()
			}
			if e.IsDir() {
				countFiles(path)
			} else {
				count++
			}
		}
	}
	countFiles(".")

	// Should have all 15 templates
	assert.Equal(t, 15, count, "Should embed all 15 template files")
}
