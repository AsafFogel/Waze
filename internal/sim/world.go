package sim

import (
	"fmt"
	"log"
	"sync"
	"time"
	"waze/internal/config"
	"waze/internal/graph"
	"waze/internal/types"
)

const TIME_TO_REPORT = 2

type World struct {
	Graph         *graph.Graph
	Cars          []*Car
	SimTime       float64 // time from the start of the sim
	ReportsBuffer []types.TrafficReport
	Client        *Client

	VirtualStartTime time.Time

	EdgeDensity map[int]int // edgeId->number of cars in the edge
}

func NewWorld(mapFile, serverUrl string) (*World, error) {
	// load graph
	g, err := graph.LoadGraph(mapFile)
	if err != nil {
		log.Fatal(err)
	}
	return &World{
		Graph:            g,
		Cars:             make([]*Car, 0), // slice of all cars in the map
		SimTime:          0,
		VirtualStartTime: time.Now(),
		Client:           NewClient(serverUrl),
	}, nil
}

func (w *World) GetCurrentTime() int64 {
	currentTime := w.VirtualStartTime.Add(time.Duration(w.SimTime) * time.Second)
	return currentTime.Unix()
}

func (w *World) AddCar(id, userId int) *Car {
	car := NewCar(id, userId, w.SimTime)
	w.Cars = append(w.Cars, car)
	return car
}

func (w *World) HasActiveCars() bool {
	return len(w.Cars) > 0
}

func (w *World) CleanArrivedCars() {
	activeCars := w.Cars[:0]

	for _, car := range w.Cars {
		if car.State != Arrived {
			activeCars = append(activeCars, car)
		}
	}
	w.Cars = activeCars
}

func (w *World) GenarateTrafficReportsParallel() []types.TrafficReport {
	// defer config.TimeTrack(time.Now(), "GenarateTrafficReportsParallel")
	carsCount := len(w.Cars)
	if carsCount == 0 {
		return nil
	}

	// optimize the reports slice size
	if cap(w.ReportsBuffer) < carsCount {
		w.ReportsBuffer = make([]types.TrafficReport, carsCount)
	} else {
		w.ReportsBuffer = w.ReportsBuffer[:carsCount]
	}

	// set number of workers to the number of cores in the machine
	numWorkers := 6

	if carsCount < numWorkers {
		numWorkers = 1
	}

	chunkSize := (carsCount + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := min(start+chunkSize, carsCount)
		go func(startIdx, endIdx int) {
			// signal barrier we are done at the end of the function
			defer wg.Done()

			for j := startIdx; j < endIdx; j++ {
				car := w.Cars[j]

				if car.State == Driving && car.ActiveRoute != nil {
					// id of the current edge
					currentEdge := car.ActiveRoute.RouteEdges[car.ActiveRoute.CurrentEdgeIndex]

					// write to the slice of reports
					w.ReportsBuffer[j] = types.TrafficReport{
						CarID:     car.Id,
						EdgeID:    currentEdge,
						Speed:     car.CurrentSpeed,
						Timestamp: w.GetCurrentTime(),
					}
				} else {
					// The car is not in driving right now
					w.ReportsBuffer[j].CarID = -1
				}
			}
		}(start, end)

	}
	// Wait for all
	wg.Wait()
	return w.ReportsBuffer
}

func (w *World) GenarateTrafficReports() []types.TrafficReport {
	// defer config.TimeTrack(time.Now(), "GenarateTrafficReports")
	carsCount := len(w.Cars)
	if carsCount == 0 {
		return nil
	}

	// optimize the reports slice size
	if cap(w.ReportsBuffer) < carsCount {
		w.ReportsBuffer = make([]types.TrafficReport, carsCount)
	} else {
		w.ReportsBuffer = w.ReportsBuffer[:carsCount]
	}
	for i := range carsCount {
		car := w.Cars[i]
		if car.State == Driving && car.ActiveRoute != nil {
			// id of the current edge
			currentEdge := car.ActiveRoute.RouteEdges[car.ActiveRoute.CurrentEdgeIndex]

			// write to the slice of reports
			w.ReportsBuffer[i] = types.TrafficReport{
				CarID:     car.Id,
				EdgeID:    currentEdge,
				Speed:     car.CurrentSpeed,
				Timestamp: w.GetCurrentTime(),
			}
		} else {
			// The car is not in driving right now
			w.ReportsBuffer[i].CarID = -1
		}
	}
	return w.ReportsBuffer
}

func (w *World) Tick(dt float64) {
	w.SimTime += dt

	// update edge density
	w.EdgeDensity = make(map[int]int)
	for _, car := range w.Cars {
		if car.State == Driving && car.ActiveRoute != nil {
			edgeId := car.ActiveRoute.RouteEdges[car.ActiveRoute.CurrentEdgeIndex]
			w.EdgeDensity[edgeId]++
		}
	}

	// timeString := time.Unix(w.GetCurrentTime(), 0).Format("15:04:05")

	// move all cars
	for _, car := range w.Cars {

		select {
		case newRoute := <-car.NewRouteChan:
			// jammedEdge := 150

			// wasInJam := contains(car.ActiveRoute.RouteEdges, jammedEdge)
			// isInJam := contains(newRoute, jammedEdge)

			// if wasInJam && !isInJam {
			// 	fmt.Printf("\nSUCCESS! Car %d detected the jam on Edge %d and rerouted!\n", car.Id, jammedEdge)
			// 	fmt.Printf("Old Route Length: %d\n", len(car.ActiveRoute.RouteEdges))
			// 	fmt.Printf("New Route Length: %d (Longer distance, shorter time!)\n\n", len(newRoute))
			// }

			car.InitRoute(newRoute, w.Graph)
		default:

		}

		car.Move(dt, w.Graph, w.EdgeDensity)

		// check a minute passed from the last route request and the car isn't in the last roud
		if car.LastRouteReq > 60 && car.State == Driving && car.ActiveRoute != nil {
			// set Last route time to zero
			car.LastRouteReq = 0

			currentIdx := car.ActiveRoute.CurrentEdgeIndex
			routeLen := len(car.ActiveRoute.RouteEdges)

			// check if we are in the last edge in the road. no point asking a new route
			if currentIdx >= routeLen-1 {
				continue
			}

			// get current node and destination node id
			currentEdgeId := car.ActiveRoute.RouteEdges[currentIdx]
			dstEdgeId := car.ActiveRoute.RouteEdges[routeLen-1]

			// get current node and destination node
			nextNode := w.Graph.Edges[currentEdgeId].To
			dstNode := w.Graph.Edges[dstEdgeId].To

			// copy of the current route
			oldRoute := make([]int, len(car.ActiveRoute.RouteEdges))
			copy(oldRoute, car.ActiveRoute.RouteEdges)

			go func(car *Car, src, dst int, previousRoute []int, idx int) {
				newRoute, err := w.Client.RequestRoute(nextNode, dstNode)
				if err != nil {
					fmt.Printf("❌ Car %d reroute failed: %v\n", car.Id, err)
					return
				}
				// check if the route is different
				if different(newRoute, oldRoute, car.Id) {
					car.NewRouteChan <- newRoute
				}
			}(car, nextNode, dstNode, oldRoute, currentIdx)
		}
	}

	// every fixed number of second we send a speed report
	if int(w.SimTime)%int(config.Global.Simulation.ReportInterval) == 0 {
		reports := w.GenarateTrafficReports()
		// send the traffic report to server
		go func(batch []types.TrafficReport) {
			err := w.Client.SendTrafficBatch(batch)
			if err != nil {
				fmt.Println("Failed to send traffic batch: ", err)
			}
		}(reports)
	}
}

func (w *World) TickParallel(dt float64) {
	w.SimTime += dt

	var wg sync.WaitGroup
	numCars := len(w.Cars)

	numWorkers := 12
	chunkSize := (numCars + numWorkers - 1) / numWorkers

	// timeString := time.Unix(w.GetCurrentTime(), 0).Format("15:04:05")

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				if j < numCars {
					// car := w.Cars[j]
					// car.Move(dt, w.Graph)
					// if car.ActiveRoute != nil && car.State == Driving {
					// 	if car.Id == 1 {
					// 		fmt.Printf("[%s] Car 1  | Edge Idx: %d | Progress: %.3f / %.3f km\n | Speed: %.1f\n",
					// 			timeString,
					// 			car.ActiveRoute.CurrentEdgeIndex,
					// 			car.ActiveRoute.EdgeProgress,
					// 			car.ActiveRoute.CurrentEdgeLen,
					// 			car.CurrentSpeed)
					// 	}
					// }
				}
			}
		}(i*chunkSize, (i+1)*chunkSize)
	}
	wg.Wait()
}

func different(newRoute, currentRoute []int, currentIndex int) bool {
	// compare length
	if len(newRoute) != len(currentRoute)-currentIndex {
		return true
	}

	// the lengths are equal - compare each edge
	for i := 0; i < len(newRoute); i++ {
		if newRoute[i] != currentRoute[i+currentIndex] {
			return true
		}
	}
	return false
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
