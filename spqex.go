package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func process(path string) error {
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", path, err)
	}

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, path, source, 0)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %v", path, err)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// TODO format and lint
		return true
	})

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return fmt.Errorf("failed to print AST: %v", err)
	}

	return nil
}
