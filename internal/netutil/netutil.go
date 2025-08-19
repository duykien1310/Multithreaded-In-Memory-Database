package netutil

import (
	"fmt"
	"net"
	"strconv"

	"golang.org/x/sys/unix"
)

func SetNonblock(fd int) error { return unix.SetNonblock(fd, true) }

func NewTCPListenerFD(addr string) (int, unix.Sockaddr, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, nil, err
	}
	port, _ := strconv.Atoi(portStr)

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return 0, nil, err
	}
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		return 0, nil, err
	}
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return 0, nil, err
	}

	sa := &unix.SockaddrInet4{Port: port}
	if host != "" && host != "0.0.0.0" {
		ip := net.ParseIP(host).To4()
		copy(sa.Addr[:], ip)
	}

	if err := unix.Bind(fd, sa); err != nil {
		return 0, nil, err
	}
	if err := unix.Listen(fd, 1024); err != nil {
		return 0, nil, err
	}
	if err := unix.SetNonblock(fd, true); err != nil {
		return 0, nil, err
	}
	return fd, sa, nil
}

func PrettySockaddr(sa unix.Sockaddr) string {
	switch x := sa.(type) {
	case *unix.SockaddrInet4:
		return fmt.Sprintf("%s:%d", net.IP(x.Addr[:]).String(), x.Port)
	case *unix.SockaddrInet6:
		return fmt.Sprintf("[%s]:%d", net.IP(x.Addr[:]).String(), x.Port)
	default:
		return "unknown"
	}
}
