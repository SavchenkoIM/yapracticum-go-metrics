package main

import (
	"fmt"
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
	"strings"
)

// Analyzer: Checks main function of main package for direct os.Exit() call
var ExitCallInMainAnalyzer = &analysis.Analyzer{
	Name: "exitinmain",
	Doc:  "check for call of os.Exit in main func of main package",
	Run:  exitCallInMainCheckRun,
}

// Main function of ExitCallInMainAnalyzer
func exitCallInMainCheckRun(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		isExitCalledInFile(file, pass.Fset)
	}

	return nil, nil
}

// Auxilary function for ExitCallInMainAnalyzer: searches call of function callFunc in function fun
func isFuncCalledInFunc(fun *ast.FuncDecl, callFunc string, fset *token.FileSet) bool {
	// "found" is not only flag for Inspect termination.
	// This variable is ret value for this function (in case it is used by another logic)
	// "Double if" construction means "If found: stop, else keep scanning", and "found" is state of whole process,
	// not only current conditional jump. Return in Inspect doesnt return parent function, that's why I need flag
	// "found" to be declared explicitly.
	found := false
	ast.Inspect(fun, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if found = fmt.Sprint(c.Fun) == callFunc; found {
				fmt.Printf("%v: Found direct call of os.Exit in main function\n", fset.Position(c.Pos()))
				return false
			}
		}
		return !found
	})
	return found
}

// Auxilary function for ExitCallInMainAnalyzer: searches import of pack package and its alias
func getPackageImportName(file *ast.File, pack string) (string, bool) {
	packName := ""
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if c, ok := n.(*ast.ImportSpec); ok && c.Path.Value == fmt.Sprintf(`"%s"`, pack) {
			found = true
			if c.Name == nil {
				packName = pack
			} else {
				packName = c.Name.Name
			}
		}
		return !found
	})
	return packName, found
}

// Auxilary function for ExitCallInMainAnalyzer: searches entry point (main() function of package main)
func isExitCalledInFile(file *ast.File, fset *token.FileSet) bool {
	// "found" is not only flag for Inspect termination.
	// This variable is ret value for this function (in case it is used by another logic)
	// "Double if" construction means "If found: stop, else keep scanning", and "found" is state of whole process,
	// not only current conditional jump. Return in Inspect doesnt return parent function, that's why I need flag
	// "found" to be declared explicitly.
	found := false

	if file == nil {
		return false
	}

	if file.Name.Name != "main" {
		return false
	}

	osName, osImported := getPackageImportName(file, "os")

	if !osImported {
		return false
	}

	callFunc := fmt.Sprintf("&{%s Exit}", osName)
	ast.Inspect(file, func(n ast.Node) bool {
		if c, ok := n.(*ast.FuncDecl); ok && c.Name.Name == "main" {
			found = isFuncCalledInFunc(n.(*ast.FuncDecl), callFunc, fset)
			return false
		}
		return !found
	})

	return found
}

// Custom analyzer (implements Increment 19 requirements)
func main() {
	// Rules slice
	var mychecks []*analysis.Analyzer

	// All rules, starting with "SA" and "S1000" from staticcheck
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") || v.Analyzer.Name == "S1000" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// All analyzers from tools/go/analysys package
	mychecks = append(mychecks, appends.Analyzer)
	mychecks = append(mychecks, asmdecl.Analyzer)
	mychecks = append(mychecks, assign.Analyzer)
	mychecks = append(mychecks, atomic.Analyzer)
	mychecks = append(mychecks, atomicalign.Analyzer)
	mychecks = append(mychecks, bools.Analyzer)
	mychecks = append(mychecks, buildssa.Analyzer)
	mychecks = append(mychecks, buildtag.Analyzer)
	mychecks = append(mychecks, cgocall.Analyzer)
	mychecks = append(mychecks, composite.Analyzer)
	mychecks = append(mychecks, copylock.Analyzer)
	mychecks = append(mychecks, ctrlflow.Analyzer)
	mychecks = append(mychecks, deepequalerrors.Analyzer)
	mychecks = append(mychecks, defers.Analyzer)
	mychecks = append(mychecks, directive.Analyzer)
	mychecks = append(mychecks, errorsas.Analyzer)
	//mychecks = append(mychecks, fieldalignment.Analyzer)
	mychecks = append(mychecks, findcall.Analyzer)
	mychecks = append(mychecks, framepointer.Analyzer)
	mychecks = append(mychecks, httpmux.Analyzer)
	mychecks = append(mychecks, httpresponse.Analyzer)
	mychecks = append(mychecks, ifaceassert.Analyzer)
	mychecks = append(mychecks, inspect.Analyzer)
	mychecks = append(mychecks, loopclosure.Analyzer)
	mychecks = append(mychecks, lostcancel.Analyzer)
	mychecks = append(mychecks, nilfunc.Analyzer)
	mychecks = append(mychecks, nilness.Analyzer)
	mychecks = append(mychecks, pkgfact.Analyzer)
	mychecks = append(mychecks, printf.Analyzer)
	mychecks = append(mychecks, reflectvaluecompare.Analyzer)
	mychecks = append(mychecks, shadow.Analyzer)
	mychecks = append(mychecks, shift.Analyzer)
	mychecks = append(mychecks, sigchanyzer.Analyzer)
	mychecks = append(mychecks, slog.Analyzer)
	mychecks = append(mychecks, sortslice.Analyzer)
	mychecks = append(mychecks, stdmethods.Analyzer)
	mychecks = append(mychecks, stringintconv.Analyzer)
	mychecks = append(mychecks, structtag.Analyzer)
	mychecks = append(mychecks, testinggoroutine.Analyzer)
	mychecks = append(mychecks, tests.Analyzer)
	mychecks = append(mychecks, timeformat.Analyzer)
	mychecks = append(mychecks, unmarshal.Analyzer)
	mychecks = append(mychecks, unreachable.Analyzer)
	mychecks = append(mychecks, unsafeptr.Analyzer)
	mychecks = append(mychecks, unusedresult.Analyzer)
	mychecks = append(mychecks, unusedwrite.Analyzer)
	mychecks = append(mychecks, usesgenerics.Analyzer)

	// My own analyzer for checking of Exit call in main func
	mychecks = append(mychecks, ExitCallInMainAnalyzer)

	// Two ore more other public analyzers
	mychecks = append(mychecks, ineffassign.Analyzer)
	mychecks = append(mychecks, bodyclose.Analyzer)

	multichecker.Main(
		mychecks...,
	)
}
