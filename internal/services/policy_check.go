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

	"github.com/self-host/self-host/postgres"
)

type PolicyCheckService struct {
	q *postgres.Queries
}

// NewPolicyCheck service
func NewPolicyCheckService(db *sql.DB) *PolicyCheckService {
	return &PolicyCheckService{
		q: postgres.New(db),
	}
}

func (pc *PolicyCheckService) UserHasAccessViaToken(ctx context.Context, token []byte, action string, resource string) (bool, error) {
	params := postgres.CheckUserTokenHasAccessParams{
		Action:   postgres.PolicyAction(action),
		Resource: resource,
		Token:    token,
	}

	hasAccess, err := pc.q.CheckUserTokenHasAccess(ctx, params)
	if err != nil {
		return false, err
	}

	return hasAccess, nil
}
