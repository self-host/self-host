// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"

	"github.com/self-host/self-host/api/aapije/rest"
)

func TestAddUser(t *testing.T) {
	u := NewUserService(db)
	g := NewGroupService(db)

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
		var hasAlice bool
		for _, item := range gr {
			if item.Name == "alice" {
				hasAlice = true
			}
		}
		return hasAlice
	}(alice.Groups) == false {
		log.Fatal("User alice is missing alice group")
	}

	// Ensure the bob user belongs to the bob group
	if func(gr []rest.Group) bool {
		var hasBob bool
		for _, item := range gr {
			if item.Name == "bob" {
				hasBob = true
			}
		}
		return hasBob
	}(bob.Groups) == false {
		log.Fatal("User bob is missing bob group")
	}

	var limit = int64(100)
	var offset = int64(0)

	// Check and ensure the user specific groups has been added to the groups table
	groups, err := g.FindAll(context.Background(), []byte(rootToken), &limit, &offset)
	if err != nil {
		log.Fatal(err)
	}

	if func(gr []*rest.Group) bool {
		var hasAlice bool
		var hasBob bool

		for _, item := range gr {
			if item.Name == "alice" {
				hasAlice = true
			} else if item.Name == "bob" {
				hasBob = true
			}
		}

		return hasAlice && hasBob
	}(groups) == false {
		log.Fatal("Either alice or bob group is missing")
	}
}
