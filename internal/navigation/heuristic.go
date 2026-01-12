package navigation

import (
	"math"
	"waze/internal/graph"
)

func heuristic(n1, n2 *graph.Node) float64 {
	return math.Sqrt((n1.X-n2.X)*(n1.X-n2.X) + (n1.Y-n2.Y)*(n1.Y-n2.Y))
}
