package spqex

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
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
	if (s[0] != '"' && s[0] != '`') || (s[len(s)-1] != '"' && s[len(s)-1] != '`') {
		return s
	}
	return s[1 : len(s)-1]
}

func trimNewlines(data []byte) []byte {
	for len(data) > 0 && (data[0] == '\n' || data[0] == '\r') {
		data = data[1:]
	}

	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == '\r') {
		data = data[:len(data)-1]
	}

	return data
}

func hasNewline(s string) bool {
	return strings.Contains(s, "\n")
}

type CommandResult struct {
	Output   string
	ExitCode int
}

func RunCommand(command, sql string) (*CommandResult, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdin = strings.NewReader(sql)

	output, err := cmd.CombinedOutput()

	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return nil, fmt.Errorf("failed to execute command %q: %v", command, err)
	}

	output = trimNewlines(output)

	if exitError != nil {
		return &CommandResult{
			Output:   string(output),
			ExitCode: exitError.ExitCode(),
		}, nil
	}
	return &CommandResult{
		Output:   string(output),
		ExitCode: 0,
	}, nil
}

type ErrorMessage struct {
	Message string
	PosText string
}

func (e *ErrorMessage) String() string {
	return fmt.Sprintf("%s:\n%s", e.PosText, e.Message)
}

type ProcessResult struct {
	Output        []byte
	ErrorMessages []*ErrorMessage
	IsChanged     bool
}

func Process(path string, externalCmd string, replace bool) (*ProcessResult, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
	}

	basicLitExprs := findSpannerSQLExpr(node)

	errMessages := make([]*ErrorMessage, 0, len(basicLitExprs))
	if len(basicLitExprs) == 0 {
		return &ProcessResult{
			Output:        nil,
			ErrorMessages: errMessages,
			IsChanged:     false,
		}, nil
	}

	for _, basicLitExpr := range basicLitExprs {
		r, err := RunCommand(externalCmd, trimQuotes(basicLitExpr.Value))
		if err != nil {
			return nil, fmt.Errorf("failed to run command: %v", err)
		}
		if r.ExitCode != 0 {
			errMessages = append(errMessages, &ErrorMessage{
				Message: r.Output,
				PosText: fset.Position(basicLitExpr.Pos()).String(),
			})
			continue
		}
		if replace {
			if hasNewline(r.Output) {
				basicLitExpr.Value = fmt.Sprintf("`\n%s\n`", r.Output)
			} else {
				basicLitExpr.Value = fmt.Sprintf("\"%s\"", r.Output)
			}
		}
	}

	if !replace || len(errMessages) == len(basicLitExprs) {
		return &ProcessResult{
			Output:        nil,
			ErrorMessages: errMessages,
			IsChanged:     false,
		}, nil
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return nil, fmt.Errorf("%s: failed to print AST: %v", path, err)
	}

	result, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to format source: %v", path, err)
	}

	return &ProcessResult{
		Output:        result,
		ErrorMessages: errMessages,
		IsChanged:     true,
	}, nil
}

func FindGoFiles(directory string) ([]string, error) {
	files := make([]string, 0)

	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if info.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %v", directory, err)
	}

	return files, nil
}
