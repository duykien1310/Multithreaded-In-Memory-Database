package main

import (
	"backend/internal/config"
	"backend/internal/datastore"
	"backend/internal/io_multiplxeing/poller"
	"backend/internal/server"
	"backend/internal/worker"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func init() {
	config.SetConfigFile("../../internal/config")
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	configEnv := config.EnvConfig{
		Host:     config.GetString("host.address"),
		Port:     config.GetInt("host.port"),
		Protocol: config.GetString("protocol"),
	}

	// Calculate numIOHandler and numWorker
	numCores := runtime.NumCPU()
	numIOHandler := numCores / 2
	numWorker := numCores / 2
	log.Printf("Initializing server with %d workers and %d io handler\n", numWorker, numIOHandler)

	// Create Workers
	workers := make([]*worker.Worker, numWorker)
	for i := 0; i < numWorker; i++ {
		db := datastore.NewDataStore()
		workers[i] = worker.NewWorker(i, config.BufferSize, db)
	}

	// Create IOHandler
	ioHandlers := make([]*poller.IOHandler, numIOHandler)
	for i := 0; i < numIOHandler; i++ {
		ioHandler, err := poller.NewIOHandler(i, workers, numWorker)
		if err != nil {
			log.Fatalf("Failed to create I/O handler %d: %v", i, err)
		}

		ioHandlers[i] = ioHandler
	}

	s := server.NewServer(configEnv.Host, configEnv.Port, configEnv.Protocol, workers, ioHandlers, numWorker, numIOHandler)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
