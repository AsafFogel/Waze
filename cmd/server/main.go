package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"waze/internal/config"
	"waze/internal/server"
)

func main() {
	// get constants fron config.json
	if err := config.Load("config.json"); err != nil {
		panic(err)
	}
	fmt.Printf("Hello Waze!, map file is: %s\n", config.Global.Server.MapFile)
	srv := server.NewServer(config.Global.Server.MapFile)
	fmt.Printf("The num of cores is: %d\n", runtime.NumCPU())

	server.WakeWorkers(runtime.NumCPU(), srv.Graph)

	http.HandleFunc("/api/traffic", srv.HandleTrafficBatch)
	http.HandleFunc("/api/navigate", srv.HandleNavigation)

	// go func() {
	// 	targetEdge := 150
	// 	for {
	// 		time.Sleep(3 * time.Second)
	// 		if edge, exists := srv.Graph.Edges[targetEdge]; exists {
	// 			speed := edge.GetCurrentSpeed()
	// 			fmt.Printf("SERVER STATUS: Edge %d Speed: %.2f km/h (Should be close to 1.0)\n", targetEdge, speed)
	// 		}
	// 	}
	// }()

	log.Printf("Server running on: %s\n", config.Global.Server.Port)
	log.Fatal(http.ListenAndServe(config.Global.Server.Port, nil))
}

// func solution(g *graph.Graph, startNodeID, endNodeID int) (*navigation.PathResult, error) {
// 	defer config.TimeTrack(time.Now(), "Regular A*")
// 	result, err := navigation.FindPathAstar(g, startNodeID, endNodeID)
// 	if err != nil {
// 		log.Printf("Navigation failed: %v\n", err)
// 		return nil, err
// 	}
// 	return result, nil
// }
