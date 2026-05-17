package mathutil

import "testing"

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
		name string
		a, b int
		want int
		err  bool
	}{
		{"normal", 6, 3, 2, false},
		{"zero", 1, 0, 0, true},
		{"negative", -4, 2, -2, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Div(tc.a, tc.b)
			if (err != nil) != tc.err {
				t.Errorf("Div(%d, %d) error = %v; wantErr %v", tc.a, tc.b, err, tc.err)
				return
			}
			if !tc.err && got != tc.want {
				t.Errorf("Div(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}