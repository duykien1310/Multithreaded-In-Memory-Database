package poller

import (
	"backend/internal/constant"
	"backend/internal/io_multiplxeing/event"
	"log"
	"syscall"
)

type Epoll struct {
	fd            int
	epollEvents   []syscall.EpollEvent
	genericEvents []event.Event
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
		genericEvents: make([]event.Event, constant.MaxConnection),
	}, nil
}

func (ep *Epoll) Monitor(event event.Event) error {
	epollEvent := event.ToNative()
	// Add event.Fd to the monitoring list of ep.fd
	return syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, event.Fd, &epollEvent)
}

func (ep *Epoll) Wait() ([]event.Event, error) {
	n, err := syscall.EpollWait(ep.fd, ep.epollEvents, -1)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		ep.genericEvents[i] = event.CreateEvent(ep.epollEvents[i])
	}

	return ep.genericEvents[:n], nil
}

func (ep *Epoll) Close() error {
	return syscall.Close(ep.fd)
}
