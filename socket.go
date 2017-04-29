// Copyright 2017 Julien Schmidt. All rights reserved.
// Use of this source code is governed by MIT license,
// a copy can be found in the LICENSE file.

// Package systemd provides functions for native systemd integration.
package systemd

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// See https://github.com/systemd/systemd/blob/master/src/libsystemd/sd-daemon/sd-daemon.c

const fdStart = 3 // first systemd socket

type Socket struct {
	f *os.File
}

// Fd returns the integer Unix file descriptor referencing the open socket.
// The file descriptor is valid only until s.Close is called or s is garbage
// collected.
func (s *Socket) Fd() uintptr {
	return s.f.Fd()
}

// Close closes the Socket, rendering it unusable for I/O.
// It returns an error, if any.
func (s *Socket) Close() error {
	return s.f.Close()
}

// File returns the underlying os.File of the socket.
// Closing f does also close s and closing s does also close f.
func (s *Socket) File() (f *os.File) {
	return s.f
}

// Listener returns a copy of the network listener corresponding to the open
// socket s.
// It is the caller's responsibility to close ln when finished.
// Closing ln does not affect s, and closing s does not affect ln.
func (s *Socket) Listener() (ln net.Listener, err error) {
	return net.FileListener(s.f)
}

// Conn returns a copy of the network connection corresponding to the open
// socket s.
// It is the caller's responsibility to close s when finished.
// Closing c does not affect s, and closing s does not affect c.
func (s *Socket) Conn() (c net.Conn, err error) {
	return net.FileConn(s.f)
}

// PacketConn returns a copy of the packet network connection corresponding
// to the open socket s.
// It is the caller's responsibility to close s when finished.
// Closing c does not affect s, and closing s does not affect c.
func (s *Socket) PacketConn() (c net.PacketConn, err error) {
	return net.FilePacketConn(s.f)
}

func parseEnv() (n int, err error) {
	envPID := os.Getenv("LISTEN_PID")
	envFDs := os.Getenv("LISTEN_FDS")

	// In Go programs there should be no need to unset the environment variables
	// as there is no API for forking.
	// if unsetEnv {
	//  os.Unsetenv("LISTEN_PID")
	//  os.Unsetenv("LISTEN_FDS")
	//  os.Unsetenv("LISTEN_FDNAMES")
	// }

	if len(envPID) == 0 {
		err = errors.New("listen enviornment not set")
		return
	}

	pid, err := strconv.Atoi(envPID)
	if err != nil {
		err = errors.New("invalid listen PID")
		return
	}

	if pid != os.Getpid() {
		err = errors.New("listen PID does not match")
		return
	}

	n, err = strconv.Atoi(envFDs)
	if err != nil {
		err = errors.New("invalid number of file descriptors")
	}
	return
}

func parseNames(n int) (names []string, err error) {
	envNames := os.Getenv("LISTEN_FDNAMES")
	if len(envNames) < 1 {
		return nil, errors.New("socket names not set")
	}

	names = strings.SplitN(envNames, ":", n)
	if len(names) != n || strings.IndexByte(names[n-1], ':') > 0 {
		return nil, errors.New("mismatch between number of socket and socket names:" +
			" expected " + strconv.Itoa(n) +
			", got " + strconv.Itoa(strings.Count(envNames, ":")+1))
	}
	return
}

// Listen returns sockets passed by the service manager as part of the
// socket-based activation logic.
// If no sockets have been received, an empty slice is returned.
// If more than one socket is received, they will be passed in the same order as
// configured in the systemd socket unit file.
func Listen() (files []Socket, err error) {
	n, err := parseEnv()
	if n < 1 { // includes err != nil case
		return
	}

	files = make([]Socket, n)
	for i := 0; i < n; i++ {
		fd := fdStart + i

		// set the close-on-exec flag for the file descriptor
		syscall.CloseOnExec(fd)

		files[i] = Socket{
			os.NewFile(uintptr(fd), "/proc/self/fd/"+strconv.Itoa(fd)),
		}
	}
	return
}

func ListenWithNames() (files []Socket, err error) {
	n, err := parseEnv()
	if n < 1 { // includes err != nil case
		return
	}

	names, err := parseNames(n)
	if err != nil {
		return
	}

	files = make([]Socket, n)
	for i := 0; i < n; i++ {
		fd := fdStart + i

		// set the close-on-exec flag for the file descriptor
		syscall.CloseOnExec(fd)

		files[i] = Socket{os.NewFile(uintptr(fd), names[i])}
	}
	return
}
