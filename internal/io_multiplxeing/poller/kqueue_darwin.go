package poller

import (
	"backend/internal/constant"
	"backend/internal/payload"
	"log"
	"syscall"
)

type KQueue struct {
	fd            int
	kqEvents      []syscall.Kevent_t
	genericEvents []payload.Event
}

func CreatePoller() (*KQueue, error) {
	epollFD, err := syscall.Kqueue()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &KQueue{
		fd:            epollFD,
		kqEvents:      make([]syscall.Kevent_t, constant.MaxConnection),
		genericEvents: make([]payload.Event, constant.MaxConnection),
	}, nil
}

func (kq *KQueue) Monitor(event payload.Event) error {
	kqEvent := toNative(event, syscall.EV_ADD)
	// Add event.Fd to the monitoring list of kq.fd
	_, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{kqEvent}, nil, nil)

	return err
}

func (kq *KQueue) Wait() ([]payload.Event, error) {
	n, err := syscall.Kevent(kq.fd, nil, kq.kqEvents, nil) // It will sleep
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		kq.genericEvents[i] = createEvent(kq.kqEvents[i])
	}

	return kq.genericEvents[:n], nil
}

func (kq *KQueue) Close() error {
	return syscall.Close(kq.fd)
}

func toNative(e payload.Event, flags uint16) syscall.Kevent_t {
	var filter int16 = syscall.EVFILT_WRITE
	if e.Op == constant.OpRead {
		filter = syscall.EVFILT_READ
	}
	return syscall.Kevent_t{
		Ident:  uint64(e.Fd),
		Filter: filter,
		Flags:  flags,
	}
}

func createEvent(kq syscall.Kevent_t) payload.Event {
	var op uint32 = constant.OpWrite
	if kq.Filter == syscall.EVFILT_READ {
		op = constant.OpRead
	}
	return payload.Event{
		Fd: int(kq.Ident),
		Op: op,
	}
}
