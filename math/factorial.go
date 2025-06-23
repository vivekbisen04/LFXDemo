// math/factorial.go
package math

// Factorial calculates the factorial of a non-negative integer n.
// It returns 1 for n = 0 as 0! is defined to be 1.
func Factorial(n int) int {
	if n < 0 {
		panic("Factorial is not defined for negative numbers")
	}
	if n == 0 || n == 1 {
		return 1
	}
	return n * Factorial(n-1)
}
