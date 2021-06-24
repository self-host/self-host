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

type DatasetFile struct {
	Format   string
	Content  []byte
	Checksum string
}

// DatasetService represents the repository used for interacting with Dataset records.
type DatasetService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewDatasetService instantiates the DatasetService repository.
func NewDatasetService(db *sql.DB) *DatasetService {
	if db == nil {
		return nil
	}

	return &DatasetService{
		q:  postgres.New(db),
		db: db,
	}
}

type AddDatasetParams struct {
	Name      string
	Format    string
	Content   []byte
	CreatedBy uuid.UUID
	ThingUuid uuid.UUID
	Tags      []string
}

func (svc *DatasetService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := svc.q.ExistsDataset(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

func (svc *DatasetService) AddDataset(ctx context.Context, p *AddDatasetParams) (*rest.Dataset, error) {
	tags := make([]string, 0)
	if p.Tags != nil {
		for _, tag := range p.Tags {
			tags = append(tags, tag)
		}
	}

	params := postgres.CreateDatasetParams{
		Name:      p.Name,
		Content:   p.Content,
		Format:    p.Format,
		CreatedBy: p.CreatedBy,
		BelongsTo: p.ThingUuid,
		Tags:      tags,
	}

	dataset, err := svc.q.CreateDataset(ctx, params)
	if err != nil {
		return nil, err
	}

	v := &rest.Dataset{
		Uuid:      dataset.Uuid.String(),
		Name:      dataset.Name,
		Format:    rest.DatasetFormat(dataset.Format),
		Checksum:  dataset.Checksum,
		Size:      int64(dataset.Size),
		Created:   dataset.Created,
		Updated:   dataset.Updated,
		CreatedBy: dataset.CreatedBy.String(),
		UpdatedBy: dataset.UpdatedBy.String(),
		Tags:      dataset.Tags,
	}

	if dataset.BelongsTo != NilUUID {
		belongsTo := dataset.BelongsTo.String()
		v.ThingUuid = &belongsTo
	}

	return v, nil
}

func (svc *DatasetService) FindDatasetByUuid(ctx context.Context, id uuid.UUID) (*rest.Dataset, error) {
	dataset, err := svc.q.FindDatasetByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	v := &rest.Dataset{
		Uuid:      dataset.Uuid.String(),
		Name:      dataset.Name,
		Format:    rest.DatasetFormat(dataset.Format),
		Checksum:  dataset.Checksum,
		Size:      int64(dataset.Size),
		Created:   dataset.Created,
		Updated:   dataset.Updated,
		CreatedBy: dataset.CreatedBy.String(),
		UpdatedBy: dataset.UpdatedBy.String(),
		Tags:      dataset.Tags,
	}

	if dataset.BelongsTo != NilUUID {
		belongsTo := dataset.BelongsTo.String()
		v.ThingUuid = &belongsTo
	}

	return v, nil
}

func (svc *DatasetService) FindByThing(ctx context.Context, id uuid.UUID) ([]*rest.Dataset, error) {
	datasets := make([]*rest.Dataset, 0)

	datasetsList, err := svc.q.FindDatasetByThing(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, t := range datasetsList {
		dataset := &rest.Dataset{
			Uuid:      t.Uuid.String(),
			Name:      t.Name,
			Format:    rest.DatasetFormat(t.Format),
			Checksum:  t.Checksum,
			Size:      int64(t.Size),
			Created:   t.Created,
			Updated:   t.Updated,
			CreatedBy: t.CreatedBy.String(),
			UpdatedBy: t.UpdatedBy.String(),
			Tags:      t.Tags,
		}

		if t.BelongsTo != NilUUID {
			v := t.BelongsTo.String()
			dataset.ThingUuid = &v
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

func (svc *DatasetService) FindAll(ctx context.Context, p FindAllParams) ([]*rest.Dataset, error) {
	datasets := make([]*rest.Dataset, 0)

	params := postgres.FindDatasetsParams{
		Token: p.Token,
	}

	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	datasetsList, err := svc.q.FindDatasets(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, t := range datasetsList {
		dataset := &rest.Dataset{
			Uuid:      t.Uuid.String(),
			Name:      t.Name,
			Format:    rest.DatasetFormat(t.Format),
			Checksum:  t.Checksum,
			Size:      int64(t.Size),
			Created:   t.Created,
			Updated:   t.Updated,
			CreatedBy: t.CreatedBy.String(),
			UpdatedBy: t.UpdatedBy.String(),
			Tags:      t.Tags,
		}

		if t.BelongsTo != NilUUID {
			v := t.BelongsTo.String()
			dataset.ThingUuid = &v
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

func (svc *DatasetService) FindByTags(ctx context.Context, p FindByTagsParams) ([]*rest.Dataset, error) {
	datasets := make([]*rest.Dataset, 0)

	params := postgres.FindDatasetsByTagsParams{
		Tags:  p.Tags,
		Token: p.Token,
	}
	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	dsList, err := svc.q.FindDatasetsByTags(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, t := range dsList {
		dataset := &rest.Dataset{
			Uuid:      t.Uuid.String(),
			Name:      t.Name,
			Format:    rest.DatasetFormat(t.Format),
			Checksum:  t.Checksum,
			Size:      int64(t.Size),
			Created:   t.Created,
			Updated:   t.Updated,
			CreatedBy: t.CreatedBy.String(),
			UpdatedBy: t.UpdatedBy.String(),
			Tags:      t.Tags,
		}

		if t.BelongsTo != NilUUID {
			v := t.BelongsTo.String()
			dataset.ThingUuid = &v
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

func (svc *DatasetService) GetDatasetContentByUuid(ctx context.Context, id uuid.UUID) (*DatasetFile, error) {
	content, err := svc.q.GetDatasetContentByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &DatasetFile{
		Format:   content.Format,
		Content:  content.Content,
		Checksum: content.Checksum,
	}, nil
}

type UpdateDatasetByUuidParams struct {
	Content *[]byte
	Format  *string
	Name    *string
	Tags    *[]string
}

func (svc *DatasetService) UpdateDatasetByUuid(ctx context.Context, id uuid.UUID, p UpdateDatasetByUuidParams) (int64, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	var count int64

	q := svc.q.WithTx(tx)

	if p.Name != nil {
		c, err := q.SetDatasetNameByUUID(ctx, postgres.SetDatasetNameByUUIDParams{
			Uuid: id,
			Name: *p.Name,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Format != nil {
		c, err := q.SetDatasetFormatByUUID(ctx, postgres.SetDatasetFormatByUUIDParams{
			Uuid:   id,
			Format: *p.Format,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Content != nil {
		c, err := q.SetDatasetContentByUUID(ctx, postgres.SetDatasetContentByUUIDParams{
			Uuid:    id,
			Content: *p.Content,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Tags != nil {
		params := postgres.SetDatasetTagsParams{
			Uuid: id,
			Tags: *p.Tags,
		}
		c, err := q.SetDatasetTags(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	tx.Commit()

	return count, nil
}

func (svc *DatasetService) DeleteDataset(ctx context.Context, id uuid.UUID) (int64, error) {
	count, err := svc.q.DeleteDataset(ctx, id)
	if err != nil {
		return 0, err
	}

	return count, nil
}
