// Copyright 2017 Julien Schmidt. All rights reserved.
// Use of this source code is governed by MIT license,
// a copy can be found in the LICENSE file.

package systemd

import (
	"os/exec"
	"testing"
)

func TestBooted(t *testing.T) {
	booted := Booted()
	isDir := exec.Command("ls", "/run/systemd/system").Run() == nil
	if booted != isDir {
		t.Fatalf("/run/systemd/system is a dir: %t, Booted(): %t", isDir, booted)
	}
}
