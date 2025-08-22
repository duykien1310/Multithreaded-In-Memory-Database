package poller

import (
	"backend/internal/constant"
	"backend/internal/io_multiplxeing/event"
	"log"
	"syscall"
)

type KQueue struct {
	fd            int
	kqEvents      []syscall.Kevent_t
	genericEvents []event.Event
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
		genericEvents: make([]event.Event, constant.MaxConnection),
	}, nil
}

func (kq *KQueue) Monitor(event event.Event) error {
	kqEvent := event.ToNative(syscall.EV_ADD)
	// Add event.Fd to the monitoring list of kq.fd
	_, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{kqEvent}, nil, nil)

	return err
}

func (kq *KQueue) Wait() ([]event.Event, error) {
	n, err := syscall.Kevent(kq.fd, nil, kq.kqEvents, nil) // It will sleep
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		kq.genericEvents[i] = event.CreateEvent(kq.kqEvents[i])
	}

	return kq.genericEvents[:n], nil
}

func (kq *KQueue) Close() error {
	return syscall.Close(kq.fd)
}
