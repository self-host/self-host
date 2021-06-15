/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Self-host.  If not, see <http://www.gnu.org/licenses/>.
*/

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
