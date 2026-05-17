package mathutil

import "fmt"

// Add returns the sum of a and b.
func Add(a, b int) int {
	return a + b
}

// Sub returns the difference of a and b.
func Sub(a, b int) int {
	return a - b
}

// Mul returns the product of a and b.
func Mul(a, b int) int {
	return a * b
}

// Div returns a divided by b and an error if b is zero.
func Div(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("div: division by zero")
	}
	return a / b, nil
}
