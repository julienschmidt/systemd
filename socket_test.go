// Copyright 2017 Julien Schmidt. All rights reserved.
// Use of this source code is governed by MIT license,
// a copy can be found in the LICENSE file.

package systemd

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
)

func prepareEnv(t *testing.T, setPID, setFDs, openFDs bool) (r, w *os.File) {
	os.Clearenv()
	if setPID {
		os.Setenv("LISTEN_PID", strconv.Itoa(os.Getpid()))
	}

	if setFDs {
		os.Setenv("LISTEN_FDS", "2")
	}

	if openFDs {
		// adds 2 more FDs
		r, w, _ = os.Pipe()
		if rfd := r.Fd(); rfd != fdStart {
			cleanEnv(r, w)
			t.Fatalf("unexpected fd: expected %d, got %d", fdStart, rfd)
		}
		if wfd := w.Fd(); wfd != fdStart+1 {
			cleanEnv(r, w)
			t.Fatalf("unexpected fd: expected %d, got %d", fdStart, wfd)
		}
	}

	return
}

func prepareNames(n int) {
	if n < 1 {
		os.Setenv("LISTEN_FDNAMES", "")
		return
	}

	names := ""
	for i := 0; i < n; i++ {
		names += ":fd" + strconv.Itoa(i+fdStart)
	}
	os.Setenv("LISTEN_FDNAMES", names[1:])
}

func cleanEnv(r, w *os.File) {
	os.Unsetenv("LISTEN_PID")
	os.Unsetenv("LISTEN_FDS")
	os.Unsetenv("LISTEN_FDNAMES")

	r.Close()
	w.Close()
}

func checkWrite(w io.WriteCloser, r io.ReadCloser) (err error) {
	testStr := "This test is totally sufficient\n"

	if _, err = w.Write([]byte(testStr)); err != nil {
		return
	}
	if err = w.Close(); err != nil {
		return
	}

	out, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	if err = r.Close(); err != nil {
		return
	}

	if string(out) != testStr {
		return errors.New("string mismatch")
	}

	return
}

func TestListen(t *testing.T) {
	r, w := prepareEnv(t, true, true, true)
	defer cleanEnv(r, w)

	sockets, err := Listen()
	if err != nil {
		t.Fatal(err)
	}

	if len(sockets) != 2 {
		t.Fatalf("unexpected number of sockets: expected 2, got %d", len(sockets))
	}

	if r.Fd() != sockets[0].Fd() || w.Fd() != sockets[1].Fd() {
		t.Fatalf("file descriptor mismatch: %d=%d, %d=%d", r.Fd(), sockets[0].Fd(), w.Fd(), sockets[1].Fd())
	}

	if err = checkWrite(sockets[1].File(), sockets[0].File()); err != nil {
		t.Fatal(err)
	}
}

func TestListenNoPID(t *testing.T) {
	r, w := prepareEnv(t, false, true, true)
	defer cleanEnv(r, w)

	if _, err := Listen(); err == nil {
		t.Fatal("did not fail when PID was not set")
	}
}

func TestListenInvalidPID(t *testing.T) {
	r, w := prepareEnv(t, true, true, true)
	os.Setenv("LISTEN_PID", "Gordon")
	defer cleanEnv(r, w)

	if _, err := Listen(); err == nil {
		t.Fatal("did not fail when PID was invalid")
	}
}

func TestListenWrongPID(t *testing.T) {
	r, w := prepareEnv(t, true, true, true)
	os.Setenv("LISTEN_PID", "1")
	defer cleanEnv(r, w)

	if _, err := Listen(); err == nil {
		t.Fatal("did not fail when PID mismatched")
	}
}

func TestListenNoFDs(t *testing.T) {
	r, w := prepareEnv(t, true, false, true)
	defer cleanEnv(r, w)

	if _, err := Listen(); err == nil {
		t.Fatal("did not fail when FDs were not set")
	}
}

func TestListenNoOpen(t *testing.T) {
	r, w := prepareEnv(t, true, true, false)
	defer cleanEnv(r, w)

	sockets, _ := Listen()
	if checkWrite(sockets[1].File(), sockets[0].File()) == nil {
		t.Fatal("did not fail when FDs were not opened")
	}
}

func checkListenWithNames(t *testing.T, names []string) {
	r, w := prepareEnv(t, true, true, true)
	os.Setenv("LISTEN_FDNAMES", strings.Join(names, ":"))
	defer cleanEnv(r, w)

	sockets, err := ListenWithNames()
	if err != nil {
		t.Fatal(err)
	}

	if len(sockets) != len(names) {
		t.Fatalf("unexpected number of sockets: expected %d, got %d", len(names), len(sockets))
	}

	for i, name := range names {
		if sockets[i].Name() != name {
			t.Fatalf("unexpected socket name: expected %s, got %s", name, sockets[i].Name())
		}
	}

	if r.Fd() != sockets[0].Fd() || w.Fd() != sockets[1].Fd() {
		t.Fatalf("file descriptor mismatch: %d=%d, %d=%d", r.Fd(), sockets[0].Fd(), w.Fd(), sockets[1].Fd())
	}

	if err = checkWrite(sockets[1].File(), sockets[0].File()); err != nil {
		t.Fatal(err)
	}
}

func TestListenWithNames(t *testing.T) {
	checkListenWithNames(t, []string{"fd3", "fd4"})
}

func TestListenWithNamesEmpty(t *testing.T) {
	checkListenWithNames(t, []string{"", ""})
}

func TestListenWithNamesNoPID(t *testing.T) {
	r, w := prepareEnv(t, false, true, true)
	prepareNames(2)
	defer cleanEnv(r, w)

	if _, err := ListenWithNames(); err == nil {
		t.Fatal("did not fail when PID was not set")
	}
}

func TestListenWithNamesInvalidPID(t *testing.T) {
	r, w := prepareEnv(t, true, true, true)
	prepareNames(2)
	os.Setenv("LISTEN_PID", "Gordon")
	defer cleanEnv(r, w)

	if _, err := ListenWithNames(); err == nil {
		t.Fatal("did not fail when PID was invalid")
	}
}

func TestListenWithNamesWrongPID(t *testing.T) {
	r, w := prepareEnv(t, true, true, true)
	prepareNames(2)
	os.Setenv("LISTEN_PID", "1")
	defer cleanEnv(r, w)

	if _, err := ListenWithNames(); err == nil {
		t.Fatal("did not fail when PID mismatched")
	}
}

func TestListenWithNamesNoFDs(t *testing.T) {
	r, w := prepareEnv(t, true, false, true)
	prepareNames(2)
	defer cleanEnv(r, w)

	if _, err := ListenWithNames(); err == nil {
		t.Fatal("did not fail when FDs were not set")
	}
}

func TestListenWithNamesMismatch(t *testing.T) {
	r, w := prepareEnv(t, true, true, true)
	defer cleanEnv(r, w)

	if _, err := ListenWithNames(); err == nil {
		t.Fatal("no error when no names were set")
	}

	prepareNames(0)
	if _, err := ListenWithNames(); err == nil {
		t.Fatal("no error when no names were set")
	}

	prepareNames(1)
	if _, err := ListenWithNames(); err == nil {
		t.Fatal("no error when too few names were set")
	}

	prepareNames(3)
	if _, err := ListenWithNames(); err == nil {
		t.Fatal("no error when too many names were set")
	}
}

func TestSocket(t *testing.T) {
	r, w := prepareEnv(t, false, false, true)
	defer cleanEnv(r, w)

	s := Socket{w}

	if s.Fd() != w.Fd() {
		t.Fatalf("socket FD mismatch: expected %d, got %d", w.Fd(), s.Fd())
	}

	if s.Name() != w.Name() {
		t.Fatalf("socket name mismatch: expected %s, got %s", w.Name(), s.Name())
	}

	if err := s.Close(); err != nil {
		t.Fatalf("error while closing socket: %v", err)
	}
}

func TestSocketListener(t *testing.T) {
	l1, err := net.Listen("tcp", ":55555")
	if err != nil {
		t.Fatal(err)
	}
	defer l1.Close()

	f, err := l1.(*net.TCPListener).File()
	if err != nil {
		t.Fatal(err)
	}

	s := newSocket(int(f.Fd()), f.Name())

	if s.Fd() != f.Fd() {
		t.Fatalf("socket FD mismatch: expected %d, got %d", f.Fd(), s.Fd())
	}

	if s.Name() != f.Name() {
		t.Fatalf("socket name mismatch: expected %s, got %s", f.Name(), s.Name())
	}

	l2, err := s.Listener()
	if err != nil {
		t.Fatal(err)
	}

	if err = l2.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSocketConn(t *testing.T) {
	l1, err := net.Listen("tcp", ":55556")
	if err != nil {
		t.Fatal(err)
	}
	defer l1.Close()

	c1, err := net.Dial("tcp", ":55556")
	if err != nil {
		t.Fatalf(err.Error())
	}

	f, err := c1.(*net.TCPConn).File()
	if err != nil {
		t.Fatal(err)
	}

	s := newSocket(int(f.Fd()), f.Name())

	if s.Fd() != f.Fd() {
		t.Fatalf("socket FD mismatch: expected %d, got %d", f.Fd(), s.Fd())
	}

	if s.Name() != f.Name() {
		t.Fatalf("socket name mismatch: expected %s, got %s", f.Name(), s.Name())
	}

	c2, err := s.Conn()
	if err != nil {
		t.Fatal(err)
	}

	if err = c2.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSocketPacketConn(t *testing.T) {
	c1, err := net.ListenPacket("udp", ":55557")
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	f, err := c1.(*net.UDPConn).File()
	if err != nil {
		t.Fatal(err)
	}

	s := newSocket(int(f.Fd()), f.Name())

	if s.Fd() != f.Fd() {
		t.Fatalf("socket FD mismatch: expected %d, got %d", f.Fd(), s.Fd())
	}

	if s.Name() != f.Name() {
		t.Fatalf("socket name mismatch: expected %s, got %s", f.Name(), s.Name())
	}

	c2, err := s.PacketConn()
	if err != nil {
		t.Fatal(err)
	}

	if err = c2.Close(); err != nil {
		t.Fatal(err)
	}
}
