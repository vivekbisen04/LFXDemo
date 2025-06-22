package calculator

import (
	"testing"
)

func TestNewCalculator(t *testing.T) {
	// Test that NewCalculator creates a calculator with an empty history.
	calc := NewCalculator()
	if len(calc.history) != 0 {
		t.Error("Expected empty history, but got", calc.history)
	}
}

func TestAdd(t *testing.T) {
	calc := NewCalculator()

	// Test with positive numbers
	result := calc.Add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, but got %d", result)
	}
	if len(calc.history) != 1 || calc.history[0] != "2 + 3 = 5" {
		t.Errorf("Expected history to be updated correctly, but got %v", calc.history)
	}

	// Test with negative numbers
	result = calc.Add(-2, -3)
	if result != -5 {
		t.Errorf("Expected -5, but got %d", result)
	}
	if len(calc.history) != 2 || calc.history[1] != "-2 + -3 = -5" {
		t.Errorf("Expected history to be updated correctly, but got %v", calc.history)
	}

	// Test with zero
	result = calc.Add(5, 0)
	if result != 5 {
		t.Errorf("Expected 5, but got %d", result)
	}
	if len(calc.history) != 3 || calc.history[2] != "5 + 0 = 5" {
		t.Errorf("Expected history to be updated correctly, but got %v", calc.history)
	}

	// Test with large numbers
	result = calc.Add(1000000, 2000000)
	if result != 3000000 {
		t.Errorf("Expected 3000000, but got %d", result)
	}
}


func TestMultiply(t *testing.T) {
	calc := NewCalculator()

	// Test with positive numbers
	result := calc.Multiply(2, 3)
	if result != 6 {
		t.Errorf("Expected 6, but got %d", result)
	}
	if len(calc.history) != 1 || calc.history[0] != "2 * 3 = 6" {
		t.Errorf("Expected history to be updated correctly, but got %v", calc.history)
	}

	// Test with negative numbers
	result = calc.Multiply(-2, -3)
	if result != 6 {
		t.Errorf("Expected 6, but got %d", result)
	}
	if len(calc.history) != 2 || calc.history[1] != "-2 * -3 = 6" {
		t.Errorf("Expected history to be updated correctly, but got %v", calc.history)
	}

	// Test with zero
	result = calc.Multiply(5, 0)
	if result != 0 {
		t.Errorf("Expected 0, but got %d", result)
	}
	if len(calc.history) != 3 || calc.history[2] != "5 * 0 = 0" {
		t.Errorf("Expected history to be updated correctly, but got %v", calc.history)
	}

	// Test with large numbers
	result = calc.Multiply(1000, 2000)
	if result != 2000000 {
		t.Errorf("Expected 2000000, but got %d", result)
	}
}

func TestDivide(t *testing.T) {
	calc := NewCalculator()

	// Test with positive numbers
	result, err := calc.Divide(6, 3)
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	if result != 2 {
		t.Errorf("Expected 2, but got %d", result)
	}
	if len(calc.history) != 1 || calc.history[0] != "6 / 3 = 2" {
		t.Errorf("Expected history to be updated correctly, but got %v", calc.history)
	}

	// Test with division by zero
	_, err = calc.Divide(6, 0)
	if err == nil {
		t.Error("Expected error for division by zero, but got nil")
	}

	// Test with negative numbers
	result, err = calc.Divide(-6, -3)
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	if result != 2 {
		t.Errorf("Expected 2, but got %d", result)
	}

	// Test with negative and positive numbers
	result, err = calc.Divide(6, -3)
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	if result != -2 {
		t.Errorf("Expected -2, but got %d", result)
	}
}

func TestGetHistory(t *testing.T) {
	calc := NewCalculator()
	calc.Add(2,3)
	calc.Multiply(4,5)
	history := calc.GetHistory()
	if len(history) != 2 || history[0] != "2 + 3 = 5" || history[1] != "4 * 5 = 20"{
		t.Errorf("Expected history to be ['2 + 3 = 5', '4 * 5 = 20'], but got %v", history)
	}

	//Test empty history
	calc2 := NewCalculator()
	history = calc2.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, but got %v", history)
	}
}