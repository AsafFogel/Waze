package graph

import (
	"fmt"
	"sort"
	"strings"
)

type Graph struct {
	Nodes map[int]*Node // maps from id of node to struct
	Edges map[int]*Edge // maps from id of edge to struct

	// for random sampling
	NodesArr []int

	// nodeId maps to an array of edges id
	AdjList map[int][]*Edge
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:    make(map[int]*Node),
		Edges:    make(map[int]*Edge),
		NodesArr: make([]int, 0),
		AdjList:  make(map[int][]*Edge),
	}
}

func (g *Graph) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "=== Graph Summary ===\n")
	sb.WriteString(fmt.Sprintf("Nodes: %d | Edges: %d\n", len(g.Nodes), len(g.Edges)))
	sb.WriteString("-------------------\n")

	var keys []int
	for k := range g.Nodes {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, id := range keys {
		node := g.Nodes[id]
		fmt.Fprintf(&sb, "[Node %d] %s (Lat: %.4f, Lon: %.4f)\n", node.Id, node.Name, node.X, node.Y)

		neighbors := g.AdjList[id]
		if len(neighbors) == 0 {
			sb.WriteString("    (No outgoing roads)\n")
		} else {
			for _, edge := range neighbors {
				fmt.Fprintf(&sb, "    --> To Node %d | Speed: %.0f km/h | Len: %.2f km\n",
					edge.To, edge.SpeedLimit, edge.Length)
			}
		}
	}

	return sb.String()
}

func (g *Graph) AddNode(n *Node) {
	g.Nodes[n.Id] = n

	g.NodesArr = append(g.NodesArr, n.Id)
}

func (g *Graph) AddEdge(e *Edge) error {
	// check for existance of source and destination nodes
	if _, ok := g.Nodes[e.From]; !ok {
		return fmt.Errorf("Source node %d not found", e.From)
	}
	if _, ok := g.Nodes[e.To]; !ok {
		return fmt.Errorf("Destination node %d not found", e.To)
	}
	// add edge to the graph
	g.Edges[e.Id] = e

	// add to adjacency list
	g.AdjList[e.From] = append(g.AdjList[e.From], e)

	// return no error
	return nil
}

func (g *Graph) GetNeighbors(nodeId int) []*Edge {
	return g.AdjList[nodeId]
}
