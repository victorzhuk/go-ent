package agent

import (
	"fmt"
)

// Resolver resolves agent dependencies.
type Resolver struct {
	graph *DependencyGraph
}

// NewResolver creates a new dependency resolver.
func NewResolver(graph *DependencyGraph) *Resolver {
	return &Resolver{
		graph: graph,
	}
}

// ResolveDependencies resolves dependencies for given agent names.
// Returns a topologically sorted list where dependencies come before dependents.
func (r *Resolver) ResolveDependencies(agentNames []string) ([]string, error) {
	if len(agentNames) == 0 {
		return []string{}, nil
	}

	subset := make(map[string]struct{})
	queue := make([]string, 0, len(agentNames))

	for _, name := range agentNames {
		if !r.graph.HasNode(name) {
			return nil, fmt.Errorf("agent not found: %s", name)
		}
		subset[name] = struct{}{}
		queue = append(queue, name)
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dep := range r.graph.GetAdjacencyList(current) {
			if _, visited := subset[dep]; !visited {
				if !r.graph.HasNode(dep) {
					return nil, fmt.Errorf("agent not found: %s", dep)
				}
				subset[dep] = struct{}{}
				queue = append(queue, dep)
			}
		}
	}

	result, err := r.topologicalSortSubset(subset)
	if err != nil {
		return nil, fmt.Errorf("topological sort: %w", err)
	}

	return result, nil
}

// TopologicalSort returns all nodes in topological order.
func (r *Resolver) TopologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for name := range r.graph.Nodes {
		inDegree[name] = 0
	}

	for _, deps := range r.graph.AdjacencyList {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	queue := make([]string, 0)
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]string, 0, len(r.graph.Nodes))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		result = append(result, current)

		for _, neighbor := range r.graph.AdjacencyList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(r.graph.Nodes) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return result, nil
}

// topologicalSortSubset performs topological sort on a subset of nodes.
func (r *Resolver) topologicalSortSubset(subset map[string]struct{}) ([]string, error) {
	inDegree := make(map[string]int)
	for name := range subset {
		inDegree[name] = 0
	}

	for from := range subset {
		for _, dep := range r.graph.GetAdjacencyList(from) {
			if _, inSubset := subset[dep]; inSubset {
				inDegree[dep]++
			}
		}
	}

	queue := make([]string, 0)
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]string, 0, len(subset))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		result = append(result, current)

		for _, dep := range r.graph.GetAdjacencyList(current) {
			if _, inSubset := subset[dep]; inSubset {
				inDegree[dep]--
				if inDegree[dep] == 0 {
					queue = append(queue, dep)
				}
			}
		}
	}

	if len(result) != len(subset) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return result, nil
}
