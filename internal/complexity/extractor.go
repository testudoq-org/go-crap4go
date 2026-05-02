// Package complexity provides AST-based cyclomatic complexity (CC) analysis
// for Go source files using the standard go/ast and go/token packages.
//
// # CC Counting Rules
//
// These rules are adapted from crap4js for Go-specific constructs.
// Each function/method/closure starts with a base CC of 1.
//
//	Construct                          Δ CC
//	─────────────────────────────────────────
//	Base (per func/method/closure)     +1
//	IfStmt                             +1
//	ForStmt                            +1
//	RangeStmt                          +1
//	SwitchStmt  (per non-default case) +1
//	TypeSwitchStmt (per non-default)   +1
//	SelectStmt  (per non-default comm) +1
//	BinaryExpr  &&  or  ||             +1
//	GotoStmt                           +1
//
// Closures (ast.FuncLit) are reported as separate Function entries; their
// internal branches do NOT contribute to the enclosing function's CC.
//
// # Method Naming
//
// Methods are formatted as "ReceiverType.MethodName". Pointer receivers
// ("*T") are normalised to "T.MethodName".
// Anonymous functions are named "<anonymous:line>".
package complexity

import (
	"go/ast"
	"go/token"
)

// Function holds extracted metadata and CC for a single Go function,
// method, or closure found during AST analysis.
type Function struct {
	// Name is the display name of the function.
	//   Named function: "FuncName"
	//   Method:         "ReceiverType.MethodName"  (pointer receiver → "T.MethodName")
	//   Closure:        "<anonymous:line>"
	Name string

	// File is the source file path as provided to Extract.
	File string

	// StartLine is the 1-based line number where the function body begins
	// (the opening brace of the function body).
	StartLine int

	// EndLine is the 1-based line number where the function body ends
	// (the closing brace of the function body).
	EndLine int

	// CC is the cyclomatic complexity of the function (minimum 1).
	CC int

	// LOC is the count of non-blank, non-comment source lines inside the
	// function body, computed from token.FileSet positions.
	LOC int
}

// Extract analyses a single parsed Go source file and returns one Function
// descriptor per top-level function, method, and closure found.
//
// fileSet must be the token.FileSet that was used to parse file.
// filename is stored verbatim in each returned Function.File field.
//
// Extract never returns a nil slice. On a partial error it returns whatever
// results were accumulated before the failure alongside a non-nil error.
//
// TODO(prompt-2): implement full CC extraction logic.
func Extract(fileSet *token.FileSet, file *ast.File, filename string) ([]*Function, error) {
	// Stub implementation — full logic is added in Prompt 2.
	_ = fileSet
	_ = file
	_ = filename
	return []*Function{}, nil
}
