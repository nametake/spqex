package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"strings"
)

func findSpannerSQLExpr(node *ast.File) []*ast.BasicLit {
	basicLitExprs := make([]*ast.BasicLit, 0)
	ast.Inspect(node, func(n ast.Node) bool {
		compositeLitExpr, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		selectorExpr, ok := compositeLitExpr.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkgIdent, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		if pkgIdent.Name != "spanner" && selectorExpr.Sel.Name != "Statement" {
			return true
		}

		for _, elt := range compositeLitExpr.Elts {
			elt, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := elt.Key.(*ast.Ident)
			if !ok {
				continue
			}
			if key.Name != "SQL" {
				continue
			}
			value, ok := elt.Value.(*ast.BasicLit)
			if !ok {
				continue
			}

			basicLitExprs = append(basicLitExprs, value)
		}

		return true
	})

	return basicLitExprs
}

func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	if s[0] != '"' || s[len(s)-1] != '"' {
		return s
	}
	if s[0] != '`' || s[len(s)-1] != '`' {
		return s
	}
	return s[1 : len(s)-1]
}

func process(path string, externalCmd string) ([]byte, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, path, source, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
	}

	basicLitExprs := findSpannerSQLExpr(node)

	for _, basicLitExpr := range basicLitExprs {
		cmd := exec.Command("sh", "-c", externalCmd)

		cmd.Stdin = strings.NewReader(trimQuotes(basicLitExpr.Value))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to execute command %q: %v", externalCmd, err)
		}
		basicLitExpr.Value = fmt.Sprintf("`%s`", output)
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return nil, fmt.Errorf("failed to print AST: %v", err)
	}

	result, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format source: %v", err)
	}

	return result, nil
}
