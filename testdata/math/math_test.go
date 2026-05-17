package mathutil

import (
	"testing"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"zero", 0, 0, 0},
		{"negative", -1, -2, -3},
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

func TestSub(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 5, 3, 2},
		{"zero", 0, 0, 0},
		{"negative", -1, -2, 1},
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

func TestMul(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 6},
		{"zero", 0, 5, 0},
		{"negative", -1, 2, -2},
		{"negative zero", 5, 0, 0},
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

func TestDiv(t *testing.T) {
	cases := []struct {
		name      string
		a, b     int
		want     int
		expectErr bool
	}{
		{"positive", 6, 3, 2, false},
		{"zero dividend", 0, 1, 0, false},
		{"negative dividend", -6, 2, -3, false},
		{"division by zero", 1, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Div(tc.a, tc.b)
			if tc.expectErr {
				if err == nil {
					t.Errorf("Div(%d, %d) = %d; expected an error", tc.a, tc.b, got)
				}
			} else {
				if err != nil {
					t.Errorf("Div(%d, %d) returned an error: %v", tc.a, tc.b, err)
				} else if got != tc.want {
					t.Errorf("Div(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
				}
			}
		})
	}
}