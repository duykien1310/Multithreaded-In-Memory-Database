package event

import (
	"backend/internal/constant"
	"syscall"
)

func (e Event) ToNative(flags uint16) syscall.Kevent_t {
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

func CreateEvent(kq syscall.Kevent_t) Event {
	var op uint32 = constant.OpWrite
	if kq.Filter == syscall.EVFILT_READ {
		op = constant.OpRead
	}
	return Event{
		Fd: int(kq.Ident),
		Op: op,
	}
}
