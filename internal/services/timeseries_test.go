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

package services

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"
)

// Tests can run in any order, so we need to run everything (Timeseries related) in one function
// as we are required to do things in a certain order since we are not mocking the PostgreSQL data-store.
func TestTimeseriesAll(t *testing.T) {
	svc := NewTimeseriesService(db)

	params := &NewTimeseriesParams{
		Name:      "MyTimeseries",
		CreatedBy: uuid.MustParse("00000000-0000-1000-8000-000000000000"), // UUID for Root user
		Tags:      []string{},
	}

	timeseries, err := svc.AddTimeseries(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	tsUUID, err := uuid.Parse(timeseries.Uuid)
	if err != nil {
		log.Fatal(err)
	}

	if timeseries.Name != "MyTimeseries" {
		log.Fatal("Name does not match expected")
	}

	if tsUUID == uuid.MustParse("00000000-0000-0000-0000-000000000000") {
		log.Fatal("UUID of new time series is nil")
	}

	count, err := svc.DeleteTimeseries(context.Background(), tsUUID)
	if err != nil {
		log.Fatal(err)
	} else if count == 0 {
		log.Fatal("Timeseries was not deleted")
	}
}

type RangeRowT struct {
	V   float32
	Le  *float32
	Ge  *float32
	Res bool
}

func Float32P(v float32) *float32 {
	return &v
}

func TestInValidRange(t *testing.T) {
	// value, le, ge
	checks := []RangeRowT{
		{
			V:   10,
			Le:  Float32P(10),
			Ge:  Float32P(10),
			Res: true,
		},
		{
			V:   10,
			Le:  Float32P(100),
			Ge:  Float32P(-100),
			Res: true,
		},
		{
			V:   10,
			Le:  Float32P(-100),
			Ge:  Float32P(100),
			Res: false,
		},
		{
			V:   -150,
			Le:  Float32P(-100),
			Ge:  Float32P(100),
			Res: true,
		},
		{
			V:   150,
			Le:  Float32P(-100),
			Ge:  Float32P(100),
			Res: true,
		},
		{
			V:   -100,
			Le:  Float32P(-100),
			Res: true,
		},
		{
			V:   100,
			Ge:  Float32P(100),
			Res: true,
		},
		{
			V:   101,
			Le:  Float32P(100),
			Res: false,
		},
		{
			V:   99,
			Ge:  Float32P(100),
			Res: false,
		},
	}

	for _, row := range checks {
		if inValidRange(row.V, row.Le, row.Ge) != row.Res {
			log.Fatal("Check failed")
		}
	}
}
