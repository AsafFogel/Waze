package sim

import (
	"sync"
	"waze/internal/graph"
	"waze/internal/types"
)

type MoveChunkJob struct {
	Cars       []*Car
	DeltaTime  float64
	Graph      *graph.Graph
	DensityMap map[int]int
	Client     *Client
	Timestamp  int64
}

var chunkJobQueue chan MoveChunkJob
var chunkWg sync.WaitGroup

func StartChunkWorkers(numWorkers int) {
	chunkJobQueue = make(chan MoveChunkJob, 100)
	for i := 0; i < numWorkers; i++ {
		go chunkWorker()
	}
}

func chunkWorker() {
	for job := range chunkJobQueue {
		reports := make([]types.TrafficReport, 0, len(job.Cars))

		for _, car := range job.Cars {
			select {
			case newRoute := <-car.NewRouteChan:
				car.InitRoute(newRoute, job.Graph)
			default:
			}
			car.Move(job.DeltaTime, job.Graph, job.DensityMap)

			if car.State == Driving && car.ActiveRoute != nil {
				reports = append(reports, types.TrafficReport{
					CarID:     car.Id,
					EdgeID:    car.ActiveRoute.RouteEdges[car.ActiveRoute.CurrentEdgeIndex],
					Speed:     car.CurrentSpeed,
					Timestamp: job.Timestamp,
				})
			}
		}

		if len(reports) > 0 {
			job.Client.SendTrafficBatch(reports)
		}

		chunkWg.Done()
	}
}

func MoveCarsChunked(cars []*Car, dt float64, g *graph.Graph, density map[int]int, client *Client, timestamp int64) {
	if len(cars) == 0 {
		return
	}

	chunkSize := 100
	numChunks := (len(cars) + chunkSize - 1) / chunkSize

	chunkWg.Add(numChunks)

	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := min(start+chunkSize, len(cars))

		chunkJobQueue <- MoveChunkJob{
			Cars:       cars[start:end],
			DeltaTime:  dt,
			Graph:      g,
			DensityMap: density,
			Client:     client,
			Timestamp:  timestamp,
		}
	}

	chunkWg.Wait()
}