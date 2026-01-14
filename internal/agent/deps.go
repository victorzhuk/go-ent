package agent

// Node represents an agent node in the dependency graph.
type Node struct {
	Name string
	Meta AgentMeta
}

// Edge represents a dependency relationship.
// From depends on To.
type Edge struct {
	From string
	To   string
}

// DependencyGraph represents the agent dependency graph.
type DependencyGraph struct {
	Nodes         map[string]Node
	AdjacencyList map[string][]string
}

// NewDependencyGraph creates a new dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes:         make(map[string]Node),
		AdjacencyList: make(map[string][]string),
	}
}

// AddNode adds a node to the graph.
func (g *DependencyGraph) AddNode(name string, meta AgentMeta) {
	g.Nodes[name] = Node{
		Name: name,
		Meta: meta,
	}
	if _, exists := g.AdjacencyList[name]; !exists {
		g.AdjacencyList[name] = []string{}
	}
}

// AddEdge adds a dependency edge to the graph.
// From depends on To.
func (g *DependencyGraph) AddEdge(from, to string) {
	if _, exists := g.AdjacencyList[from]; !exists {
		g.AdjacencyList[from] = []string{}
	}
	g.AdjacencyList[from] = append(g.AdjacencyList[from], to)
}

// GetAdjacencyList returns the adjacency list for a node.
func (g *DependencyGraph) GetAdjacencyList(name string) []string {
	deps, exists := g.AdjacencyList[name]
	if !exists {
		return []string{}
	}
	return deps
}

// HasDependency checks if a dependency relationship exists.
func (g *DependencyGraph) HasDependency(from, to string) bool {
	deps, exists := g.AdjacencyList[from]
	if !exists {
		return false
	}
	for _, dep := range deps {
		if dep == to {
			return true
		}
	}
	return false
}

// HasNode checks if a node exists in the graph.
func (g *DependencyGraph) HasNode(name string) bool {
	_, exists := g.Nodes[name]
	return exists
}

// GetNode returns a node by name.
func (g *DependencyGraph) GetNode(name string) (Node, bool) {
	node, exists := g.Nodes[name]
	return node, exists
}
