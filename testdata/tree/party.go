package tree

// Node represents a person in the hierarchy
type Node struct {
	ID       int
	Manager  int
	Children []int
}

// MaxTreeDepth finds the maximum depth in a forest of trees
func MaxTreeDepth(managers []int, hierarchy map[int][]int) int {
	maxDepth := 0

	var dfs func(int) int
	dfs = func(nodeID int) int {
		if hierarchy == nil {
			return 1
		}

		children, exists := hierarchy[nodeID]
		if !exists || len(children) == 0 {
			return 1
		}

		maxChildDepth := 0
		for _, child := range children {
			childDepth := dfs(child)
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}

		return maxChildDepth + 1
	}

	for _, manager := range managers {
		depth := dfs(manager)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// BuildHierarchy constructs the tree structure from manager relationships
func BuildHierarchy(n int, parentOf []int) ([]int, map[int][]int) {
	hierarchy := make(map[int][]int)
	var managers []int

	for i := 1; i <= n; i++ {
		if parentOf[i] == -1 {
			managers = append(managers, i)
		} else {
			hierarchy[parentOf[i]] = append(hierarchy[parentOf[i]], i)
		}
	}

	return managers, hierarchy
}

// FindDeepestNode returns the deepest node and its depth
func FindDeepestNode(managers []int, hierarchy map[int][]int) (int, int) {
	maxDepth := 0
	deepestNode := -1

	var dfs func(int) int
	dfs = func(nodeID int) int {
		children := hierarchy[nodeID]
		if len(children) == 0 {
			return 1
		}

		maxChildDepth := 0
		for _, child := range children {
			childDepth := dfs(child)
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}

		return maxChildDepth + 1
	}

	for _, manager := range managers {
		depth := dfs(manager)
		if depth > maxDepth {
			maxDepth = depth
			deepestNode = manager
		}
	}

	return deepestNode, maxDepth
}
