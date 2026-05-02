package foo

// Simple returns 1. CC = 1 (no branches).
func Simple() int {
	return 1
}

// Branch returns "positive" or "non-positive" based on x. CC = 2.
func Branch(x int) string {
	if x > 0 {
		return "positive"
	}
	return "non-positive"
}
