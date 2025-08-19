package poller

import "golang.org/x/sys/unix"

type kq struct{ fd int }

func newPlatformPoller() (Poller, error) {
	kqfd, err := unix.Kqueue()
	if err != nil {
		return nil, err
	}
	// register a user event so kqueue is valid
	return &kq{fd: kqfd}, nil
}

func (p *kq) Add(fd int, flags uint32) error {
	ev := []unix.Kevent_t{}
	if flags&EvRead != 0 {
		ev = append(ev, unix.Kevent_t{Ident: uint64(fd), Filter: unix.EVFILT_READ, Flags: unix.EV_ADD})
	}
	if flags&EvWrite != 0 {
		ev = append(ev, unix.Kevent_t{Ident: uint64(fd), Filter: unix.EVFILT_WRITE, Flags: unix.EV_ADD})
	}
	// Edge-triggered-ish via EV_CLEAR (deliver when state changes)
	for i := range ev {
		ev[i].Flags |= unix.EV_CLEAR
	}
	_, err := unix.Kevent(p.fd, ev, nil, nil)
	return err
}

func (p *kq) Mod(fd int, flags uint32) error {
	// For simplicity, delete then add
	_ = p.Del(fd)
	return p.Add(fd, flags)
}

func (p *kq) Del(fd int) error {
	ev := []unix.Kevent_t{
		{Ident: uint64(fd), Filter: unix.EVFILT_READ, Flags: unix.EV_DELETE},
		{Ident: uint64(fd), Filter: unix.EVFILT_WRITE, Flags: unix.EV_DELETE},
	}
	_, err := unix.Kevent(p.fd, ev, nil, nil)
	return err
}

func (p *kq) Wait(dst []Event) (int, error) {
	ke := make([]unix.Kevent_t, len(dst))
	n, err := unix.Kevent(p.fd, nil, ke, nil)
	if err != nil {
		return 0, err
	}
	for i := 0; i < n; i++ {
		e := Event{FD: int(ke[i].Ident)}
		switch ke[i].Filter {
		case unix.EVFILT_READ:
			e.Events |= EvRead
		case unix.EVFILT_WRITE:
			e.Events |= EvWrite
		}
		if ke[i].Flags&(unix.EV_EOF) != 0 {
			e.Events |= EvHup
		}
		dst[i] = e
	}
	return n, nil
}
