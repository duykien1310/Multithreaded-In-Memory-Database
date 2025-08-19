package poller

import "golang.org/x/sys/unix"

type ep struct{ fd int }

func newPlatformPoller() (Poller, error) {
	epfd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	return &ep{fd: epfd}, nil
}

func (p *ep) toUnixFlags(flags uint32) uint32 {
	ev := uint32(0)
	if flags&EvRead != 0 {
		ev |= unix.EPOLLIN
	}
	if flags&EvWrite != 0 {
		ev |= unix.EPOLLOUT
	}
	if flags&EvEdge != 0 {
		ev |= unix.EPOLLET
	}
	return ev
}

func (p *ep) Add(fd int, flags uint32) error {
	ev := &unix.EpollEvent{Events: p.toUnixFlags(flags), Fd: int32(fd)}
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, ev)
}
func (p *ep) Mod(fd int, flags uint32) error {
	ev := &unix.EpollEvent{Events: p.toUnixFlags(flags), Fd: int32(fd)}
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd, ev)
}
func (p *ep) Del(fd int) error { return unix.EpollCtl(p.fd, unix.EPOLL_CTL_DEL, fd, nil) }

func (p *ep) Wait(dst []Event) (int, error) {
	evs := make([]unix.EpollEvent, len(dst))
	n, err := unix.EpollWait(p.fd, evs, -1)
	if err != nil {
		return 0, err
	}
	for i := 0; i < n; i++ {
		e := Event{FD: int(evs[i].Fd)}
		if evs[i].Events&(unix.EPOLLIN) != 0 {
			e.Events |= EvRead
		}
		if evs[i].Events&(unix.EPOLLOUT) != 0 {
			e.Events |= EvWrite
		}
		if evs[i].Events&(unix.EPOLLERR) != 0 {
			e.Events |= EvErr
		}
		if evs[i].Events&(unix.EPOLLHUP|unix.EPOLLRDHUP) != 0 {
			e.Events |= EvHup
		}
		dst[i] = e
	}
	return n, nil
}
