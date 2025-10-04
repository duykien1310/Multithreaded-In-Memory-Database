package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultBufferSize  = 2048
	DefaultQueueSize   = 100
	DefaultWorkerCount = 5
)

type Job struct {
	Conn net.Conn
}

type Worker struct {
	ID      int
	JobChan chan Job
}

type Pool struct {
	JobQueue chan Job
	Workers  []*Worker
	Wg       sync.WaitGroup
}

type Server struct {
	Host       string
	Port       int
	Pool       *Pool
	Listener   net.Listener
	quit       chan struct{}
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewWorker(id int, jobChan chan Job) *Worker {
	return &Worker{ID: id, JobChan: jobChan}
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case job, ok := <-w.JobChan:
				if !ok {
					return
				}
				log.Printf("[Worker %d] Handling connection from %s", w.ID, job.Conn.RemoteAddr())
				handleConnection(job.Conn, w.ID)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func NewPool(workerCount, queueSize int) *Pool {
	return &Pool{
		JobQueue: make(chan Job, queueSize),
		Workers:  make([]*Worker, workerCount),
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := range p.Workers {
		worker := NewWorker(i, p.JobQueue)
		p.Workers[i] = worker
		worker.Start(ctx, &p.Wg)
	}
}

func (p *Pool) Stop() {
	close(p.JobQueue)
	p.Wg.Wait()
}

func (p *Pool) AddJob(conn net.Conn) {
	p.JobQueue <- Job{Conn: conn}
}

func NewServer(host string, port, workerCount, queueSize int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Host:       host,
		Port:       port,
		Pool:       NewPool(workerCount, queueSize),
		quit:       make(chan struct{}),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (s *Server) Start() error {
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.Listener = ln
	log.Printf("Server listening on %s", addr)

	s.Pool.Start(s.ctx)

	go s.acceptLoop()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChan:
		log.Println("Shutdown signal received...")
		s.Stop()
	case <-s.quit:
		s.Stop()
	}

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.Listener.Accept() // Blocking
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}
		log.Printf("New connection from %s", conn.RemoteAddr())
		s.Pool.AddJob(conn)
	}
}

func (s *Server) Stop() {
	s.cancelFunc()
	if s.Listener != nil {
		s.Listener.Close()
	}
	s.Pool.Stop()
	close(s.quit)
	log.Println("Server stopped gracefully")
}

func handleConnection(conn net.Conn, workerID int) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second)) // prevent idle hang
		data, err := reader.ReadString('\n')                   // Blocking
		if err != nil {
			log.Printf("[Worker %d] Client %s disconnected", workerID, conn.RemoteAddr())
			return
		}

		msg := fmt.Sprintf("Worker %d received: %s", workerID, data)
		log.Print(msg)

		// Example: respond like HTTP for demonstration
		if data == "GET / HTTP/1.1\r\n" {
			response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello from Master Server\r\n"
			conn.Write([]byte(response))
			return // close after one HTTP response
		}

		// Echo back
		conn.Write([]byte("Echo from worker " + strconv.Itoa(workerID) + ": " + data))
	}
}

// ===== Main =====
func main() {
	server := NewServer("0.0.0.0", 3000, DefaultWorkerCount, DefaultQueueSize)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
