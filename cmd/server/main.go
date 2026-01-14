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
	if err := config.Load("config.json"); err != nil {
		panic(err)
	}
	fmt.Printf("Hello Waze!, map file is: %s\n", config.Global.Server.MapFile)
	srv := server.NewServer(config.Global.Server.MapFile)
	fmt.Printf("The num of cores is: %d\n", runtime.NumCPU())

	server.WakeWorkers(runtime.NumCPU(), srv.Graph)
	
	// הפעלת WebSocket Hub
	server.GlobalHub = server.NewHub()
	go server.GlobalHub.Run()

	// API endpoints
	http.HandleFunc("/api/traffic", srv.HandleTrafficBatch)
	http.HandleFunc("/api/navigate", srv.HandleNavigation)
	http.HandleFunc("/ws", srv.HandleWebSocket)
	
	// הגשת קבצי GUI סטטיים
	http.Handle("/", http.FileServer(http.Dir("web")))

	log.Printf("Server running on: %s\n", config.Global.Server.Port)
	log.Printf("GUI available at: http://localhost%s\n", config.Global.Server.Port)
	log.Fatal(http.ListenAndServe(config.Global.Server.Port, nil))
}