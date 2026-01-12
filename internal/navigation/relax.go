package navigation

import (
	"waze/internal/graph"
)

func relaxNeighbors(g *graph.Graph, pq *PriorityQueue, gScore map[int]float64,
	cameFrom map[int]int, closed map[int]bool, u int, goal *graph.Node) {
	for _, edge := range g.GetNeighbors(u) {
		v := edge.To
		if closed[v] {
			continue
		}
	}
}
