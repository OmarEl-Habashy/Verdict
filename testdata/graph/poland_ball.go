package graph

// CountConnectedComponents counts the number of trees (connected components) in a forest.
// Uses DFS to explore each component.
func CountConnectedComponents(n int, edges [][2]int) int {
	// Build adjacency list
	adjList := make([][]int, n+1)
	for _, edge := range edges {
		u, v := edge[0], edge[1]
		adjList[u] = append(adjList[u], v)
		adjList[v] = append(adjList[v], u) // Undirected graph
	}

	visited := make([]bool, n+1)
	componentCount := 0

	// DFS helper function
	var dfs func(int)
	dfs = func(node int) {
		visited[node] = true
		for _, neighbor := range adjList[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			}
		}
	}

	// Count components
	for i := 1; i <= n; i++ {
		if !visited[i] {
			dfs(i)
			componentCount++
		}
	}

	return componentCount
}

// FindAllComponents returns all connected components as separate slices
func FindAllComponents(n int, edges [][2]int) [][]int {
	adjList := make([][]int, n+1)
	for _, edge := range edges {
		u, v := edge[0], edge[1]
		adjList[u] = append(adjList[u], v)
		adjList[v] = append(adjList[v], u)
	}

	visited := make([]bool, n+1)
	var components [][]int

	var dfs func(int, *[]int)
	dfs = func(node int, component *[]int) {
		visited[node] = true
		*component = append(*component, node)
		for _, neighbor := range adjList[node] {
			if !visited[neighbor] {
				dfs(neighbor, component)
			}
		}
	}

	for i := 1; i <= n; i++ {
		if !visited[i] {
			var component []int
			dfs(i, &component)
			components = append(components, component)
		}
	}

	return components
}
