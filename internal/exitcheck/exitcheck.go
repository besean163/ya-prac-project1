// Package exitcheck определяет проверку на использование os.Exit в фукциях main
package exitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit use",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			if file.Name.Name != "main" {
				return false
			}
			switch x := node.(type) {
			case *ast.FuncDecl: // выражение
				if x.Name.Name == "main" {
					ast.Inspect(x, func(node ast.Node) bool {
						switch x := node.(type) {
						case *ast.SelectorExpr:
							if x.Sel.Name == "Exit" {
								pass.Reportf(x.Sel.Pos(), "call os.Exit")
							}
						}
						return true
					})
					return false
				}
			}
			return true
		})
	}
	return nil, nil
}
