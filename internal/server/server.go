package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"waze/internal/graph"
	"waze/internal/types"
)
type Server struct {
	Graph *graph.Graph
}

func NewServer(mapFile string) *Server {
	g, err := graph.LoadGraph(mapFile)
	if err != nil {
		log.Fatal(err)
	}
	return &Server{Graph: g}
}

// func (s *Server) HandleTrafficBatch(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var reports []types.TrafficReport
// 	if err := json.NewDecoder(r.Body).Decode(&reports); err != nil {
// 		http.Error(w, "Invalid Json", http.StatusBadRequest)
// 		return
// 	}

// 	// NOT PARALLEL YET !!!!
// 	count := 0
// 	for _, report := range reports {
// 		if edge, exists := s.Graph.Edges[report.EdgeID]; exists && report.CarID != 1 {
// 			edge.UpdateSpeed(report.Speed)
// 			count++
// 		}
// 	}

// 	// fmt.Printf("Processed batch of %d reports\n", count)
// 	w.WriteHeader(http.StatusOK)
// }


func (s *Server) HandleTrafficBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
		return
	}

	var reports []types.TrafficReport
	if err := json.NewDecoder(r.Body).Decode(&reports); err != nil {
		http.Error(w, "Invalid Json", http.StatusBadRequest)
		return
	}

	// מקביליות בעדכון
	numWorkers := 8
	reportsCount := len(reports)
	
	if reportsCount == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	if reportsCount < numWorkers {
		numWorkers = 1
	}

	chunkSize := (reportsCount + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := min(start+chunkSize, reportsCount)

		go func(startIdx, endIdx int) {
			defer wg.Done()
			for j := startIdx; j < endIdx; j++ {
				report := reports[j]
				if edge, exists := s.Graph.Edges[report.EdgeID]; exists && report.CarID != -1 {
					edge.UpdateSpeed(report.Speed)
				}
			}
		}(start, end)
	}

	wg.Wait()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleNavigation(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	// fmt.Printf("From str is: %s, To str is: %s\n", fromStr, toStr)

	fromId, err1 := strconv.Atoi(fromStr)
	toId, err2 := strconv.Atoi(toStr)
	// fmt.Printf("From is: %d, To is: %d\n", fromId, toId)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid 'from' or 'to' parameters", http.StatusBadRequest)
		return
	}
	req := PathRequest{
		StartNodeId:     fromId,
		EndNodeId:       toId,
		ResponseChannel: make(chan PathResult),
	}

	JobQueue <- req

	result := <-req.ResponseChannel

	// if no route was found
	if result.Err != nil {
		fmt.Printf("The Error is: %s\n", result.Err.Error())
		http.Error(w, result.Err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Response)
}
