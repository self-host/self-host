// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/aapije/rest"
	"github.com/self-host/self-host/postgres"
)

// ThingService represents the repository used for interacting with Thing records.
type ThingService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewThingService instantiates the ThingService repository.
func NewThingService(db *sql.DB) *ThingService {
	if db == nil {
		return nil
	}

	return &ThingService{
		q:  postgres.New(db),
		db: db,
	}
}

func (svc *ThingService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := svc.q.ExistsThing(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

type AddThingParams struct {
	Name      string
	Type      *string
	CreatedBy *uuid.UUID
	Tags      []string
}

func (svc *ThingService) AddThing(ctx context.Context, p *AddThingParams) (*rest.Thing, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		// Log?
		return nil, err
	}

	tags := make([]string, 0)
	if p.Tags != nil {
		for _, tag := range p.Tags {
			tags = append(tags, tag)
		}
	}

	params := postgres.CreateThingParams{
		Name: p.Name,
		Tags: tags,
	}

	if p.Type != nil {
		params.Type.Scan(*p.Type)
	}

	if p.CreatedBy != nil {
		params.CreatedBy = *p.CreatedBy
	}

	q := svc.q.WithTx(tx)

	thing, err := q.CreateThing(ctx, params)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	v := &rest.Thing{
		Uuid:      thing.Uuid.String(),
		Name:      thing.Name,
		CreatedBy: thing.CreatedBy.String(),
		State:     rest.ThingState(thing.State),
		Tags:      thing.Tags,
	}

	if thing.Type.Valid {
		v.Type = &thing.Type.String
	}

	return v, nil
}

func (svc *ThingService) FindThingByUuid(ctx context.Context, thingUUID uuid.UUID) (*rest.Thing, error) {
	t, err := svc.q.FindThingByUUID(ctx, thingUUID)
	if err != nil {
		return nil, err
	}

	thing := &rest.Thing{
		Uuid:      t.Uuid.String(),
		Name:      t.Name,
		State:     rest.ThingState(t.State),
		CreatedBy: t.CreatedBy.String(),
		Tags:      t.Tags,
	}

	if t.Type.Valid {
		thing.Type = &t.Type.String
	}

	return thing, nil
}

func (svc *ThingService) FindAll(ctx context.Context, p FindAllParams) ([]*rest.Thing, error) {
	things := make([]*rest.Thing, 0)

	params := postgres.FindThingsParams{
		Token: p.Token,
	}

	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	thingList, err := svc.q.FindThings(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, t := range thingList {
		thing := &rest.Thing{
			Uuid:      t.Uuid.String(),
			Name:      t.Name,
			State:     rest.ThingState(t.State),
			CreatedBy: t.CreatedBy.String(),
			Tags:      t.Tags,
		}
		if t.Type.Valid {
			thing.Type = &t.Type.String
		}

		things = append(things, thing)
	}

	return things, nil
}

func (svc *ThingService) FindByTags(ctx context.Context, p FindByTagsParams) ([]*rest.Thing, error) {
	things := make([]*rest.Thing, 0)

	params := postgres.FindThingsByTagsParams{
		Tags:  p.Tags,
		Token: p.Token,
	}
	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	thingList, err := svc.q.FindThingsByTags(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, t := range thingList {
		thing := &rest.Thing{
			Uuid:      t.Uuid.String(),
			Name:      t.Name,
			State:     rest.ThingState(t.State),
			CreatedBy: t.CreatedBy.String(),
			Tags:      t.Tags,
		}
		if t.Type.Valid {
			thing.Type = &t.Type.String
		}

		things = append(things, thing)
	}

	return things, nil
}

type UpdateThingParams struct {
	Uuid  uuid.UUID
	Name  *string
	Type  *string
	State *string
	Tags  *[]string
}

func (svc *ThingService) UpdateByUuid(ctx context.Context, p UpdateThingParams) (int64, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	q := svc.q.WithTx(tx)

	var count int64

	if p.Name != nil {
		params := postgres.SetThingNameByUUIDParams{
			Uuid: p.Uuid,
			Name: *p.Name,
		}
		c, err := q.SetThingNameByUUID(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.Type != nil {
		var ns sql.NullString
		ns.Scan(p.Type)

		params := postgres.SetThingTypeByUUIDParams{
			Uuid: p.Uuid,
			Type: ns,
		}
		c, err := q.SetThingTypeByUUID(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.State != nil {
		params := postgres.SetThingStateByUUIDParams{
			Uuid:  p.Uuid,
			State: postgres.ThingState(*p.State),
		}
		c, err := q.SetThingStateByUUID(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.Tags != nil {
		params := postgres.SetThingTagsParams{
			Uuid: p.Uuid,
			Tags: *p.Tags,
		}
		c, err := q.SetThingTags(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	tx.Commit()

	return count, nil
}

func (svc *ThingService) DeleteThing(ctx context.Context, thingUUID uuid.UUID) (int64, error) {
	count, err := svc.q.DeleteThing(ctx, thingUUID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
