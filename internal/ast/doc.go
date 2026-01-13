// Package ast provides Go abstract syntax tree (AST) parsing utilities.
//
// # Overview
//
// The ast package wraps the Go standard library's go/ast, go/parser,
// and go/token packages to provide a simplified interface for parsing
// Go source code into abstract syntax trees.
//
// # Parser
//
// The Parser type provides methods for parsing Go source code:
//
//   - ParseFile: Parse a Go source file from the filesystem
//   - ParseString: Parse Go source code from a string
//   - FileSet: Access the underlying token.FileSet
//
// # Usage
//
// Basic file parsing:
//
//	p := ast.NewParser()
//	f, err := p.ParseFile("path/to/file.go")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parsing from string:
//
//	p := ast.NewParser()
//	src := `package main
//
//	func main() {
//		    println("hello")
//		}`
//		f, err := p.ParseString(src)
//		if err != nil {
//		    log.Fatal(err)
//		}
//
// # Error Handling
//
// Parse errors are wrapped with context:
//
//   - File not found: errors with "open" context
//   - Syntax errors: errors with "parse" context
//
// Use errors.Is or errors.As to check for specific error conditions.
//
// # Token FileSet
//
// The parser maintains a token.FileSet for position information:
//
//	fset := p.FileSet()
//	// Use fset with ast.Node.Pos() to get file positions
//
// See the Go standard library documentation for go/ast, go/parser,
// and go/token for more details on working with Go ASTs.
package ast
