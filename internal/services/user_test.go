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

package services_test

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"

	"github.com/self-host/self-host/api/selfserv/rest"
	"github.com/self-host/self-host/internal/services"
)

func TestAddUser(t *testing.T) {
	u := services.NewUserService(db)
	g := services.NewGroupService(db)

	alice, err := u.AddUser(context.Background(), "alice")
	if err != nil {
		log.Fatal(err)
	}

	_, err = uuid.Parse(alice.Uuid)
	if err != nil {
		log.Fatal(err)
	}

	bob, err := u.AddUser(context.Background(), "bob")
	if err != nil {
		log.Fatal(err)
	}

	_, err = uuid.Parse(bob.Uuid)
	if err != nil {
		log.Fatal(err)
	}

	if bob.Name != "bob" {
		log.Fatalf("Name does not match %v != %v", bob.Name, "bob")
	}

	// Ensure the alice user belongs to the alice group
	if func(gr []rest.Group) bool {
		var has_alice bool
		for _, item := range gr {
			if item.Name == "alice" {
				has_alice = true
			}
		}
		return has_alice
	}(alice.Groups) == false {
		log.Fatal("User alice is missing alice group")
	}

	// Ensure the bob user belongs to the bob group
	if func(gr []rest.Group) bool {
		var has_bob bool
		for _, item := range gr {
			if item.Name == "bob" {
				has_bob = true
			}
		}
		return has_bob
	}(bob.Groups) == false {
		log.Fatal("User bob is missing bob group")
	}

	var limit = int64(100)
	var offset = int64(0)

	// Check and ensure the user specific groups has been added to the groups table
	groups, err := g.FindAll(context.Background(), []byte(ROOT_TOKEN), &limit, &offset)
	if err != nil {
		log.Fatal(err)
	}

	if func(gr []*rest.Group) bool {
		var has_alice bool
		var has_bob bool

		for _, item := range gr {
			if item.Name == "alice" {
				has_alice = true
			} else if item.Name == "bob" {
				has_bob = true
			}
		}

		return has_alice && has_bob
	}(groups) == false {
		log.Fatal("Either alice or bob group is missing")
	}
}
