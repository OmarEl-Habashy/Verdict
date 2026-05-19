package dp

import "testing"

func TestFibonacciModular(t *testing.T) {
	cases := []struct {
		name string
		n    int64
		want int64
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"two", 2, 1},
		{"negative", -5, 0},
		{"small positive", 5, 5},
		{"large positive", 20, 6765},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FibonacciModular(tc.n)
			if tc.n > 2 && got != FibonacciModular(tc.n-1)+FibonacciModular(tc.n-2) {
				t.Errorf("FibonacciModular(%d) broke recurrence: got %d want %d", tc.n, got, FibonacciModular(tc.n-1)+FibonacciModular(tc.n-2))
			}
		})
	}
}

func TestFibonacciModularOptimized(t *testing.T) {
	cases := []struct {
		name string
		n    int64
		want int64
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"two", 2, 1},
		{"negative", -5, 0},
		{"small positive", 5, 5},
		{"large positive", 20, 6765},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FibonacciModularOptimized(tc.n)
			if tc.n > 2 && got != FibonacciModularOptimized(tc.n-1)+FibonacciModularOptimized(tc.n-2) {
				t.Errorf("FibonacciModularOptimized(%d) broke recurrence: got %d want %d", tc.n, got, FibonacciModularOptimized(tc.n-1)+FibonacciModularOptimized(tc.n-2))
			}
		})
	}
}

func TestFibonacciRange(t *testing.T) {
	cases := []struct {
		name  string
		start int64
		end   int64
		want  []int64
	}{
		{"single zero", 0, 0, []int64{0}},
		{"single one", 1, 1, []int64{1}},
		{"small range", 1, 5, []int64{1, 1, 2, 3, 5}},
		{"negative range", -3, 0, []int64{0}},
		{"large range", 5, 10, []int64{5, 8, 13, 21, 34, 55}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FibonacciRange(tc.start, tc.end)
			for i, v := range got {
				if v != FibonacciModular(tc.start+int64(i)) {
					t.Errorf("FibonacciRange(%d, %d)[%d] = %d; want %d", tc.start, tc.end, i, v, FibonacciModular(tc.start+int64(i)))
				}
			}
		})
	}
}