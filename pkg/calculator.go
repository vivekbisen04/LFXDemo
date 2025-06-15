package main

import (
	"errors"
	"fmt"
	"strings"
)

// Calculator provides basic mathematical operations
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
func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	c.history = append(c.history, fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result))
	return result
}

// Subtract performs subtraction of two numbers
func (c *Calculator) Subtract(a, b float64) float64 {
	result := a - b
	c.history = append(c.history, fmt.Sprintf("%.2f - %.2f = %.2f", a, b, result))
	return result
}

// Multiply performs multiplication of two numbers
func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	c.history = append(c.history, fmt.Sprintf("%.2f * %.2f = %.2f", a, b, result))
	return result
}

// Divide performs division of two numbers
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero is not allowed")
	}
	result := a / b
	c.history = append(c.history, fmt.Sprintf("%.2f / %.2f = %.2f", a, b, result))
	return result, nil
}

// GetHistory returns the calculation history
func (c *Calculator) GetHistory() []string {
	return c.history
}

// ClearHistory clears the calculation history
func (c *Calculator) ClearHistory() {
	c.history = make([]string, 0)
}

// StringProcessor provides string manipulation utilities
type StringProcessor struct{}

// NewStringProcessor creates a new string processor
func NewStringProcessor() *StringProcessor {
	return &StringProcessor{}
}

// Reverse reverses a string
func (sp *StringProcessor) Reverse(s string) string {
	if s == "" {
		return ""
	}
	
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome checks if a string is a palindrome
func (sp *StringProcessor) IsPalindrome(s string) bool {
	if s == "" {
		return true
	}
	
	// Convert to lowercase and remove spaces for comparison
	cleaned := strings.ToLower(strings.ReplaceAll(s, " ", ""))
	return cleaned == sp.Reverse(cleaned)
}

// CountWords counts the number of words in a string
func (sp *StringProcessor) CountWords(s string) int {
	if s == "" {
		return 0
	}
	
	words := strings.Fields(s)
	return len(words)
}

// Capitalize capitalizes the first letter of each word
func (sp *StringProcessor) Capitalize(s string) string {
	if s == "" {
		return ""
	}
	
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	
	return strings.Join(words, " ")
}

func main() {
	calc := NewCalculator()
	fmt.Printf("Addition: %.2f\n", calc.Add(10, 5))
	fmt.Printf("Division: %.2f\n", func() float64 {
		result, _ := calc.Divide(10, 2)
		return result
	}())
	
	sp := NewStringProcessor()
	fmt.Printf("Reverse: %s\n", sp.Reverse("hello"))
	fmt.Printf("Is Palindrome: %t\n", sp.IsPalindrome("racecar"))
}