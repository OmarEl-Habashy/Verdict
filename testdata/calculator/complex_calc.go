package calculator

import (
	"errors"
)

// CalculatorOperation represents an operation type
type CalculatorOperation int

const (
	Add CalculatorOperation = iota
	Subtract
	Multiply
	Divide
	Modulo
	Power
)

// Calculate performs complex calculations with multiple branches
func Calculate(a, b int, op CalculatorOperation) (int, error) {
	switch op {
	case Add:
		return a + b, nil
	case Subtract:
		return a - b, nil
	case Multiply:
		return a * b, nil
	case Divide:
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	case Modulo:
		if b == 0 {
			return 0, errors.New("modulo by zero")
		}
		return a % b, nil
	case Power:
		if b < 0 {
			return 0, errors.New("negative exponent")
		}
		return power(a, b), nil
	default:
		return 0, errors.New("unknown operation")
	}
}

// power computes a^b using binary exponentiation
func power(base, exp int) int {
	if exp == 0 {
		return 1
	}

	result := 1
	for exp > 0 {
		if exp%2 == 1 {
			result *= base
		}
		base *= base
		exp /= 2
	}
	return result
}

// ChainedCalculation performs a sequence of calculations
func ChainedCalculation(values []int, operations []CalculatorOperation) (int, error) {
	if len(values) == 0 {
		return 0, errors.New("no values provided")
	}

	if len(values) != len(operations)+1 {
		return 0, errors.New("mismatched values and operations")
	}

	result := values[0]

	for i, op := range operations {
		newResult, err := Calculate(result, values[i+1], op)
		if err != nil {
			return 0, err
		}
		result = newResult
	}

	return result, nil
}

// CalculateWithCondition applies conditional logic before calculation
func CalculateWithCondition(a, b int, op CalculatorOperation) (int, error) {
	// Branch 1: Handle special cases
	if a == 0 && b == 0 {
		switch op {
		case Add, Subtract:
			return 0, nil
		case Multiply, Divide, Modulo:
			return 0, errors.New("undefined operation on zero")
		}
	}

	// Branch 2: Handle negative numbers
	if a < 0 || b < 0 {
		switch op {
		case Divide:
			if b == 0 {
				return 0, errors.New("division by zero")
			}
			return a / b, nil
		case Modulo:
			if b == 0 {
				return 0, errors.New("modulo by zero")
			}
			return a % b, nil
		case Power:
			return 0, errors.New("negative base in power")
		}
	}

	// Branch 3: Normal calculation
	return Calculate(a, b, op)
}

// AdvancedCalculate performs calculation with limit checks
func AdvancedCalculate(a, b int, op CalculatorOperation, maxResult int) (int, error) {
	result, err := Calculate(a, b, op)
	if err != nil {
		return 0, err
	}

	if result > maxResult {
		return 0, errors.New("result exceeds maximum")
	}

	if result < -maxResult {
		return 0, errors.New("result below minimum")
	}

	return result, nil
}
