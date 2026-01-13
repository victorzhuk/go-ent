package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type Parser struct {
	fset *token.FileSet
}

func NewParser() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

func (p *Parser) ParseFile(path string) (*ast.File, error) {
	f, err := parser.ParseFile(p.fset, path, nil, parser.AllErrors)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("open: %w", err)
		}
		return nil, fmt.Errorf("parse: %w", err)
	}
	return f, nil
}

func (p *Parser) ParseString(src string) (*ast.File, error) {
	f, err := parser.ParseFile(p.fset, "", src, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return f, nil
}

func (p *Parser) FileSet() *token.FileSet {
	return p.fset
}
