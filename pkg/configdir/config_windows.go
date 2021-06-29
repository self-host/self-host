// +build windows

// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package configdir

import (
	"os"
)

func findSystemPaths() []string {
	return []string{
		os.Getenv("PROGRAMDATA"),
	}
}

func findLocalPaths() []string {
	return []string{
		os.Getenv("APPDATA"),
	}
}
