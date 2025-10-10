package poller

import (
	"backend/internal/config"
	"backend/internal/payload"
	"errors"
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
		log.Printf("Epoll: failed to create epoll instance: %v", err)
		return nil, err
	}

	return &Epoll{
		fd:            epollFD,
		epollEvents:   make([]syscall.EpollEvent, config.MaxConnection),
		genericEvents: make([]payload.Event, config.MaxConnection),
	}, nil
}

func (ep *Epoll) Monitor(event payload.Event) error {
	ev := toNative(event)

	err := syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, event.Fd, &ev)
	if err != nil {
		if errors.Is(err, syscall.EEXIST) {
			err = syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_MOD, event.Fd, &ev)
		}
	}

	if err != nil {
		log.Printf("Epoll.Monitor: failed to add/mod fd=%d, err=%v", event.Fd, err)
	}
	return err
}

func (ep *Epoll) Remove(fd int) error {
	err := syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil && !errors.Is(err, syscall.ENOENT) {
		log.Printf("Epoll.Remove: failed to remove fd=%d: %v", fd, err)
		return err
	}
	return nil
}

func (ep *Epoll) Wait() ([]payload.Event, error) {
	for {
		n, err := syscall.EpollWait(ep.fd, ep.epollEvents, -1)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return nil, err
		}

		for i := 0; i < n; i++ {
			ep.genericEvents[i] = createEvent(ep.epollEvents[i])
		}
		return ep.genericEvents[:n], nil
	}
}

func (ep *Epoll) Close() error {
	if ep == nil {
		return nil
	}
	return syscall.Close(ep.fd)
}

func toNative(e payload.Event) syscall.EpollEvent {
	var ev uint32 = syscall.EPOLLIN
	if e.Op == config.OpWrite {
		ev = syscall.EPOLLOUT
	}

	return syscall.EpollEvent{
		Events: ev,
		Fd:     int32(e.Fd),
	}
}

func createEvent(ep syscall.EpollEvent) payload.Event {
	var op uint32 = config.OpRead
	if (ep.Events & syscall.EPOLLOUT) != 0 {
		op = config.OpWrite
	}
	return payload.Event{
		Fd: int(ep.Fd),
		Op: op,
	}
}
