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
along with dogtag.  If not, see <http://www.gnu.org/licenses/>.
*/

package services

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/selfserv/rest"
	pg "github.com/self-host/self-host/postgres"
)

type DatasetFile struct {
	Format  string
	Content []byte
}

// DatasetService represents the repository used for interacting with Dataset records.
type DatasetService struct {
	q  *pg.Queries
	db *sql.DB
}

// NewDatasetService instantiates the DatasetService repository.
func NewDatasetService(db *sql.DB) *DatasetService {
	if db == nil {
		return nil
	}

	return &DatasetService{
		q:  pg.New(db),
		db: db,
	}
}

type AddDatasetParams struct {
	Name      string
	Format    string
	Content   []byte
	CreatedBy uuid.UUID
	BelongsTo uuid.UUID
}

func (svc *DatasetService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := svc.q.ExistsDataset(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

func (svc *DatasetService) AddDataset(ctx context.Context, p *AddDatasetParams) (*rest.Dataset, error) {
	params := pg.CreateDatasetParams{
		Name:      p.Name,
		Content:   p.Content,
		Format:    p.Format,
		CreatedBy: p.CreatedBy,
		BelongsTo: p.BelongsTo,
	}

	dataset, err := svc.q.CreateDataset(ctx, params)
	if err != nil {
		return nil, err
	}

	v := &rest.Dataset{
		Uuid:      dataset.Uuid.String(),
		Name:      dataset.Name,
		Format:    rest.DatasetFormat(dataset.Format),
		Size:      int64(dataset.Size),
		Created:   dataset.Created,
		Updated:   dataset.Updated,
		CreatedBy: dataset.CreatedBy.String(),
		UpdatedBy: dataset.UpdatedBy.String(),
	}

	if dataset.BelongsTo != NilUUID {
		belongs_to := dataset.BelongsTo.String()
		v.BelongsTo = &belongs_to
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
		Size:      int64(dataset.Size),
		Created:   dataset.Created,
		Updated:   dataset.Updated,
		CreatedBy: dataset.CreatedBy.String(),
		UpdatedBy: dataset.UpdatedBy.String(),
	}

	if dataset.BelongsTo != NilUUID {
		belongs_to := dataset.BelongsTo.String()
		v.BelongsTo = &belongs_to
	}

	return v, nil
}

func (svc *DatasetService) FindByThing(ctx context.Context, id uuid.UUID) ([]*rest.Dataset, error) {
	datasets := make([]*rest.Dataset, 0)

	datasets_list, err := svc.q.FindDatasetByThing(ctx, id)
	if err != nil {
		return nil, err
	} else {
		for _, t := range datasets_list {
			dataset := &rest.Dataset{
				Uuid:      t.Uuid.String(),
				Name:      t.Name,
				Size:      int64(t.Size),
				Format:    rest.DatasetFormat(t.Format),
				Created:   t.Created,
				Updated:   t.Updated,
				CreatedBy: t.CreatedBy.String(),
				UpdatedBy: t.UpdatedBy.String(),
			}

			if t.BelongsTo != NilUUID {
				v := t.BelongsTo.String()
				dataset.BelongsTo = &v
			}

			datasets = append(datasets, dataset)
		}
	}

	return datasets, nil
}

func (svc *DatasetService) FindAll(ctx context.Context, token []byte, limit *int64, offset *int64) ([]*rest.Dataset, error) {
	datasets := make([]*rest.Dataset, 0)

	params := pg.FindDatasetsParams{
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

	datasets_list, err := svc.q.FindDatasets(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, t := range datasets_list {
			dataset := &rest.Dataset{
				Uuid:      t.Uuid.String(),
				Name:      t.Name,
				Size:      int64(t.Size),
				Format:    rest.DatasetFormat(t.Format),
				Created:   t.Created,
				Updated:   t.Updated,
				CreatedBy: t.CreatedBy.String(),
				UpdatedBy: t.UpdatedBy.String(),
			}

			if t.BelongsTo != NilUUID {
				v := t.BelongsTo.String()
				dataset.BelongsTo = &v
			}

			datasets = append(datasets, dataset)
		}
	}

	return datasets, nil
}

func (svc *DatasetService) GetDatasetContentByUuid(ctx context.Context, id uuid.UUID) (*DatasetFile, error) {
	content, err := svc.q.GetDatasetContentByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &DatasetFile{
		Format:  content.Format,
		Content: content.Content,
	}, nil
}

type UpdateDatasetByUuidParams struct {
	Content *[]byte
	Format  *string
	Name    *string
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
		c, err := q.SetDatasetNameByUUID(ctx, pg.SetDatasetNameByUUIDParams{
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
		c, err := q.SetDatasetFormatByUUID(ctx, pg.SetDatasetFormatByUUIDParams{
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
		c, err := q.SetDatasetContentByUUID(ctx, pg.SetDatasetContentByUUIDParams{
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
