package poller

import "backend/internal/payload"

type Poller interface {
	Monitor(event payload.Event) error
	Wait() ([]payload.Event, error)
	Close() error
	Remove(fd int) error
}
