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
		{"negative", -1, 2, -3},
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
		{"negative", -1, 3, -3},
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
		name  string
		a, b  int
		want  int
		isErr bool
	}{
		{"positive", 6, 3, 2, false},
		{"zero", 0, 1, 0, false},
		{"division by zero", 1, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Div(tc.a, tc.b)
			if tc.isErr {
				if err == nil {
					t.Errorf("Div(%d, %d) expected an error; got nil", tc.a, tc.b)
				}
			} else {
				if err != nil {
					t.Errorf("Div(%d, %d) = error; want %d", tc.a, tc.b, tc.want)
				} else if got != tc.want {
					t.Errorf("Div(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
				}
			}
		})
	}
}