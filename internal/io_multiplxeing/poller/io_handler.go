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

			// inside IOHandler.Start() loop, after building `task` and before dispatching normally:
			if task.Command.Cmd == "KEYS" {
				// Broadcast to all workers and aggregate
				// create per-worker reply channels and send tasks
				replyChs := make([]chan []byte, len(h.Workers))
				for i, wk := range h.Workers {
					ch := make(chan []byte, 1)
					replyChs[i] = ch
					// create shallow copy of task for this worker
					t := &payload.Task{
						Command: task.Command,
						ReplyCh: ch,
					}
					wk.TaskCh <- t
				}

				// Collect replies from all workers, merge arrays of keys
				var allKeys []string
				for _, ch := range replyChs {
					res := <-ch
					// decode response — resp.ParseResp? your resp package encodes.
					// Here we assume the worker returned an array encoded with resp.Encode([]string,...)
					// So we need to parse it back. Use resp.ParseReply helper if available,
					// otherwise, decode a simple RESP array manually. For simplicity we use resp.DecodeBytes
					// (You may need to adapt to your resp package.)

					parsed, err := resp.Decode(res)
					if err != nil {
						// if a worker returned an error, send that error to client
						conn.Write(res)
						break
					}
					// parsed should be []string (or []interface{}) depending on your resp.Decode
					if arr, ok := parsed.([]string); ok {
						allKeys = append(allKeys, arr...)
					} else if arrI, ok := parsed.([]interface{}); ok {
						for _, v := range arrI {
							if s, ok := v.(string); ok {
								allKeys = append(allKeys, s)
							}
						}
					} else {
						// unknown decode form -- ignore or handle
					}
				}

				// remove duplicates (since same key should be on only one worker, duplicates unlikely)
				// but we'll deduplicate defensively:
				uniq := make(map[string]struct{}, len(allKeys))
				deduped := make([]string, 0, len(allKeys))
				for _, k := range allKeys {
					if _, seen := uniq[k]; !seen {
						uniq[k] = struct{}{}
						deduped = append(deduped, k)
					}
				}

				// encode final response
				final := resp.Encode(deduped, false)
				conn.Write(final)

				// continue to next event (we already wrote response)
				continue
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
