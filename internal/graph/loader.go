package graph

import (
	"encoding/json"
	"fmt"
	"os"
)

type container struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func LoadGraph(fileName string) (*Graph, error) {
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file %w", err)
	}
	var container container

	if err := json.Unmarshal(fileData, &container); err != nil {
		return nil, fmt.Errorf("Failed to parse JSON: %w", err)
	}

	g := NewGraph()

	// add all nodes to the graph
	for i := range container.Nodes {
		g.AddNode(&container.Nodes[i])
	}

	// add all edges to the graph
	for i := range container.Edges {
		e := &container.Edges[i]

		// init current speed to the speed limit
		e.SetCurrentSpeed(e.SpeedLimit)

		// add the edge to the graph
		if err := g.AddEdge(e); err != nil {
			fmt.Printf("Warning: Skipping edge %d: %v\n", e.Id, err)
		}
	}
	return g, nil
}
