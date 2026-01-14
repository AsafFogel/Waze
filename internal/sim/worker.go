package sim

import (
	"sync"
	"waze/internal/graph"
)

type MoveJob struct {
	Car        *Car
	DeltaTime  float64
	Graph      *graph.Graph
	DensityMap map[int]int
}

var moveJobQueue chan MoveJob
var moveWg sync.WaitGroup

func StartMoveWorkers(numWorkers int) {
	moveJobQueue = make(chan MoveJob, 1000)
	for i := 0; i < numWorkers; i++ {
		go moveWorker()
	}
}

func moveWorker() {
	for job := range moveJobQueue {
		select {
		case newRoute := <-job.Car.NewRouteChan:
			job.Car.InitRoute(newRoute, job.Graph)
		default:
		}
		job.Car.Move(job.DeltaTime, job.Graph, job.DensityMap)
		moveWg.Done()
	}
}

func MoveCarsParallel(cars []*Car, dt float64, g *graph.Graph, density map[int]int) {
	moveWg.Add(len(cars))
	for _, car := range cars {
		moveJobQueue <- MoveJob{
			Car:        car,
			DeltaTime:  dt,
			Graph:      g,
			DensityMap: density,
		}
	}
	moveWg.Wait()
}