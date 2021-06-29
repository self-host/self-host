// +build darwin

// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package configdir

import (
	"os"
	"path/filepath"
)

func findSystemPaths() []string {
	return []string{
		"/Library/Application Support",
	}
}

func findLocalPaths() []string {
	return []string{
		filepath.Join(
			os.Getenv("HOME"),
			"/Library/Application Support",
		),
	}
}
