package poller

import (
	"backend/internal/constant"
	"backend/internal/payload"
	"log"
	"syscall"
)

type Epoll struct {
	fd            int
	epollEvents   []syscall.EpollEvent
	genericEvents []payload.Event
}

func CreatePoller() (*Epoll, error) {
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Epoll{
		fd:            epollFD,
		epollEvents:   make([]syscall.EpollEvent, constant.MaxConnection),
		genericEvents: make([]payload.Event, constant.MaxConnection),
	}, nil
}

func (ep *Epoll) Monitor(event payload.Event) error {
	epollEvent := toNative(event)
	// Add event.Fd to the monitoring list of ep.fd
	return syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, event.Fd, &epollEvent)
}

func (ep *Epoll) Wait() ([]payload.Event, error) {
	n, err := syscall.EpollWait(ep.fd, ep.epollEvents, -1)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		ep.genericEvents[i] = createEvent(ep.epollEvents[i])
	}

	return ep.genericEvents[:n], nil
}

func (ep *Epoll) Close() error {
	return syscall.Close(ep.fd)
}

func toNative(e payload.Event) syscall.EpollEvent {
	var event uint32 = syscall.EPOLLIN
	if e.Op == constant.OpWrite {
		event = syscall.EPOLLOUT
	}
	return syscall.EpollEvent{
		Fd:     int32(e.Fd),
		Events: event,
	}
}

func createEvent(ep syscall.EpollEvent) payload.Event {
	var op uint32 = constant.OpRead
	if ep.Events == syscall.EPOLLOUT {
		op = constant.OpWrite
	}
	return payload.Event{
		Fd: int(ep.Fd),
		Op: op,
	}
}
