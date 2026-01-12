package navigation

import (
	"container/heap"
	"fmt"
	"slices"
	"waze/internal/graph"
)

const V_REF float64 = 70

func FindPathAstar(g *graph.Graph, srcId, dstId int) (*PathResult, error) {
	srcNode, ok1 := g.Nodes[srcId]
	dstNode, ok2 := g.Nodes[dstId]

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("one of the nodes does not exist inside the graph")
	}

	pq := newPriorityQueue()
	heap.Init(pq)

	gScore := make(map[int]float64)
	gScore[srcId] = 0

	cameFrom := make(map[int]int)
	closed := make(map[int]bool)

	heap.Push(pq, &AstarNode{
		NodeId:   srcId,
		Gscore:   0,
		Priority: heuristic(srcNode, dstNode) / V_REF,
	})

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*AstarNode)
		u := current.NodeId

		// we reached the dst node
		if u == dstId {
			route := reconstructRoute(cameFrom, u, g)
			distance := calcDist(g, route)

			return &PathResult{
				Route:    route,
				ETA:      current.Gscore * 60, // convert to minutes
				Distance: distance,
			}, nil
		}

		// if the node is in the closed set - continue to the next node in the heap
		if closed[u] {
			continue
		}
		// else put the node in the closed set
		closed[u] = true

		for _, edge := range g.GetNeighbors(u) {
			v := edge.To
			// if the node is in the closed set - continue to the next neighbor
			if closed[v] {
				continue
			}
			speed := edge.GetCurrentSpeed()
			// safety check
			if speed <= 0 {
				speed = 1.0
			}

			timeCost := edge.Length / speed
			newGscore := gScore[u] + timeCost

			oldScore, exists := gScore[v]
			if !exists || newGscore < oldScore {
				gScore[v] = newGscore

				h := heuristic(g.Nodes[v], dstNode) / V_REF
				f := newGscore + h

				// if v already in pq
				if _, exists := pq.index[v]; exists {
					pq.Update(v, f, newGscore)
				} else {
					heap.Push(pq, &AstarNode{
						NodeId:   v,
						Gscore:   newGscore,
						Priority: f,
					})
				}
				cameFrom[v] = u
			}
		}
	}

	// no path was found
	return nil, fmt.Errorf("No path found between %d and %d", srcId, dstId)
}

// return the total distance of the route in KM
func calcDist(g *graph.Graph, route []int) float64 {
	total_distance := 0.0

	for i := 0; i < len(route)-1; i++ {
		u := route[i]
		v := route[i+1]

		for _, edge := range g.GetNeighbors(u) {
			if edge.To == v {
				total_distance += edge.Length
				break
			}
		}
	}
	return total_distance
}

// func reconstructRoute(cameFrom map[int]int, current int) []int {
// 	path := make([]int, 0, len(cameFrom))
// 	for {
// 		path = append(path, current)
// 		prev, ok := cameFrom[current]
// 		if !ok {
// 			break
// 		}
// 		current = prev
// 	}
// 	slices.Reverse(path)
// 	return path
// }

func reconstructRoute(cameFrom map[int]int, current int, g *graph.Graph) []int {
	path := make([]int, 0)

	for {
		prev, ok := cameFrom[current]

		// prev doesn't exists. That means current is the src node
		if !ok {
			break
		}
		// The edge that connects prev and current
		edgeId := -1

		for _, edge := range g.GetNeighbors(prev) {
			if edge.To == current {
				edgeId = edge.Id
				break
			}
		}

		// we find the edge
		if edgeId != -1 {
			path = append(path, edgeId)
		}
		current = prev
	}
	// return the current path (in reverse)
	slices.Reverse(path)
	return path
}
