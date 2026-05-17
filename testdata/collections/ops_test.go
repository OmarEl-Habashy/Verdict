package collections

import (
	"testing"
)

func TestFindMax(t *testing.T) {
	cases := []struct {
		name    string
		numbers []int
		want    int
		wantErr bool
	}{
		{"empty slice", []int{}, 0, true},
		{"single element", []int{5}, 5, false},
		{"multiple elements", []int{1, 5, 3, 4}, 5, false},
		{"negative and positive", []int{-1, -5, 0, 2}, 2, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FindMax(tc.numbers)
			if (err != nil) != tc.wantErr {
				t.Errorf("FindMax() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("FindMax() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestFindMin(t *testing.T) {
	cases := []struct {
		name    string
		numbers []int
		want    int
		wantErr bool
	}{
		{"empty slice", []int{}, 0, true},
		{"single element", []int{5}, 5, false},
		{"multiple elements", []int{3, 1, 4, 2}, 1, false},
		{"negative and positive", []int{-1, -5, 0, 2}, -5, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FindMin(tc.numbers)
			if (err != nil) != tc.wantErr {
				t.Errorf("FindMin() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("FindMin() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRemoveElement(t *testing.T) {
	cases := []struct {
		name   string
		slice  []int
		target int
		want   []int
	}{
		{"remove from non-empty", []int{1, 2, 3, 2}, 2, []int{1, 3}},
		{"remove from empty", []int{}, 1, []int{}},
		{"remove non-existing", []int{1, 2, 3}, 4, []int{1, 2, 3}},
		{"remove all elements", []int{2, 2, 2}, 2, []int{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RemoveElement(tc.slice, tc.target)
			if !equal(got, tc.want) {
				t.Errorf("RemoveElement() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	cases := []struct {
		name string
		a, b []int
		want []int
	}{
		{"no common", []int{1, 2, 3}, []int{4, 5, 6}, []int{}},
		{"some common", []int{1, 2, 3}, []int{3, 4, 5}, []int{3}},
		{"all common", []int{1, 2, 3}, []int{1, 2, 3}, []int{1, 2, 3}},
		{"empty slices", []int{}, []int{}, []int{}},
		{"one empty slice", []int{1, 2}, []int{}, []int{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Intersection(tc.a, tc.b)
			if !equal(got, tc.want) {
				t.Errorf("Intersection() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUnion(t *testing.T) {
	cases := []struct {
		name string
		a, b []int
		want []int
	}{
		{"both empty", []int{}, []int{}, []int{}},
		{"a is empty", []int{}, []int{1, 2}, []int{1, 2}},
		{"b is empty", []int{1, 2}, []int{}, []int{1, 2}},
		{"no common", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{"some common", []int{1, 2}, []int{2, 3}, []int{1, 2, 3}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Union(tc.a, tc.b)
			if !equal(got, tc.want) {
				t.Errorf("Union() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGroupByValue(t *testing.T) {
	cases := []struct {
		name    string
		numbers []int
		want    map[int]int
	}{
		{"empty slice", []int{}, map[int]int{}},
		{"single element", []int{1}, map[int]int{1: 1}},
		{"multiple elements", []int{1, 2, 2, 3}, map[int]int{1: 1, 2: 2, 3: 1}},
		{"all identical", []int{1, 1, 1}, map[int]int{1: 3}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GroupByValue(tc.numbers)
			if !equalMaps(got, tc.want) {
				t.Errorf("GroupByValue() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	cases := []struct {
		name  string
		slice []int
		want  []int
	}{
		{"empty slice", []int{}, []int{}},
		{"single element", []int{1}, []int{1}},
		{"multiple elements", []int{1, 2, 3}, []int{3, 2, 1}},
		{"all identical", []int{1, 1, 1}, []int{1, 1, 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Reverse(tc.slice)
			if !equal(got, tc.want) {
				t.Errorf("Reverse() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRotate(t *testing.T) {
	cases := []struct {
		name  string
		slice []int
		k     int
		want  []int
	}{
		{"empty slice", []int{}, 2, []int{}},
		{"single element", []int{1}, 1, []int{1}},
		{"rotate by length", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"rotate by one", []int{1, 2, 3}, 1, []int{3, 1, 2}},
		{"negative rotate", []int{1, 2, 3}, -1, []int{2, 3, 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Rotate(tc.slice, tc.k)
			if !equal(got, tc.want) {
				t.Errorf("Rotate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterOdd(t *testing.T) {
	cases := []struct {
		name    string
		numbers []int
		want    []int
	}{
		{"empty slice", []int{}, []int{}},
		{"no odd", []int{2, 4, 6}, []int{}},
		{"some odd", []int{1, 2, 3, 4}, []int{1, 3}},
		{"all odd", []int{1, 3, 5}, []int{1, 3, 5}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterOdd(tc.numbers)
			if !equal(got, tc.want) {
				t.Errorf("FilterOdd() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterEven(t *testing.T) {
	cases := []struct {
		name    string
		numbers []int
		want    []int
	}{
		{"empty slice", []int{}, []int{}},
		{"no even", []int{1, 3, 5}, []int{}},
		{"some even", []int{1, 2, 3, 4}, []int{2, 4}},
		{"all even", []int{2, 4, 6}, []int{2, 4, 6}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterEven(tc.numbers)
			if !equal(got, tc.want) {
				t.Errorf("FilterEven() = %v, want %v", got, tc.want)
			}
		})
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalMaps(a, b map[int]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}