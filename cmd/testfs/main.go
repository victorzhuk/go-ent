package main

import (
	"fmt"
	"io/fs"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: testfs <path-to-plugin-dir>")
		os.Exit(1)
	}

	dir := os.Args[1]
	pluginFS := os.DirFS(dir)

	// Test reading shared file
	path := "plugins/go-ent/agents/prompts/shared/_tooling.md"
	content, err := fs.ReadFile(pluginFS, path)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", path, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Read %d bytes from %s\n", len(content), path)
	fmt.Printf("Content preview: %s\n", string(content[:min(100, len(content))]))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
