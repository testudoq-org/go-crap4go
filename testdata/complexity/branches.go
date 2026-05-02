// Package branches contains functions with various branching constructs
// to verify accurate CC counting for if, for, range, switch, type switch,
// select, and goto statements.
package branches

import "fmt"

// IfOnly has one if statement: CC = 2.
func IfOnly(x int) string {
	if x > 0 {
		return "positive"
	}
	return "non-positive"
}

// ElseIf has if + else if: CC = 3 (base + 2 conditions).
func ElseIf(x int) string {
	if x > 0 {
		return "positive"
	} else if x < 0 {
		return "negative"
	}
	return "zero"
}

// ForLoop has a plain for loop: CC = 2.
func ForLoop(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}

// RangeLoop has a range loop: CC = 2.
func RangeLoop(items []string) []string {
	result := make([]string, 0, len(items))
	for _, v := range items {
		result = append(result, v)
	}
	return result
}

// SwitchStmt has a switch with 3 non-default cases: CC = 4 (base + 3).
func SwitchStmt(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	default:
		return "other"
	}
}

// TypeSwitch has a type switch with 2 non-default type cases: CC = 3.
func TypeSwitch(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	default:
		return "unknown"
	}
}

// WithGoto has an if statement and a goto: CC = 3 (base + if + goto).
func WithGoto(n int) int {
	if n <= 0 {
		goto done
	}
	n--
done:
	return n
}

// SelectFunc demonstrates a select with 2 non-default cases: CC = 3.
func SelectFunc(ch1, ch2 <-chan int) int {
	select {
	case v := <-ch1:
		return v
	case v := <-ch2:
		return v * 2
	default:
		return 0
	}
}

// MixedBranches exercises multiple constructs in one function.
// CC = 1 (base) + 1 (if) + 1 (for) + 1 (range) = 4
func MixedBranches(items []int) string {
	if len(items) == 0 {
		return ""
	}
	total := 0
	for i := 0; i < len(items); i++ {
		total += items[i]
	}
	parts := make([]string, 0)
	for _, v := range items {
		parts = append(parts, fmt.Sprintf("%d", v))
	}
	_ = parts
	return fmt.Sprintf("total=%d", total)
}
