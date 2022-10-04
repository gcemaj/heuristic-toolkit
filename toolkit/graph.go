package toolkit

type Node struct {
	Id       int
	Metadata map[string]interface{}
}

type Edge struct {
	Phermone float64
	Distance int
}

type Graph struct {
	Nodes map[int]Node
	Edges map[int]map[int]*Edge
}

func NewGraph() Graph {
	return Graph{
		Nodes: map[int]Node{},
		Edges: map[int]map[int]*Edge{},
	}
}

func (g *Graph) AddNode(node Node) {
	g.Nodes[node.Id] = node
	g.Edges[node.Id] = map[int]*Edge{}
}

func (g *Graph) UpsertEdge(from int, to int, edge *Edge) {
	if _, ok := g.Nodes[from]; !ok {
		g.AddNode(Node{Id: from})
	}
	if _, ok := g.Nodes[to]; !ok {
		g.AddNode(Node{Id: to})
	}

	g.Edges[from][to] = edge
}
