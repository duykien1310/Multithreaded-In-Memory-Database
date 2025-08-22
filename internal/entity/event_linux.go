package entity

import (
	"backend/internal/constant"
	"syscall"
)

func (e Event) ToNative() syscall.EpollEvent {
	var event uint32 = syscall.EPOLLIN
	if e.Op == constant.OpWrite {
		event = syscall.EPOLLOUT
	}
	return syscall.EpollEvent{
		Fd:     int32(e.Fd),
		Events: event,
	}
}

func CreateEvent(ep syscall.EpollEvent) Event {
	var op uint32 = constant.OpRead
	if ep.Events == syscall.EPOLLOUT {
		op = constant.OpWrite
	}
	return Event{
		Fd: int(ep.Fd),
		Op: op,
	}
}
