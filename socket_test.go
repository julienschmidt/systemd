// Copyright 2017 Julien Schmidt. All rights reserved.
// Use of this source code is governed by MIT license,
// a copy can be found in the LICENSE file.

package systemd

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
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
			t.Fatalf("unexpected fd: expected %d, got %d", fdStart, rfd)
		}
		if wfd := w.Fd(); wfd != fdStart+1 {
			t.Fatalf("unexpected fd: expected %d, got %d", fdStart, wfd)
		}
	}

	return
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
	_, err := Listen()
	r.Close()
	w.Close()

	if err == nil {
		t.Fatal("did not fail when PID was not set")
	}
}

func TestListenNoFDs(t *testing.T) {
	r, w := prepareEnv(t, true, false, true)
	_, err := Listen()
	r.Close()
	w.Close()

	if err == nil {
		t.Fatal("did not fail when FDs were not set")
	}
}

func TestListenNoOpen(t *testing.T) {
	_, _ = prepareEnv(t, true, true, false)
	sockets, _ := Listen()

	if checkWrite(sockets[1].File(), sockets[0].File()) == nil {
		t.Fatal("did not fail when FDs were not opened")
	}
}
