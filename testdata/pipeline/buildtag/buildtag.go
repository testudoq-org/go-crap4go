//go:build ignore

// Package buildtag is a testdata fixture whose build constraint (ignore)
// is never satisfied. The pipeline must exclude this file silently.
package buildtag

// IgnoredFunction must never appear in pipeline analysis results.
func IgnoredFunction() {}
