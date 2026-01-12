package server

import (
	"log"
	"waze/internal/graph"
	"waze/internal/navigation"
)

var JobQueue chan PathRequest

func WakeWorkers(numWorkers int, g *graph.Graph) {
	// set the limit of buffer of the route requests to 100
	JobQueue = make(chan PathRequest, 100)
	for i := 0; i < numWorkers; i++ {
		// wake worker to take jobs from the work queue
		go worker( /*i, */ g)
	}
	log.Printf("JobQueue started with %d workers", numWorkers)
}

func worker( /*id int,*/ g *graph.Graph) {
	for req := range JobQueue {
		// fmt.Printf("worker %d is calculating route frod nodeID %d to nodeId %d\n", id, req.StartNodeId, req.EndNodeId)
		pathRes, err := navigation.FindPathAstar(g, req.StartNodeId, req.EndNodeId)

		result := PathResult{}
		if err != nil {
			result.Err = err
		} else {
			result.Response.RouteNodes = pathRes.Route
			result.Response.ETA = pathRes.ETA
			result.Response.Distance = pathRes.Distance
			result.Response.Err = err
		}
		req.ResponseChannel <- result
	}
}
