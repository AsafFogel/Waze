package server

import (
	"log"
	"waze/internal/graph"
	"waze/internal/navigation"
)

var JobQueue chan PathRequest

func WakeWorkers(numWorkers int, g *graph.Graph) {
	JobQueue = make(chan PathRequest, 100)
	for i := 0; i < numWorkers; i++ {
		go worker(g)
	}
	log.Printf("JobQueue started with %d workers", numWorkers)
}

func worker(g *graph.Graph) {
	for req := range JobQueue {
		pathRes, err := navigation.FindPathAstar(g, req.StartNodeId, req.EndNodeId)

		result := PathResult{}
		if err != nil {
			result.Err = err
		} else {
			result.Response.RouteNodes = pathRes.Route
			result.Response.ETA = pathRes.ETA
			result.Response.Distance = pathRes.Distance
		}
		req.ResponseChannel <- result
	}
}