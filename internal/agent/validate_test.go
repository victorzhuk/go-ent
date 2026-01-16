package agent

//nolint:gosec // test file with necessary file operations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDependencies_AllValid(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b", "c"}})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.NoError(t, err)
}

func TestValidateDependencies_MissingDependency(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddEdge("a", "b")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
	assert.Contains(t, err.Error(), "b")
}

func TestValidateDependencies_MultipleMissingDependencies(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b", "c", "d"}})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddEdge("a", "b")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
	assert.Contains(t, err.Error(), "c")
}

func TestValidateDependencies_CycleDetection_Simple(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"a"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
	assert.Contains(t, err.Error(), "->")
	assert.Contains(t, err.Error(), "a")
	assert.Contains(t, err.Error(), "b")
}

func TestValidateDependencies_CycleDetection_Complex(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c", Dependencies: []string{"a"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "a")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
	assert.Contains(t, err.Error(), "a -> b -> c -> a")
}

func TestValidateDependencies_CycleDetection_DiamondToCycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "b")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestValidateDependencies_CycleDetection_SelfDependency(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"a"}})
	graph.AddEdge("a", "a")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestValidateDependencies_DAG_NoCycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b", "c"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"d"}})
	graph.AddNode("c", AgentMeta{Name: "c", Dependencies: []string{"d"}})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("b", "d")
	graph.AddEdge("c", "d")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.NoError(t, err)
}

func TestValidateDependencies_EmptyGraph(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.NoError(t, err)
}

func TestValidateDependencies_SingleNode(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.NoError(t, err)
}

func TestValidateDependencies_MultipleIndependentNodes(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.NoError(t, err)
}

func TestValidateDependencies_MissingDepInChain(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
	assert.Contains(t, err.Error(), "c")
}

func TestValidateDependencies_TransitiveValidation(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.NoError(t, err)
}

func TestValidateDependencies_ComplexDAG(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b", "c", "d"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"e"}})
	graph.AddNode("c", AgentMeta{Name: "c", Dependencies: []string{"e", "f"}})
	graph.AddNode("d", AgentMeta{Name: "d"})
	graph.AddNode("e", AgentMeta{Name: "e"})
	graph.AddNode("f", AgentMeta{Name: "f"})
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("a", "d")
	graph.AddEdge("b", "e")
	graph.AddEdge("c", "e")
	graph.AddEdge("c", "f")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.NoError(t, err)
}

func TestValidateDependencies_CycleErrorMessageIncludesPath(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c", Dependencies: []string{"a"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "a")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
	assert.Contains(t, err.Error(), "->")
	assert.Contains(t, err.Error(), "a")
	assert.Contains(t, err.Error(), "b")
	assert.Contains(t, err.Error(), "c")
}

func TestValidateAgentReferences_AllDependenciesExist(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b", "c"}})
	graph.AddNode("b", AgentMeta{Name: "b"})
	graph.AddNode("c", AgentMeta{Name: "c"})

	validator := NewValidator(graph)
	err := validator.validateAgentReferences()

	assert.NoError(t, err)
}

func TestValidateAgentReferences_MissingDependency(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})

	validator := NewValidator(graph)
	err := validator.validateAgentReferences()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
	assert.Contains(t, err.Error(), "b")
}

func TestValidateAgentReferences_NoDependencies(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a"})

	validator := NewValidator(graph)
	err := validator.validateAgentReferences()

	assert.NoError(t, err)
}

func TestValidateAgentReferences_EmptyDependencyList(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{}})

	validator := NewValidator(graph)
	err := validator.validateAgentReferences()

	assert.NoError(t, err)
}

func TestValidateAgentReferences_MultipleNodesWithDeps(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b", "c"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c"})

	validator := NewValidator(graph)
	err := validator.validateAgentReferences()

	assert.NoError(t, err)
}

func TestDetectCycle_SimpleCycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"a"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")

	validator := NewValidator(graph)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	path, err := validator.detectCycle("a", visited, recStack)

	assert.NoError(t, err)
	assert.NotNil(t, path)
	assert.Contains(t, path, "a")
	assert.Contains(t, path, "b")
}

func TestDetectCycle_NoCycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	validator := NewValidator(graph)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	path, err := validator.detectCycle("a", visited, recStack)

	assert.NoError(t, err)
	assert.Nil(t, path)
}

func TestDetectCycle_ComplexCycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c", Dependencies: []string{"d"}})
	graph.AddNode("d", AgentMeta{Name: "d", Dependencies: []string{"a"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "d")
	graph.AddEdge("d", "a")

	validator := NewValidator(graph)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	path, err := validator.detectCycle("a", visited, recStack)

	assert.NoError(t, err)
	assert.NotNil(t, path)
	assert.Contains(t, path, "a")
	assert.Contains(t, path, "b")
}

func TestValidateNoCycles_DAG(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c"})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	validator := NewValidator(graph)
	err := validator.validateNoCycles()

	assert.NoError(t, err)
}

func TestValidateNoCycles_WithCycle(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"a"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")

	validator := NewValidator(graph)
	err := validator.validateNoCycles()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestValidateDependencies_OrderOfValidation(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestValidateDependencies_CycleBeforeMissingDep(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"c"}})
	graph.AddNode("c", AgentMeta{Name: "c", Dependencies: []string{"a"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "a")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestValidateDependencies_MultipleCycles(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph()
	graph.AddNode("a", AgentMeta{Name: "a", Dependencies: []string{"b"}})
	graph.AddNode("b", AgentMeta{Name: "b", Dependencies: []string{"a"}})
	graph.AddNode("c", AgentMeta{Name: "c", Dependencies: []string{"d"}})
	graph.AddNode("d", AgentMeta{Name: "d", Dependencies: []string{"c"}})
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")
	graph.AddEdge("c", "d")
	graph.AddEdge("d", "c")

	validator := NewValidator(graph)
	err := validator.ValidateDependencies()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}
