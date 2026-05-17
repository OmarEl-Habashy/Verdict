package collections

import (
	"errors"
	"sort"
)

// FindMax finds the maximum value in a slice
func FindMax(numbers []int) (int, error) {
	if len(numbers) == 0 {
		return 0, errors.New("empty slice")
	}

	max := numbers[0]
	for _, num := range numbers {
		if num > max {
			max = num
		}
	}

	return max, nil
}

// FindMin finds the minimum value in a slice
func FindMin(numbers []int) (int, error) {
	if len(numbers) == 0 {
		return 0, errors.New("empty slice")
	}

	min := numbers[0]
	for _, num := range numbers {
		if num < min {
			min = num
		}
	}

	return min, nil
}

// RemoveElement removes all occurrences of target element
func RemoveElement(slice []int, target int) []int {
	result := make([]int, 0)
	for _, elem := range slice {
		if elem != target {
			result = append(result, elem)
		}
	}
	return result
}

// Intersection finds common elements between two slices
func Intersection(a, b []int) []int {
	if len(a) == 0 || len(b) == 0 {
		return []int{}
	}

	mapB := make(map[int]bool)
	for _, elem := range b {
		mapB[elem] = true
	}

	result := make([]int, 0)
	seen := make(map[int]bool)

	for _, elem := range a {
		if mapB[elem] && !seen[elem] {
			result = append(result, elem)
			seen[elem] = true
		}
	}

	return result
}

// Union combines two slices, removing duplicates
func Union(a, b []int) []int {
	if len(a) == 0 && len(b) == 0 {
		return []int{}
	}

	seen := make(map[int]bool)
	result := make([]int, 0)

	for _, elem := range a {
		if !seen[elem] {
			result = append(result, elem)
			seen[elem] = true
		}
	}

	for _, elem := range b {
		if !seen[elem] {
			result = append(result, elem)
			seen[elem] = true
		}
	}

	sort.Ints(result)
	return result
}

// GroupByValue groups elements by their values
func GroupByValue(numbers []int) map[int]int {
	if len(numbers) == 0 {
		return make(map[int]int)
	}

	groups := make(map[int]int)
	for _, num := range numbers {
		groups[num]++
	}

	return groups
}

// Reverse reverses a slice in place
func Reverse(slice []int) []int {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

// Rotate rotates slice by k positions
func Rotate(slice []int, k int) []int {
	if len(slice) == 0 {
		return slice
	}

	k = k % len(slice)
	result := make([]int, len(slice))

	copy(result, slice[len(slice)-k:])
	copy(result[k:], slice[:len(slice)-k])

	return result
}

// FilterOdd returns only odd numbers
func FilterOdd(numbers []int) []int {
	result := make([]int, 0)
	for _, num := range numbers {
		if num%2 != 0 {
			result = append(result, num)
		}
	}
	return result
}

// FilterEven returns only even numbers
func FilterEven(numbers []int) []int {
	result := make([]int, 0)
	for _, num := range numbers {
		if num%2 == 0 {
			result = append(result, num)
		}
	}
	return result
}
