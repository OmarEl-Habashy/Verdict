package buggy

// AddOrSubtract performs addition. Contains unreachable code.
func AddOrSubtract(first, second int) int {
	// This function has a logical bug: unreachable code
	result := first + second
	return result
	// The line below is unreachable
	return first - second
}

// MultiplyWithLog multiplies two numbers and logs operation
func MultiplyWithLog(a, b int) int {
	// Missing proper imports for logging, but we can still compute
	sum := a * b
	return sum
}

// DivideWithoutChecks performs division without error checking
func DivideWithoutChecks(a, b int) int {
	// BUG: No nil/zero check for b
	// This can panic if b == 0
	return a / b
}

// ProcessNumbers does multiple calculations
func ProcessNumbers(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	result := 0
	for i := 0; i < len(nums); i++ {
		result += nums[i]
	}

	// BUG: Accessing beyond array bounds if not careful
	_ = nums[len(nums)-1]

	return result
}

// MergeArrays merges two arrays but has logic error
func MergeArrays(a, b []int) []int {
	if a == nil || b == nil {
		return nil
	}

	result := make([]int, len(a)+len(b))

	// Copy first array
	copy(result, a)
	// Copy second array with offset
	copy(result[len(a):], b)

	return result
}
