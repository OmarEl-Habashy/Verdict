package tree

import "testing"

func TestMaxTreeDepth(t *testing.T) {
	cases := []struct {
		name      string
		managers  []int
		hierarchy map[int][]int
		expected  int
	}{
		{"empty hierarchy", []int{}, nil, 0},
		{"single manager no children", []int{1}, map[int][]int{1: {}}, 1},
		{"multiple managers with children", []int{1, 2}, map[int][]int{
			1: {2, 3},
			2: {4},
		}, 3},
		{"no children", []int{1}, map[int][]int{1: {}}, 1},
		{"deep hierarchy", []int{1}, map[int][]int{
			1: {2},
			2: {3},
			3: {},
		}, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MaxTreeDepth(tc.managers, tc.hierarchy)
			if got != tc.expected {
				t.Errorf("MaxTreeDepth(%v, %v) = %d; want %d", tc.managers, tc.hierarchy, got, tc.expected)
			}
		})
	}
}

func TestBuildHierarchy(t *testing.T) {
	cases := []struct {
		name     string
		n        int
		parentOf []int
		expectedManagers []int
		expectedHierarchy map[int][]int
	}{
		{"no managers", 0, []int{}, []int{}, map[int][]int{}},
		{"single manager", 1, []int{-1}, []int{1}, map[int][]int{}},
		{"multiple managers with relationships", 4, []int{-1, 1, 1, 2}, []int{1}, map[int][]int{
			1: {2, 3},
			2: {4},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotManagers, gotHierarchy := BuildHierarchy(tc.n, tc.parentOf)
			if !equalSlices(gotManagers, tc.expectedManagers) {
				t.Errorf("BuildHierarchy(%d, %v) = %v; want %v", tc.n, tc.parentOf, gotManagers, tc.expectedManagers)
			}
			if !equalHierarchies(gotHierarchy, tc.expectedHierarchy) {
				t.Errorf("BuildHierarchy(%d, %v) = %v; want %v", tc.n, tc.parentOf, gotHierarchy, tc.expectedHierarchy)
			}
		})
	}
}

func TestFindDeepestNode(t *testing.T) {
	cases := []struct {
		name           string
		managers       []int
		hierarchy      map[int][]int
		expectedNode   int
		expectedDepth  int
	}{
		{"no hierarchy", []int{}, nil, -1, 0},
		{"single node", []int{1}, map[int][]int{1: {}}, 1, 1},
		{"deeper child", []int{1}, map[int][]int{
			1: {2},
			2: {3},
		}, 3, 3},
		{"multiple branches", []int{1}, map[int][]int{
			1: {2, 3},
			2: {4},
			3: {},
		}, 4, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotNode, gotDepth := FindDeepestNode(tc.managers, tc.hierarchy)
			if gotNode != tc.expectedNode || gotDepth != tc.expectedDepth {
				t.Errorf("FindDeepestNode(%v, %v) = (%d, %d); want (%d, %d)", tc.managers, tc.hierarchy, gotNode, gotDepth, tc.expectedNode, tc.expectedDepth)
			}
		})
	}
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func equalHierarchies(a, b map[int][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !equalSlices(v, b[k]) {
			return false
		}
	}
	return true
}