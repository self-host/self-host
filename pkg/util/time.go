// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package util

import (
	"time"
)

func AtInterval(d time.Duration) <-chan time.Time {
	t := time.Now().Truncate(d).Add(d).Sub(time.Now())
	return time.After(t)
}
