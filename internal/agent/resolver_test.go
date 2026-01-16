package agent

//nolint:gosec // test file with necessary file operations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveDependencies_SimpleChain(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestResolveDependencies_DiamondDependency(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("b", "d")
	graph.AddEdge("c", "d")

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.NoError(t, err)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
	assert.Contains(t, result, "d")
	assert.Equal(t, []string{"a", "b", "c", "d"}, result)
}

func TestResolveDependencies_MissingAgent(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddEdge("a", "b")

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "agent not found")
	assert.Contains(t, err.Error(), "b")
}

func TestResolveDependencies_EmptyList(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{})

	assert.NoError(t, err)
	assert.Equal(t, []string{}, result)
}

func TestResolveDependencies_SingleAgent(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"a"}, result)
}

func TestResolveDependencies_MultipleIndependentAgents(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a", "b", "c"})

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
}

func TestResolveDependencies_MultipleLevelsDeep(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddNode("e", AgentMeta{Name: "e"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "d")
	graph.AddEdge("d", "e")

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, result)
}

func TestResolveDependencies_TransitiveDeps(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("b", "d")

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.NoError(t, err)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
	assert.Contains(t, result, "d")
}

func TestResolveDependencies_MissingRootAgent(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"missing"})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "agent not found")
	assert.Contains(t, err.Error(), "missing")
}

func TestResolveDependencies_MixedDependencies(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddNode("e", AgentMeta{Name: "e"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("b", "d")
	graph.AddEdge("c", "e")

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.NoError(t, err)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
	assert.Contains(t, result, "d")
	assert.Contains(t, result, "e")
	assert.Equal(t, 5, len(result))
}

func TestResolveDependencies_PartialGraph(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddNode("d", AgentMeta{Name: "d", Dependencies: []string{"e"}})
	graph.AddNode("e", AgentMeta{Name: "e"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("d", "e")

	resolver := NewResolver(graph)
	result, err := resolver.ResolveDependencies([]string{"a"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestTopologicalSort_SimpleChain(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	resolver := NewResolver(graph)
	result, err := resolver.TopologicalSort()

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestTopologicalSort_MultipleIndependent(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})

	resolver := NewResolver(graph)
	result, err := resolver.TopologicalSort()

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
}

func TestTopologicalSort_Diamond(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("b", "d")
	graph.AddEdge("c", "d")

	resolver := NewResolver(graph)
	result, err := resolver.TopologicalSort()

	assert.NoError(t, err)
	assert.Len(t, result, 4)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
	assert.Contains(t, result, "d")
}

func TestTopologicalSort_Cycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "a")

	resolver := NewResolver(graph)
	result, err := resolver.TopologicalSort()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestTopologicalSort_SelfCycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddEdge("a", "a")

	resolver := NewResolver(graph)
	result, err := resolver.TopologicalSort()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestTopologicalOrdering_DependenciesBeforeDependents(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	resolver := NewResolver(graph)
	result, err := resolver.TopologicalSort()

	assert.NoError(t, err)

	idxA, idxB, idxC := -1, -1, -1
	for i, name := range result {
		switch name {
		case "a":
			idxA = i
		case "b":
			idxB = i
		case "c":
			idxC = i
		}
	}

	assert.Less(t, idxA, idxB, "a must come before b (a depends on b)")
	assert.Less(t, idxB, idxC, "b must come before c (b depends on c)")
}

func TestTopologicalOrdering_ComplexGraph(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddNode("e", AgentMeta{Name: "e"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("b", "d")
	graph.AddEdge("c", "d")

	resolver := NewResolver(graph)
	result, err := resolver.TopologicalSort()

	assert.NoError(t, err)
	assert.Len(t, result, 5)

	pos := make(map[string]int)
	for i, name := range result {
		pos[name] = i
	}

	assert.Less(t, pos["a"], pos["b"], "a must come before b (a depends on b)")
	assert.Less(t, pos["a"], pos["c"], "a must come before c (a depends on c)")
	assert.Less(t, pos["b"], pos["d"], "b must come before d (b depends on d)")
	assert.Less(t, pos["c"], pos["d"], "c must come before d (c depends on d)")
}
