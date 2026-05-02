// Package closures contains functions with nested function literals (closures)
// to verify that closures are reported as separate entries and do not
// contribute their branch counts to the enclosing function's CC.
package closures

// WithClosure contains one if statement in the outer function and a closure
// with its own if statement inside.
//
// Expected extraction:
//   - "WithClosure"         CC = 2 (base + outer if)
//   - "<anonymous:line>"    CC = 2 (base + closure if)
func WithClosure(x int) func(int) string {
	if x < 0 {
		x = 0
	}
	return func(y int) string {
		if y > x {
			return "greater"
		}
		return "not greater"
	}
}

// MultiClosure contains two closures in the same function.
//
// Expected extraction:
//   - "MultiClosure"        CC = 1 (base, no branches in outer body)
//   - "<anonymous:line>"    CC = 2 (base + if in first closure)
//   - "<anonymous:line>"    CC = 2 (base + for in second closure)
func MultiClosure() (func(int) bool, func([]int) int) {
	isPositive := func(n int) bool {
		if n > 0 {
			return true
		}
		return false
	}
	sumAll := func(nums []int) int {
		total := 0
		for _, v := range nums {
			total += v
		}
		return total
	}
	return isPositive, sumAll
}
