package buggy

import "testing"

func TestAddOrSubtract(t *testing.T) {
	cases := []struct {
		name   string
		first  int
		second int
		want   int
	}{
		{"positive", 2, 3, 5},
		{"zero", 0, 0, 0},
		{"negative", -1, -2, -3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := AddOrSubtract(tc.first, tc.second)
			if got != tc.want {
				t.Errorf("AddOrSubtract(%d, %d) = %d; want %d", tc.first, tc.second, got, tc.want)
			}
		})
	}
}

func TestMultiplyWithLog(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 6},
		{"zero", 0, 5, 0},
		{"negative", -2, 3, -6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MultiplyWithLog(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("MultiplyWithLog(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestDivideWithoutChecks(t *testing.T) {
	cases := []struct {
		name     string
		a, b     int
		want     int
		wantPanic bool
	}{
		{"positive", 6, 2, 3, false},
		{"zero", 0, 1, 0, false},
		{"division by zero", 1, 0, 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("DivideWithoutChecks(%d, %d) did not panic", tc.a, tc.b)
					}
				}()
			}
			got := DivideWithoutChecks(tc.a, tc.b)
			if !tc.wantPanic && got != tc.want {
				t.Errorf("DivideWithoutChecks(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestProcessNumbers(t *testing.T) {
	cases := []struct {
		name  string
		input []int
		want  int
	}{
		{"empty", []int{}, 0},
		{"single element", []int{5}, 5},
		{"multiple elements", []int{1, 2, 3}, 6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ProcessNumbers(tc.input)
			if got != tc.want {
				t.Errorf("ProcessNumbers(%v) = %d; want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestMergeArrays(t *testing.T) {
	cases := []struct {
		name     string
		a, b     []int
		expected []int
	}{
		{"both non-empty", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{"first empty", []int{}, []int{1, 2}, []int{1, 2}},
		{"second empty", []int{1, 2}, []int{}, []int{1, 2}},
		{"both empty", []int{}, []int{}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeArrays(tc.a, tc.b)
			if len(got) != len(tc.expected) {
				t.Errorf("MergeArrays(%v, %v) length = %d; want %d", tc.a, tc.b, len(got), len(tc.expected))
				return
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("MergeArrays(%v, %v) = %v; want %v", tc.a, tc.b, got, tc.expected)
					return
				}
			}
		})
	}
}