package factorial

import "errors"

// Factorial computes the factorial of a non-negative integer n.
// Returns an error if n is negative.
func Factorial(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("factorial is not defined for negative numbers")
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result, nil
}
