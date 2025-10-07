package server

import (
	"backend/internal/io_multiplxeing/poller"
	"backend/internal/worker"
	"fmt"
	"log"
	"net"
	"syscall"
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
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			log.Printf("Unsupported connection type")
			conn.Close()
			continue
		}

		rawConn, err := tcpConn.SyscallConn()
		if err != nil {
			log.Printf("Failed to get syscall.RawConn: %v", err)
			conn.Close()
			continue
		}

		var dupFd int
		var controlErr error
		controlErr = rawConn.Control(func(fd uintptr) {
			// Duplicate FD so that Go runtime does not own the FD we will manage
			nfd, err := syscall.Dup(int(fd))
			if err != nil {
				controlErr = err
				return
			}
			dupFd = nfd
		})
		if controlErr != nil {
			log.Printf("Failed to dup fd: %v", controlErr)
			conn.Close()
			continue
		}

		// Make the duplicated fd non-blocking (epoll style)
		if err := syscall.SetNonblock(dupFd, true); err != nil {
			log.Printf("Failed to set non-blocking on fd %d: %v", dupFd, err)
			syscall.Close(dupFd)
			conn.Close()
			continue
		}

		// Round-robin
		handlerIndex := s.nextIOHandler % s.numIoHandler
		ioHandler := s.ioHandlers[handlerIndex]
		s.nextIOHandler++

		// add duplicated fd into epoll of I/O handler
		if err := ioHandler.AddConn(dupFd); err != nil {
			log.Printf("Failed to add fd %d to I/O handler %d: %v", dupFd, ioHandler.Id, err)
			syscall.Close(dupFd)
			conn.Close()
			continue
		}

		conn.Close()
	}
}
