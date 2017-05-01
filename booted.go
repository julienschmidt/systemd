// Copyright 2017 Julien Schmidt. All rights reserved.
// Use of this source code is governed by MIT license,
// a copy can be found in the LICENSE file.

package systemd

import (
	"os"
)

// Booted checks whether the system was booted up using the systemd init system.
// This functions internally checks whether the runtime unit file directory
// "/run/systemd/system" exists and is thus specific to systemd.
func Booted() bool {
	fi, err := os.Lstat("/run/systemd/system")
	return err == nil && fi.IsDir()
}
