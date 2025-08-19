package poller

// Cross-OS, minimal poller facade with epoll/kqueue backends.

type Event struct {
	FD     int
	Events uint32
}

const (
	EvRead uint32 = 1 << iota
	EvWrite
	EvErr
	EvHup
	EvEdge // edge-triggered
)

func (e Event) Readable() bool { return e.Events&EvRead != 0 }
func (e Event) Writable() bool { return e.Events&EvWrite != 0 }
func (e Event) HasErr() bool   { return e.Events&EvErr != 0 }
func (e Event) HasHup() bool   { return e.Events&EvHup != 0 }

// Poller is implemented by epoll (linux) or kqueue (darwin).
type Poller interface {
	Add(fd int, flags uint32) error
	Mod(fd int, flags uint32) error
	Del(fd int) error
	Wait(dst []Event) (int, error)
}

func New() (Poller, error) { return newPlatformPoller() }
