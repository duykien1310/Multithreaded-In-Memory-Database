package poller

import (
	"backend/internal/config"
	"backend/internal/payload"
	"errors"
	"log"
	"syscall"
)

type KQueue struct {
	fd            int
	kqEvents      []syscall.Kevent_t
	genericEvents []payload.Event
}

func CreatePoller() (*KQueue, error) {
	kqFD, err := syscall.Kqueue()
	if err != nil {
		log.Printf("CreatePoller: kqueue failed: %v", err)
		return nil, err
	}

	return &KQueue{
		fd:            kqFD,
		kqEvents:      make([]syscall.Kevent_t, config.MaxConnection),
		genericEvents: make([]payload.Event, config.MaxConnection),
	}, nil
}

func (kq *KQueue) Monitor(event payload.Event) error {
	flags := syscall.EV_ADD | syscall.EV_ENABLE
	kev := toNative(event, uint16(flags))

	_, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{kev}, nil, nil)
	if err != nil {
		log.Printf("KQueue.Monitor: failed to add fd=%d op=%d: %v", event.Fd, event.Op, err)
		return err
	}
	return nil
}

func (kq *KQueue) Remove(fd int) error {
	deleteRead := toNative(payload.Event{Fd: fd, Op: config.OpRead}, syscall.EV_DELETE)
	deleteWrite := toNative(payload.Event{Fd: fd, Op: config.OpWrite}, syscall.EV_DELETE)

	var firstErr error

	if _, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{deleteRead}, nil, nil); err != nil {
		if !errors.Is(err, syscall.ENOENT) {
			firstErr = err
			log.Printf("KQueue.Remove: failed to delete read filter fd=%d: %v", fd, err)
		}
	}

	if _, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{deleteWrite}, nil, nil); err != nil {
		if !errors.Is(err, syscall.ENOENT) {
			if firstErr == nil {
				firstErr = err
			}
			log.Printf("KQueue.Remove: failed to delete write filter fd=%d: %v", fd, err)
		}
	}

	return firstErr
}

func (kq *KQueue) Wait() ([]payload.Event, error) {
	for {
		n, err := syscall.Kevent(kq.fd, nil, kq.kqEvents, nil)
		if err != nil {
			// If interrupted by signal, retry.
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return nil, err
		}

		// Convert native kevent_t entries to generic events.
		for i := 0; i < n; i++ {
			kq.genericEvents[i] = createEvent(kq.kqEvents[i])
		}
		return kq.genericEvents[:n], nil
	}
}

func (kq *KQueue) Close() error {
	if kq == nil {
		return nil
	}
	return syscall.Close(kq.fd)
}

func toNative(e payload.Event, flags uint16) syscall.Kevent_t {
	var filter int16 = syscall.EVFILT_WRITE
	if e.Op == config.OpRead {
		filter = syscall.EVFILT_READ
	}
	return syscall.Kevent_t{
		Ident:  uint64(e.Fd),
		Filter: filter,
		Flags:  flags,
	}
}

func createEvent(kq syscall.Kevent_t) payload.Event {
	var op uint32 = config.OpWrite
	if kq.Filter == syscall.EVFILT_READ {
		op = config.OpRead
	}
	return payload.Event{
		Fd: int(kq.Ident),
		Op: op,
	}
}
