package poller

import (
	"backend/internal/config"
	"backend/internal/payload"
	"backend/internal/protocol/resp"
	"backend/internal/worker"
	"hash/fnv"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"syscall"
)

type IOHandler struct {
	Id        int
	Poller    Poller
	Workers   []*worker.Worker
	NumWorker int
	mu        sync.Mutex
	Conns     map[int]net.Conn
}

func NewIOHandler(id int, workers []*worker.Worker, numWorker int) (*IOHandler, error) {
	poller, err := CreatePoller()
	if err != nil {
		return nil, err
	}

	return &IOHandler{
		Id:        id,
		Poller:    poller,
		Workers:   workers,
		NumWorker: numWorker,
		Conns:     make(map[int]net.Conn), // map from fd to corresponding connection
	}, nil
}

func (h *IOHandler) AddConn(conn net.Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	tcpConn := conn.(*net.TCPConn)
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return err
	}

	// get the fd from connection and add it to the monitoring list for read operation
	var connFd int
	err = rawConn.Control(func(fd uintptr) {
		connFd = int(fd)
		log.Printf("I/O Handler %d is monitoring fd %d", h.Id, connFd)
		h.Conns[connFd] = conn

		// Add to epoll
		h.Poller.Monitor(payload.Event{
			Fd: connFd,
			Op: config.OpRead,
		})
	})

	return err
}

func (h *IOHandler) Start() {
	log.Printf("I/O Handler %d started", h.Id)
	for {
		events, err := h.Poller.Wait()
		if err != nil {
			continue
		}

		for _, event := range events {
			connFd := event.Fd
			h.mu.Lock()
			conn, ok := h.Conns[connFd]
			h.mu.Unlock()
			if !ok {
				continue
			}

			cmd, err := readCommandConn(conn)
			if err != nil {
				if err == io.EOF || err == syscall.ECONNRESET {
					// log.Printf("Client disconnected (fd: %d)", connFd)
				} else {
					log.Printf("Read error on fd %d: %v", connFd, err)
				}
				h.closeConn(connFd)
				continue
			}

			replyCh := make(chan []byte, 1)
			task := &payload.Task{
				Command: cmd,
				ReplyCh: replyCh,
			}
			// dispatch the command to the Worker
			h.dispatch(task)
			res := <-replyCh
			conn.Write(res)
		}
	}
}

func (h *IOHandler) dispatch(task *payload.Task) {
	var key string
	var workerID int
	if len(task.Command.Args) > 0 {
		key = task.Command.Args[0]
		workerID = h.getPartitionID(key)
	} else {
		workerID = rand.Intn(h.NumWorker)
	}

	h.Workers[workerID].TaskCh <- task
}

func (h *IOHandler) getPartitionID(key string) int {
	hasher := fnv.New32a()
	hasher.Write([]byte(key))
	return int(hasher.Sum32()) % h.NumWorker
}

func readCommandConn(conn net.Conn) (*payload.Command, error) {
	var buf = make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return resp.ParseCmd(buf[:n])
}

func (h *IOHandler) closeConn(fd int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.Conns[fd]; ok {
		conn.Close()
		delete(h.Conns, fd)
	}
}
