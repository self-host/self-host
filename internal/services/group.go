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
	"database/sql"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/postgres"
)

// User represents the repository used for interacting with User records.
type GroupService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewUser instantiates the User repository.
func NewGroupService(db *sql.DB) *GroupService {
	return &GroupService{
		q:  postgres.New(db),
		db: db,
	}
}

func (u *GroupService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := u.q.ExistsGroup(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

func (u *GroupService) AddGroup(ctx context.Context, name string) (*rest.Group, error) {
	// Use a transaction for this action
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		// Log?
		return nil, err
	}

	q := u.q.WithTx(tx)

	group, err := q.CreateGroup(ctx, name)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else {
		tx.Commit()
	}

	return &rest.Group{
		Uuid: group.Uuid.String(),
		Name: group.Name,
	}, nil
}

func (u *GroupService) FindGroupByUuid(ctx context.Context, groupUUID uuid.UUID) (*rest.Group, error) {
	found, err := u.Exists(ctx, groupUUID)
	if err != nil {
		return nil, err
	} else if found == false {
		return nil, ie.ErrorNotFound
	}

	group, err := u.q.FindGroupByUuid(ctx, groupUUID)
	if err != nil {
		return nil, err
	}

	return &rest.Group{
		Uuid: group.Uuid.String(),
		Name: group.Name,
	}, nil
}

// Find returns all groups
func (u *GroupService) FindAll(ctx context.Context, token []byte, limit *int64, offset *int64) ([]*rest.Group, error) {
	groups := make([]*rest.Group, 0)

	params := postgres.FindGroupsParams{
		Token:     token,
		ArgLimit:  20,
		ArgOffset: 0,
	}
	if limit != nil {
		params.ArgLimit = *limit
	}
	if offset != nil {
		params.ArgOffset = *offset
	}

	groupList, err := u.q.FindGroups(ctx, params)

	if err != nil {
		return nil, err
	} else {
		for _, g := range groupList {
			groups = append(groups, &rest.Group{
				Uuid: g.Uuid.String(),
				Name: g.Name,
			})
		}
	}

	return groups, nil
}

func (u *GroupService) UpdateGroupNameByUuid(ctx context.Context, id uuid.UUID, name string) (int64, error) {
	count, err := u.q.SetGroupNameByUUID(ctx, postgres.SetGroupNameByUUIDParams{
		Uuid: id,
		Name: name,
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (u *GroupService) DeleteGroup(ctx context.Context, groupUUID uuid.UUID) (int64, error) {
	count, err := u.q.DeleteGroup(ctx, groupUUID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
