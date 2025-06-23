package math

import (
	"testing"
)

func TestFactorial_Zero(t *testing.T) {
	// Test case for n = 0
	expected := 1
	actual := Factorial(0)
	if actual != expected {
		t.Errorf("Factorial(0) = %d; want %d", actual, expected)
	}
}

func TestFactorial_One(t *testing.T) {
	// Test case for n = 1
	expected := 1
	actual := Factorial(1)
	if actual != expected {
		t.Errorf("Factorial(1) = %d; want %d", actual, expected)
	}
}

func TestFactorial_Positive(t *testing.T) {
	// Test case for a positive integer
	expected := 120
	actual := Factorial(5)
	if actual != expected {
		t.Errorf("Factorial(5) = %d; want %d", actual, expected)
	}
}

func TestFactorial_LargePositive(t *testing.T) {
	// Test case for a larger positive integer.  May need adjustment depending on int size limits.
	expected := 3628800
	actual := Factorial(10)
	if actual != expected {
		t.Errorf("Factorial(10) = %d; want %d", actual, expected)
	}
}


func TestFactorial_Negative(t *testing.T) {
	// Test case for a negative integer; should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	Factorial(-1)
}

func TestFactorial_Overflow(t *testing.T){
	//Test case to check for potential integer overflow.  Adjust input as needed.
	defer func() {
		if r := recover(); r == nil {
			t.Log("No panic, potential overflow not detected.  Increase input value for more robust testing.")
		} else {
			t.Log("Panic detected as expected for potential overflow.")
		}
	}()
	Factorial(20) // Adjust this value to trigger overflow if your int size is smaller.
}