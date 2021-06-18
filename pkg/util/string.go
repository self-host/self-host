// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package util

import (
	"github.com/google/uuid"
)

func StringSliceIndex(haystack []string, needle string) int {
	for idx, a := range haystack {
		if a == needle {
			return idx
		}
	}
	return -1
}

func StringSliceContains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if a == needle {
			return true
		}
	}
	return false
}

func StringSliceToUuidSlice(in []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, len(in))
	for idx, item := range in {
		id, err := uuid.Parse(item)
		if err != nil {
			return out, err
		}
		out[idx] = id
	}

	return out, nil
}
