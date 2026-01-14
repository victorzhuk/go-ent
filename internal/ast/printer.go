package ast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

type Printer struct {
	fset *token.FileSet
	cfg  *printer.Config
}

func NewPrinter(fset *token.FileSet) *Printer {
	return &Printer{
		fset: fset,
		cfg: &printer.Config{
			Mode:     printer.UseSpaces | printer.TabIndent,
			Tabwidth: 8,
		},
	}
}

func (p *Printer) PrintFile(f *ast.File) (string, error) {
	if f == nil {
		return "", fmt.Errorf("nil file")
	}

	var buf bytes.Buffer
	if err := p.cfg.Fprint(&buf, p.fset, f); err != nil {
		return "", fmt.Errorf("print file: %w", err)
	}
	return buf.String(), nil
}

func (p *Printer) PrintNode(node ast.Node) (string, error) {
	if node == nil {
		return "", fmt.Errorf("nil node")
	}

	var buf bytes.Buffer
	if err := p.cfg.Fprint(&buf, p.fset, node); err != nil {
		return "", fmt.Errorf("print node: %w", err)
	}
	return buf.String(), nil
}

func (p *Printer) WriteFile(f *ast.File, path string) error {
	if f == nil {
		return fmt.Errorf("nil file")
	}
	if path == "" {
		return fmt.Errorf("empty path")
	}

	var buf bytes.Buffer
	if err := p.cfg.Fprint(&buf, p.fset, f); err != nil {
		return fmt.Errorf("print file: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
