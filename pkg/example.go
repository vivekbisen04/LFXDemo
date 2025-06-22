package calculator

import (
	"errors"
	"fmt"
)

// Calculator provides basic arithmetic operations
type Calculator struct {
	history []string
}

// NewCalculator creates a new calculator instance
func NewCalculator() *Calculator {
	return &Calculator{
		history: make([]string, 0),
	}
}

// Add performs addition of two numbers
func (c *Calculator) Add(a, b int) int {
	result := a + b
	c.history = append(c.history, fmt.Sprintf("%d + %d = %d", a, b, result))
	return result
}

// Multiply performs multiplication of two numbers
func (c *Calculator) Multiply(a, b int) int {
	result := a * b
	c.history = append(c.history, fmt.Sprintf("%d * %d = %d", a, b, result))
	return result
}

// Divide performs division with error handling
func (c *Calculator) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	result := a / b
	c.history = append(c.history, fmt.Sprintf("%d / %d = %d", a, b, result))
	return result, nil
}

// GetHistory returns the calculation history
func (c *Calculator) GetHistory() []string {
	return c.history
}
