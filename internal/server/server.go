package server

import (
	"backend/internal/constant"
	"backend/internal/io_multiplxeing/poller"
	"backend/internal/payload"
	"backend/internal/protocol/resp"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
)

type Handler interface {
	HandleCmd(cmd *payload.Command, connFd int) error
}

type Server struct {
	host     string
	port     int
	protocol string
	h        Handler
}

func NewServer(host string, port int, protocol string, h Handler) *Server {
	return &Server{
		host:     host,
		port:     port,
		protocol: protocol,
		h:        h,
	}
}

func (s *Server) Start() error {
	log.Printf("I/O multiplexing server listening on %s:%v", s.host, s.port)
	listener, err := net.Listen(s.protocol, fmt.Sprintf(":%v", s.port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Get the file descriptor from the listener
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		log.Fatal("listener is not a TCPListener")
	}
	listenerFile, err := tcpListener.File()
	if err != nil {
		log.Fatal(err)
	}
	defer listenerFile.Close()

	lnFd := int(listenerFile.Fd())

	// Create a poller instance (epoll in Linux, kqueue in MacOS)
	poller, err := poller.CreatePoller()
	if err != nil {
		log.Fatal(err)
	}
	defer poller.Close()

	// Monitor "read" events on the Server FD
	if err = poller.Monitor(payload.Event{
		Fd: lnFd,
		Op: constant.OpRead,
	}); err != nil {
		log.Fatal(err)
	}

	var events = make([]payload.Event, constant.MaxConnection)
	for {
		// wait for file descriptors in the monitoring list to be ready for I/O
		// it is a blocking call.
		events, err = poller.Wait()
		if err != nil {
			continue
		}

		for i := 0; i < len(events); i++ {
			if events[i].Fd == lnFd {
				log.Printf("new client is trying to connect")
				// set up new connection
				connFd, _, err := syscall.Accept(lnFd)
				if err != nil {
					log.Println("err", err)
					continue
				}
				log.Printf("set up a new connection")
				// ask epoll to monitor this connection
				if err = poller.Monitor(payload.Event{
					Fd: connFd,
					Op: constant.OpRead,
				}); err != nil {
					log.Fatal(err)
				}
			} else {
				cmd, err := readCommand(events[i].Fd)
				if err != nil {
					if err == io.EOF || err == syscall.ECONNRESET {
						log.Println("client disconnected")
						syscall.Close(events[i].Fd)
						continue
					}
					log.Println("read error:", err)
					continue
				}

				// Handle command (Continue)
				if err = s.h.HandleCmd(cmd, events[i].Fd); err != nil {
					log.Println("handle err:", err)
				}
			}
		}
	}
}

func readCommand(fd int) (*payload.Command, error) {
	var buf = make([]byte, 512)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}
	return resp.ParseCmd(buf)
}
