package agent

import (
	"fmt"
	"strings"
)

// Validator validates the dependency graph.
type Validator struct {
	graph *DependencyGraph
}

// NewValidator creates a new validator.
func NewValidator(graph *DependencyGraph) *Validator {
	return &Validator{
		graph: graph,
	}
}

// ValidateDependencies checks the dependency graph for:
// - All referenced agents exist
// - No cycles in dependencies
func (v *Validator) ValidateDependencies() error {
	if err := v.validateAgentReferences(); err != nil {
		return err
	}

	if err := v.validateNoCycles(); err != nil {
		return err
	}

	return nil
}

// validateAgentReferences checks that all dependencies reference existing agents.
func (v *Validator) validateAgentReferences() error {
	for _, node := range v.graph.Nodes {
		for _, dep := range node.Meta.Dependencies {
			if !v.graph.HasNode(dep) {
				return fmt.Errorf("agent not found: %s", dep)
			}
		}
	}
	return nil
}

// validateNoCycles checks for cycles in the dependency graph using DFS.
func (v *Validator) validateNoCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for node := range v.graph.Nodes {
		if !visited[node] {
			path, err := v.detectCycle(node, visited, recStack)
			if err != nil {
				return err
			}
			if path != nil {
				return fmt.Errorf("cycle detected in dependency graph: %s", strings.Join(path, " -> "))
			}
		}
	}

	return nil
}

// detectCycle performs DFS to detect cycles in the dependency graph.
// Returns the cycle path if found, otherwise nil and no error.
func (v *Validator) detectCycle(node string, visited, recStack map[string]bool) ([]string, error) {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range v.graph.GetAdjacencyList(node) {
		if !visited[neighbor] {
			path, err := v.detectCycle(neighbor, visited, recStack)
			if err != nil {
				return nil, err
			}
			if path != nil {
				return append([]string{node}, path...), nil
			}
		} else if recStack[neighbor] {
			return []string{node, neighbor}, nil
		}
	}

	recStack[node] = false
	return nil, nil
}
