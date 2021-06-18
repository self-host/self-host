// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

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
