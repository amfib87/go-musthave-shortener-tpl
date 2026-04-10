// Package exitcheck реализует анализатор статического анализа кода Go,
// который запрещает прямой вызов os.Exit в функции main пакета main.
//
// Назначение анализатора:
//
// Предотвращает аварийное завершение программы без корректной обработки ошибок.
// Прямой вызов os.Exit() в main() обходит выполнение отложенных функций (defer),
// что может привести к:
//   - утечке ресурсов;
//   - отсутствию логирования завершения;
//   - пропуску операций очистки.
//
//
// Рекомендации по исправлению:
// Вместо os.Exit(code) рекомендуется возвращать код ошибки из функции main:
//
//    func main() int {
//        if err := doWork(); err != nil {
//            log.Println("Ошибка:", err)
//            return 1
//        }
//        return 0
//    }
//
// Примеры обнаружения:
//
// 1. Прямой вызов os.Exit:
//    func main() {
//        os.Exit(1)
//    }
//
// 2. Косвенный вызов через log.Fatal (который использует os.Exit):
//    func main() {
//        log.Fatal("Ошибка")
//    }
//
// Анализатор срабатывает только в пакете main и только для функции main().

package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Doc — описание анализатора для отображения в документации.
const Doc = `exitcheck проверяет отсутствие вызовов os.Exit в функции main пакета main`

// Analyzer — экземпляр анализатора, регистрируемый в multichecker.
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  Doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Check if we're in the main package
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	var mainFunc *ast.FuncDecl

	// Находим Main
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if funcDecl.Name.Name == "main" {
					mainFunc = funcDecl
					return false
				}
			}
			return true
		})
	}

	if mainFunc == nil {
		return nil, nil
	}

	// обходим дерево разбора
	ast.Inspect(mainFunc.Body, func(n ast.Node) bool {
		// интересуют только вызовы функций
		if c, ok := n.(*ast.CallExpr); ok {

			if s, ok := c.Fun.(*ast.SelectorExpr); ok {

				ident, ok := s.X.(*ast.Ident)
				if !ok || ident.Name != "os" {
					return true
				}

				// только функции Exit
				if s.Sel.Name == "Exit" {
					pos := pass.Fset.Position(c.Pos())
					// pass.Reportf(call.Pos(), "direct call to os.Exit is prohibited in main.main; return errors instead")
					pass.Reportf(
						c.Pos(),
						"direct call to os.Exit is prohibited in main.main (file: %s:%d); return error code from main instead",
						pos.Filename,
						pos.Line,
					)
				}

			}

			return true
		}
		return true
	})

	return nil, nil
}
