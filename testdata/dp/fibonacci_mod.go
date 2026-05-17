package dp

const MOD = 1000000007

// FibonacciModular computes a modified Fibonacci sequence with modular arithmetic
func FibonacciModular(n int64) int64 {
	memo := make(map[int64]int64)

	var f func(int64) int64
	f = func(x int64) int64 {
		if x <= 2 {
			if x >= 1 {
				return 1
			}
			return 0
		}

		if val, exists := memo[x]; exists {
			return val
		}

		a := f(x / 2)
		b := f(x/2 + 1)

		var result int64
		if x%2 == 0 {
			result = (2*b - a) * a
		} else {
			result = a*a + b*b
		}

		// Handle negative modulo and apply MOD
		result = ((result % int64(MOD)) + int64(MOD)) % int64(MOD)
		memo[x] = result
		return result
	}

	return f(n)
}

// FibonacciModularOptimized uses iterative approach with memoization
func FibonacciModularOptimized(n int64) int64 {
	if n <= 2 {
		if n >= 1 {
			return 1
		}
		return 0
	}

	memo := make(map[int64]int64)
	return computeFibMemo(n, memo)
}

func computeFibMemo(n int64, memo map[int64]int64) int64 {
	if n <= 2 {
		if n >= 1 {
			return 1
		}
		return 0
	}

	if val, exists := memo[n]; exists {
		return val
	}

	a := computeFibMemo(n/2, memo)
	b := computeFibMemo(n/2+1, memo)

	var result int64
	if n%2 == 0 {
		result = (2*b - a) * a
	} else {
		result = a*a + b*b
	}

	result = ((result % int64(MOD)) + int64(MOD)) % int64(MOD)
	memo[n] = result
	return result
}

// FibonacciRange computes Fibonacci values for range of numbers
func FibonacciRange(start, end int64) []int64 {
	results := make([]int64, 0)
	memo := make(map[int64]int64)

	for i := start; i <= end; i++ {
		results = append(results, computeFibMemo(i, memo))
	}

	return results
}
