package server

import (
	"backend/internal/io_multiplxeing/poller"
	"backend/internal/worker"
	"fmt"
	"log"
	"net"
)

type Server struct {
	host          string
	port          int
	protocol      string
	numWorker     int
	numIoHandler  int
	workers       []*worker.Worker
	ioHandlers    []*poller.IOHandler
	nextIOHandler int
}

func NewServer(host string, port int, protocol string, workers []*worker.Worker, ioHandlers []*poller.IOHandler, numWorker int, numIoHandler int) *Server {
	return &Server{
		host:         host,
		port:         port,
		protocol:     protocol,
		workers:      workers,
		ioHandlers:   ioHandlers,
		numWorker:    numWorker,
		numIoHandler: numIoHandler,
	}
}

func (s *Server) Start() error {
	log.Printf("I/O multiplexing server listening on %s:%v", s.host, s.port)
	listener, err := net.Listen(s.protocol, fmt.Sprintf(":%v", s.port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for _, worker := range s.workers {
		go worker.Start()
	}

	for _, ioHandler := range s.ioHandlers {
		go ioHandler.Start()
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to acccept connection: %v", err)
			continue
		}

		// forward the new connection to an I/O handler in a round-robin manner
		handler := s.ioHandlers[s.nextIOHandler%s.numIoHandler]
		s.nextIOHandler++

		if err := handler.AddConn(conn); err != nil {
			log.Printf("Failed to add connection to I/O handler %d: %v", handler.Id, err)
			conn.Close()
		}
	}
}
