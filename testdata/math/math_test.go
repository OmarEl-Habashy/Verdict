package mathutil

import "testing"

// TestAdd tests the Add function.
func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"zero", 0, 0, 0},
		{"negative", -1, -2, -3},
		{"mixed", -1, 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Add(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Add(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// TestSub tests the Sub function.
func TestSub(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 5, 3, 2},
		{"zero", 0, 0, 0},
		{"negative", -1, -2, 1},
		{"mixed", 2, -2, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Sub(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Sub(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// TestMul tests the Mul function.
func TestMul(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 6},
		{"zero", 0, 5, 0},
		{"negative", -3, 2, -6},
		{"mixed", -4, -5, 20},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Mul(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Mul(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// TestDiv tests the Div function.
func TestDiv(t *testing.T) {
	cases := []struct {
		name     string
		a, b     int
		want     int
		wantErr  bool
	}{
		{"positive", 6, 3, 2, false},
		{"zero", 0, 1, 0, false},
		{"negative", -10, 2, -5, false},
		{"division by zero", 1, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Div(tc.a, tc.b)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Div(%d, %d) expected error, got none", tc.a, tc.b)
				}
				return
			}
			if err != nil {
				t.Errorf("Div(%d, %d) returned error: %v", tc.a, tc.b, err)
				return
			}
			if got != tc.want {
				t.Errorf("Div(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}