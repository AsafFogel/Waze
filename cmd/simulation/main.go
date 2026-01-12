package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"waze/internal/config"
	"waze/internal/graph"
	"waze/internal/sim"
	"waze/internal/types"
)

var CONFIG_FILE string = "config.json"

func main() {
	// load constants fron config.json
	if err := config.Load(CONFIG_FILE); err != nil {
		panic(err)
	}
	// runtime.GOMAXPROCS(runtime.NumCPU())
	world, err := sim.NewWorld(config.Global.Server.MapFile, (config.Global.Simulation.ServerURL + config.Global.Server.Port))
	if err != nil {
		log.Fatal(err)
	}

	numCars := config.Global.Simulation.NumCars

	// artificial jam
	go func() {
		targetEdgeID := 150
		fakeCarID := 99999

		for {
			jamReport := []types.TrafficReport{
				{
					CarID:     fakeCarID,
					EdgeID:    targetEdgeID,
					Speed:     1.0,
					Timestamp: time.Now().Unix(),
				},
			}
			err := world.Client.SendTrafficBatch(jamReport)
			if err != nil {
				fmt.Printf("Error in jam report %s\n", err)
			}

			// fmt.Printf("🔥 Generated artificial JAM on Edge %d\n", targetEdgeID)
			time.Sleep(2 * time.Second)
		}
	}()

	// var routeSize int
	fmt.Printf("Initializing %d Cars...\n", numCars)
	var route []int
	for i := range numCars {
		// init route for car
		for {
			// sample different src and dst nodes
			src, dst := randomRequest(world.Graph)
			route, err = world.Client.RequestRoute(src, dst)
			if err != nil {
				fmt.Printf("Route does not exists. Error: %s\n", err)
				continue
			}
			break
		}
		// create car and init its route
		car := world.AddCar(i, i)
		car.InitRoute(route, world.Graph)
	}

	// startNode := 0
	// endNode := 99

	// route, err = world.Client.RequestRoute(startNode, endNode)
	// if err == nil {
	// 	fmt.Println("Spawning TEST CAR (ID 777) to cross the map...")
	// 	testCar := world.AddCar(777, 777)
	// 	testCar.InitRoute(route, world.Graph)
	// } else {
	// 	fmt.Println("Failed to spawn test car")
	// }

	dt := 1.0
	loop(numCars, dt, world)

	fmt.Println("Simulation Finished!")
}

func loop(carCounter int, dt float64, world *sim.World) {
	lastLogTime := 0.0
	lastSpawnTime := 0.0

	for {
		if world.SimTime > 10 && !world.HasActiveCars() {
			fmt.Println("All cars arrived. Stopping simulation.")
			break
		}

		// print world's stats every five seconds
		if world.SimTime-lastLogTime >= 5.0 {
			fmt.Printf("[SIM] Time: %.f, | Cars: %d\n", world.SimTime, len(world.Cars))
			lastLogTime = world.SimTime
		}

		// advance the world by dt time
		world.Tick(dt)
		world.CleanArrivedCars()

		// create a new car
		if world.SimTime-lastSpawnTime >= config.Global.Simulation.SpawnRate && world.SimTime < (120.0) {
			lastSpawnTime = world.SimTime
			var (
				src, dst int
				route    []int
				err      error
			)
			spawnSuccess := false
			for range 3 {
				src, dst = randomRequest(world.Graph)
				route, err = world.Client.RequestRoute(src, dst)
				if err == nil {
					spawnSuccess = true
					break
				}
			}
			if spawnSuccess {
				lastSpawnTime = world.SimTime
				carCounter++
				newCar := world.AddCar(carCounter, carCounter)
				newCar.InitRoute(route, world.Graph)
			} else {
				fmt.Println("Skipped spawn: could not find valid route after 3 attempts")
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// return random src and dst nodeId
func randomRequest(g *graph.Graph) (int, int) {
	size := len(g.NodesArr)
	for {
		src := rand.Intn(size)
		dst := rand.Intn(size)
		if src != dst {
			return g.NodesArr[src], g.NodesArr[dst]
		}
		fmt.Println("Identical nodes id")
	}
}
