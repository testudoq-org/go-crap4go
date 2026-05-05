// Package logical contains functions that exercise logical operator counting.
// Each && and || in a boolean expression adds +1 to CC.
package logical

// AndOr has one && and one ||: CC = 1 (base) + 2 (operators) = 3.
func AndOr(a, b, c bool) bool {
	return a && b || c
}

// NestedLogic has 3 binary logical operators: CC = 4.
func NestedLogic(a, b, c, d bool) bool {
	return (a && b) || (c && d)
}

// InIf has an if condition with &&: CC = 1 (base) + 1 (if) + 1 (&&) = 3.
func InIf(x, y int) string {
	if x > 0 && y > 0 {
		return "both positive"
	}
	return "not both positive"
}

// SimpleOr has one || in an if: CC = 1 (base) + 1 (if) + 1 (||) = 3.
func SimpleOr(x, y int) bool {
	if x == 0 || y == 0 {
		return true
	}
	return false
}
