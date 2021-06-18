// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/postgres"
)

type FindAllPoliciesParams struct {
	Token      []byte
	Limit      *int64
	Offset     *int64
	GroupUuids *[]uuid.UUID
}

// PolicyService represents the repository used for interacting with Policy records.
type PolicyService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewPolicyService instantiates the PolicyService repository.
func NewPolicyService(db *sql.DB) *PolicyService {
	return &PolicyService{
		q:  postgres.New(db),
		db: db,
	}
}

type NewPolicyParams struct {
	GroupUuid uuid.UUID
	Priority  int32
	Effect    string
	Action    string
	Resource  string
}

func (u *PolicyService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := u.q.ExistsPolicy(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

func (s *PolicyService) Add(ctx context.Context, p NewPolicyParams) (*rest.Policy, error) {
	policy, err := s.q.CreatePolicy(ctx, postgres.CreatePolicyParams{
		GroupUuid: p.GroupUuid,
		Priority:  p.Priority,
		Effect:    postgres.PolicyEffect(p.Effect),
		Action:    postgres.PolicyAction(p.Action),
		Resource:  p.Resource,
	})
	if err != nil {
		return nil, err
	}
	return &rest.Policy{
		Uuid:      policy.Uuid.String(),
		GroupUuid: policy.GroupUuid.String(),
		Priority:  policy.Priority,
		Effect:    rest.PolicyEffect(policy.Effect),
		Action:    rest.PolicyAction(policy.Action),
		Resource:  policy.Resource,
	}, nil
}

func (s *PolicyService) FindByGroup(ctx context.Context, groupUUID uuid.UUID) ([]*rest.Policy, error) {
	policies := make([]*rest.Policy, 0)

	count, err := s.q.ExistsGroup(ctx, groupUUID)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, ie.ErrorNotFound
	}

	policyList, err := s.q.FindPoliciesByGroup(ctx, groupUUID)
	if err != nil {
		return nil, err
	} else {
		for _, item := range policyList {
			policies = append(policies, &rest.Policy{
				Uuid:      item.Uuid.String(),
				GroupUuid: item.GroupUuid.String(),
				Priority:  item.Priority,
				Effect:    rest.PolicyEffect(item.Effect),
				Action:    rest.PolicyAction(item.Action),
				Resource:  item.Resource,
			})
		}
	}

	return policies, nil
}

func (s *PolicyService) FindByUser(ctx context.Context, userUUID uuid.UUID) ([]*rest.Policy, error) {
	policies := make([]*rest.Policy, 0)

	count, err := s.q.ExistsUser(ctx, userUUID)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, ie.ErrorNotFound
	}

	policyList, err := s.q.FindPoliciesByUser(ctx, userUUID)
	if err != nil {
		return nil, err
	} else {
		for _, item := range policyList {
			policies = append(policies, &rest.Policy{
				Uuid:      item.Uuid.String(),
				GroupUuid: item.GroupUuid.String(),
				Priority:  item.Priority,
				Effect:    rest.PolicyEffect(item.Effect),
				Action:    rest.PolicyAction(item.Action),
				Resource:  item.Resource,
			})
		}
	}

	return policies, nil
}

func (s *PolicyService) FindByUuid(ctx context.Context, policyUUID uuid.UUID) (*rest.Policy, error) {
	item, err := s.q.FindPolicyByUUID(ctx, policyUUID)
	if err != nil {
		return nil, err
	}

	return &rest.Policy{
		Uuid:      item.Uuid.String(),
		GroupUuid: item.GroupUuid.String(),
		Priority:  item.Priority,
		Effect:    rest.PolicyEffect(item.Effect),
		Action:    rest.PolicyAction(item.Action),
		Resource:  item.Resource,
	}, nil
}

func (s *PolicyService) FindAll(ctx context.Context, p FindAllPoliciesParams) ([]*rest.Policy, error) {
	policies := make([]*rest.Policy, 0)

	params := postgres.FindPoliciesParams{
		Token:     p.Token,
		ArgLimit:  20,
		ArgOffset: 0,
	}
	if p.Limit != nil {
		params.ArgLimit = *p.Limit
	}
	if p.Offset != nil {
		params.ArgOffset = *p.Offset
	}
	if p.GroupUuids != nil {
		params.GroupUuids = *p.GroupUuids
	}

	pList, err := s.q.FindPolicies(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, v := range pList {
			policies = append(policies, &rest.Policy{
				Uuid:      v.Uuid.String(),
				GroupUuid: v.GroupUuid.String(),
				Priority:  v.Priority,
				Effect:    rest.PolicyEffect(v.Effect),
				Action:    rest.PolicyAction(v.Action),
				Resource:  v.Resource,
			})
		}
	}

	return policies, nil
}

type UpdatePolicyParams struct {
	GroupUuid *uuid.UUID
	Priority  *int
	Effect    *string
	Action    *string
	Resource  *string
}

func (s *PolicyService) Update(ctx context.Context, id uuid.UUID, p UpdatePolicyParams) (int64, error) {
	var count int64

	found, err := s.Exists(ctx, id)
	if err != nil {
		return 0, err
	} else if found == false {
		return 0, ie.ErrorNotFound
	}

	// Use a transaction for this action
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		// Log?
		return 0, err
	}

	q := s.q.WithTx(tx)

	if p.GroupUuid != nil {
		c, err := q.SetPolicyGroup(ctx, postgres.SetPolicyGroupParams{
			Uuid:      id,
			GroupUuid: *p.GroupUuid,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.Priority != nil {
		c, err := q.SetPolicyPriority(ctx, postgres.SetPolicyPriorityParams{
			Uuid:     id,
			Priority: int32(*p.Priority),
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.Effect != nil {
		c, err := q.SetPolicyEffect(ctx, postgres.SetPolicyEffectParams{
			Uuid:   id,
			Effect: postgres.PolicyEffect(*p.Effect),
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.Action != nil {
		c, err := q.SetPolicyAction(ctx, postgres.SetPolicyActionParams{
			Uuid:   id,
			Action: postgres.PolicyAction(*p.Action),
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.Resource != nil {
		c, err := q.SetPolicyResource(ctx, postgres.SetPolicyResourceParams{
			Uuid:     id,
			Resource: *p.Resource,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	tx.Commit()

	return count, nil
}

func (s *PolicyService) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64

	count, err := s.q.DeletePolicyByUUID(ctx, id)
	if err != nil {
		return 0, err
	} else if count == 0 {
		return 0, ie.ErrorNotFound
	}

	return count, nil
}
