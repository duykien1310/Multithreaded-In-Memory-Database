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
	"syscall"
)

type IOHandler struct {
	Id        int
	Poller    Poller
	Workers   []*worker.Worker
	NumWorker int
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
	}, nil
}

// Add connection to the handler's epoll monitoring list
func (h *IOHandler) AddConn(connFd int) error {
	// Add to epoll
	err := h.Poller.Monitor(payload.Event{
		Fd: connFd,
		Op: config.OpRead,
	})

	return err
}

func (h *IOHandler) Start() {
	for {
		// wait for file descriptors in the monitoring list to be ready for I/O
		// it is a blocking call.
		events, err := h.Poller.Wait()
		if err != nil {
			continue
		}

		for i := 0; i < len(events); i++ {
			fd := events[i].Fd

			cmd, err := readCommand(fd)
			if err != nil {
				if err == io.EOF || err == syscall.ECONNRESET {
					log.Printf("client disconnected fd=%d", fd)
					// Remove from epoll first, then close the fd
					if remErr := h.Poller.Remove(fd); remErr != nil {
						log.Printf("failed to remove fd %d from poller: %v", fd, remErr)
					}
					syscall.Close(fd)
					continue
				}
				// If EAGAIN/EWOULDBLOCK, just ignore (no data now)
				if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
					continue
				}
				log.Println("read error:", err)
				// On other fatal errors, also cleanup
				if remErr := h.Poller.Remove(fd); remErr == nil {
					syscall.Close(fd)
				} else {
					// ensure close anyway and log
					syscall.Close(fd)
				}
				continue
			}

			// dispatch task to worker (unchanged)
			task := &payload.Task{
				Command: cmd,
				ConnFd:  fd,
			}
			h.dispatch(task)
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

func readCommand(fd int) (*payload.Command, error) {
	var buf = make([]byte, 512)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
			return nil, err
		}
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}
	return resp.ParseCmd(buf[:n])
}
