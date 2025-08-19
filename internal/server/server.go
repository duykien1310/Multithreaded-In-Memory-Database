package server

import (
	"backend/internal/netutil"
	"backend/internal/poller"
	"errors"
	"fmt"
	"log"
	"sync"
	"syscall"
)

// Handler defines pluggable protocol logic (e.g., RESP, line-based, etc.).
type Handler interface {
	// OnRead is called when bytes arrive. Return reply bytes (may be nil) and
	// whether to close the connection after writing.
	OnRead(fd int, in []byte) (out []byte, closeAfterWrite bool, err error)
}

// Server owns the listening socket, poller, and connection buffers.
type Server struct {
	addr       string
	lpfd       int           // listening socket fd
	p          poller.Poller // epoll/kqueue underneath
	handler    Handler
	mu         sync.Mutex
	connBufMap map[int][]byte // per-connection read buffer
}

func New(addr string, h Handler) (*Server, error) {
	lpfd, sa, err := netutil.NewTCPListenerFD(addr)
	if err != nil {
		return nil, err
	}
	p, err := poller.New()
	if err != nil {
		return nil, err
	}
	if err := p.Add(lpfd, poller.EvRead|poller.EvEdge); err != nil {
		return nil, fmt.Errorf("add listen fd: %w", err)
	}
	log.Printf("listen socket ready: %s", netutil.PrettySockaddr(sa))

	return &Server{
		addr:       addr,
		lpfd:       lpfd,
		p:          p,
		handler:    h,
		connBufMap: make(map[int][]byte),
	}, nil
}

func (s *Server) Serve() error {
	defer syscall.Close(s.lpfd)

	events := make([]poller.Event, 256)
	tmp := make([]byte, 4096)

	for {
		n, err := s.p.Wait(events)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return fmt.Errorf("poller wait: %w", err)
		}
		for i := 0; i < n; i++ {
			ev := events[i]
			fd := ev.FD

			if ev.HasErr() || ev.HasHup() {
				if fd != s.lpfd {
					s.closeClient(fd)
				}
				continue
			}

			if fd == s.lpfd && ev.Readable() {
				// Accept all pending (edge-triggered)
				for {
					cfd, _, aerr := syscall.Accept(s.lpfd)
					if aerr != nil {
						if aerr == syscall.EAGAIN || aerr == syscall.EWOULDBLOCK {
							break
						}
						log.Printf("accept: %v", aerr)
						break
					}
					_ = netutil.SetNonblock(cfd)
					_ = s.p.Add(cfd, poller.EvRead|poller.EvEdge)
					log.Printf("accepted %v fd=%d", aerr, cfd)
				}
				continue
			}

			if ev.Readable() {
				// Drain reads until EAGAIN
				for {
					nread, rerr := syscall.Read(fd, tmp)
					if nread > 0 {
						s.appendBuf(fd, tmp[:nread])
						// Try handling as much as handler wants (here: single pass)
						out, closeAfter, herr := s.handler.OnRead(fd, s.getBuf(fd))
						if herr != nil {
							// Protocol error → close
							s.closeClient(fd)
							break
						}
						if len(out) > 0 {
							// Best-effort write (short writes possible)
							if _, werr := syscall.Write(fd, out); werr != nil && werr != syscall.EAGAIN {
								s.closeClient(fd)
								break
							}
						}
						if closeAfter {
							s.closeClient(fd)
							break
						}
						// Reset buffer if handler consumed it all
						s.clearBuf(fd)
						continue
					}
					if rerr != nil {
						if rerr == syscall.EAGAIN || rerr == syscall.EWOULDBLOCK {
							break
						}
						s.closeClient(fd)
						break
					}
					if nread == 0 {
						s.closeClient(fd)
						break
					}
				}
			}
		}
	}
}

func (s *Server) appendBuf(fd int, b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connBufMap[fd] = append(s.connBufMap[fd], b...)
}
func (s *Server) getBuf(fd int) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.connBufMap[fd]...)
}
func (s *Server) clearBuf(fd int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connBufMap[fd] = s.connBufMap[fd][:0]
}
func (s *Server) closeClient(fd int) {
	_ = s.p.Del(fd)
	_ = syscall.Close(fd)
	s.mu.Lock()
	delete(s.connBufMap, fd)
	s.mu.Unlock()
}
