package graph

import "testing"

func TestCountConnectedComponents(t *testing.T) {
	cases := []struct {
		name  string
		n     int
		edges [][2]int
		want  int
	}{
		{"no nodes", 0, nil, 0},
		{"one node", 1, nil, 1},
		{"two disconnected nodes", 2, nil, 2},
		{"two connected nodes", 2, [][2]int{{1, 2}}, 1},
		{"three nodes, one component", 3, [][2]int{{1, 2}, {2, 3}}, 1},
		{"three nodes, two components", 3, [][2]int{{1, 2}}, 2},
		{"multiple components", 5, [][2]int{{1, 2}, {3, 4}}, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CountConnectedComponents(tc.n, tc.edges)
			if got != tc.want {
				t.Errorf("CountConnectedComponents(%d, %v) = %d; want %d", tc.n, tc.edges, got, tc.want)
			}
		})
	}
}

func TestFindAllComponents(t *testing.T) {
	cases := []struct {
		name  string
		n     int
		edges [][2]int
		want  [][]int
	}{
		{"no nodes", 0, nil, [][]int{}},
		{"one node", 1, nil, [][]int{{1}}},
		{"two disconnected nodes", 2, nil, [][]int{{1}, {2}}},
		{"two connected nodes", 2, [][2]int{{1, 2}}, [][]int{{1, 2}}},
		{"three nodes, one component", 3, [][2]int{{1, 2}, {2, 3}}, [][]int{{1, 2, 3}}},
		{"three nodes, two components", 3, [][2]int{{1, 2}}, [][]int{{1, 2}, {3}}},
		{"multiple components", 5, [][2]int{{1, 2}, {3, 4}}, [][]int{{1, 2}, {3, 4}, {5}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindAllComponents(tc.n, tc.edges)
			if len(got) != len(tc.want) {
				t.Errorf("FindAllComponents(%d, %v) = %v; want %v", tc.n, tc.edges, got, tc.want)
				return
			}
			for i := range got {
				if len(got[i]) != len(tc.want[i]) {
					t.Errorf("FindAllComponents(%d, %v) = %v; want %v", tc.n, tc.edges, got, tc.want)
					return
				}
			}
		})
	}
}