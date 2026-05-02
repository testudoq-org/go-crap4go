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
	"fmt"
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
func Extract(fileSet *token.FileSet, file *ast.File, filename string) ([]*Function, error) {
	results := make([]*Function, 0)
	e := &extractor{fset: fileSet, file: filename, results: &results}
	e.visitDecls(file.Decls)
	return results, nil
}

// ---------------------------------------------------------------------------
// internal extractor
// ---------------------------------------------------------------------------

type extractor struct {
	fset    *token.FileSet
	file    string
	results *[]*Function
}

// visitDecls processes all top-level declarations in a file.
func (e *extractor) visitDecls(decls []ast.Decl) {
	for _, d := range decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			e.visitFuncDecl(fn)
		}
	}
}

// visitFuncDecl handles a named top-level function or method declaration.
func (e *extractor) visitFuncDecl(fn *ast.FuncDecl) {
	if fn.Body == nil {
		return // external / forward declaration
	}
	name := funcDeclName(fn)
	pos := e.fset.Position(fn.Body.Lbrace)
	end := e.fset.Position(fn.Body.Rbrace)

	f := &Function{
		Name:      name,
		File:      e.file,
		StartLine: pos.Line,
		EndLine:   end.Line,
		CC:        1, // base
	}

	// Walk the body, counting CC-contributing nodes but stopping at nested
	// FuncLit boundaries (those are reported as separate entries).
	countCC(fn.Body, f, e)
	f.LOC = countLOC(e.fset, fn.Body)

	*e.results = append(*e.results, f)
}

// visitFuncLit handles a closure (anonymous function literal).
func (e *extractor) visitFuncLit(lit *ast.FuncLit, outerStartLine int) {
	pos := e.fset.Position(lit.Body.Lbrace)
	end := e.fset.Position(lit.Body.Rbrace)

	name := fmt.Sprintf("<anonymous:%d>", pos.Line)

	f := &Function{
		Name:      name,
		File:      e.file,
		StartLine: pos.Line,
		EndLine:   end.Line,
		CC:        1, // base
	}

	countCC(lit.Body, f, e)
	f.LOC = countLOC(e.fset, lit.Body)

	*e.results = append(*e.results, f)
}

// ---------------------------------------------------------------------------
// CC counting
// ---------------------------------------------------------------------------

// countCC walks stmtBody and increments fn.CC for each CC-contributing node.
// When a nested FuncLit is encountered, countCC delegates to the extractor
// so the closure is reported as a separate entry rather than inflating fn.CC.
func countCC(body *ast.BlockStmt, fn *Function, e *extractor) {
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		switch node := n.(type) {
		case *ast.FuncLit:
			// Closure — report separately; do NOT descend into it for fn.
			e.visitFuncLit(node, fn.StartLine)
			return false // stop descending for this branch

		case *ast.IfStmt:
			fn.CC++

		case *ast.ForStmt:
			fn.CC++

		case *ast.RangeStmt:
			fn.CC++

		case *ast.CaseClause:
			// SwitchStmt and TypeSwitchStmt share CaseClause.
			// Only non-default cases add to CC.
			if node.List != nil {
				fn.CC++
			}

		case *ast.CommClause:
			// SelectStmt communication clause.
			// Only non-default comms add to CC.
			if node.Comm != nil {
				fn.CC++
			}

		case *ast.BinaryExpr:
			if node.Op.String() == "&&" || node.Op.String() == "||" {
				fn.CC++
			}

		case *ast.BranchStmt:
			// GotoStmt is represented as BranchStmt with Tok == token.GOTO.
			if node.Tok == token.GOTO {
				fn.CC++
			}
		}
		return true
	})
}

// ---------------------------------------------------------------------------
// LOC counting
// ---------------------------------------------------------------------------

// countLOC counts the number of source lines that contain at least one token
// within the function body (approximation of non-blank, non-comment lines).
func countLOC(fset *token.FileSet, body *ast.BlockStmt) int {
	if body == nil {
		return 0
	}
	startLine := fset.Position(body.Lbrace).Line
	endLine := fset.Position(body.Rbrace).Line
	if endLine <= startLine {
		return 0
	}
	// Count lines that contain at least one AST node token.
	seen := make(map[int]struct{})
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		line := fset.Position(n.Pos()).Line
		if line > startLine && line < endLine {
			seen[line] = struct{}{}
		}
		return true
	})
	return len(seen)
}

// ---------------------------------------------------------------------------
// naming helpers
// ---------------------------------------------------------------------------

// funcDeclName returns the display name for a FuncDecl.
// Methods: "ReceiverType.MethodName" (pointer receiver → "T.MethodName").
// Functions: "FuncName".
func funcDeclName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	recv := fn.Recv.List[0].Type
	typeName := receiverTypeName(recv)
	return typeName + "." + fn.Name.Name
}

// receiverTypeName extracts the base type name from a receiver type expression,
// stripping pointer indirection so "*T" → "T".
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		// Generic receiver: T[P] → T
		return receiverTypeName(t.X)
	case *ast.IndexListExpr:
		// Generic receiver with multiple type params: T[P, Q] → T
		return receiverTypeName(t.X)
	default:
		return "unknown"
	}
}

