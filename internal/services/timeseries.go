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
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/noda/selfhost/api/selfserv/rest"
	ie "github.com/noda/selfhost/internal/errors"
	pg "github.com/noda/selfhost/postgres"

	units "github.com/ganehag/go-units"
)

const insertDataToTimeseries = `
SELECT tsdata_insert(
	$1::uuid,
	x.v,
	x.ts,
	$2::uuid
)
FROM
json_to_recordset($3::json) AS x("v" double precision, "ts" timestamptz);
`

// NewTimeseries defines model for NewTimeseries.
type NewTimeseriesParams struct {
	CreatedBy  uuid.UUID
	ThingUuid  uuid.UUID
	Name       string
	SiUnit     string
	Tags       []string
	LowerBound sql.NullFloat64
	UpperBound sql.NullFloat64
}

// User represents the repository used for interacting with User records.
type TimeseriesService struct {
	q  *pg.Queries
	db *sql.DB
}

// NewUser instantiates the User repository.
func NewTimeseriesService(db *sql.DB) *TimeseriesService {
	if db == nil {
		return nil
	}

	return &TimeseriesService{
		q:  pg.New(db),
		db: db,
	}
}

func (svc *TimeseriesService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := svc.q.ExistsTimeseries(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

func (svc *TimeseriesService) AddTimeseries(ctx context.Context, opt *NewTimeseriesParams) (*rest.Timeseries, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	q := svc.q.WithTx(tx)

	tags := make([]string, 0)
	if opt.Tags != nil {
		for _, tag := range opt.Tags {
			tags = append(tags, tag)
		}
	}

	params := pg.CreateTimeseriesParams{
		CreatedBy:  opt.CreatedBy,
		ThingUuid:  opt.ThingUuid,
		Name:       opt.Name,
		SiUnit:     opt.SiUnit,
		LowerBound: opt.LowerBound,
		UpperBound: opt.UpperBound,
		Tags:       tags,
	}

	timeseries, err := q.CreateTimeseries(ctx, params)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else {
		tx.Commit()
	}

	var lb *float64
	var ub *float64

	if timeseries.LowerBound.Valid {
		lb = &timeseries.LowerBound.Float64
	}

	if timeseries.UpperBound.Valid {
		ub = &timeseries.UpperBound.Float64
	}

	t := &rest.Timeseries{
		Uuid:       timeseries.Uuid.String(),
		CreatedBy:  timeseries.CreatedBy.String(),
		Name:       timeseries.Name,
		SiUnit:     timeseries.SiUnit,
		LowerBound: lb,
		UpperBound: ub,
		Tags:       timeseries.Tags,
	}

	if timeseries.ThingUuid != NilUUID {
		v := timeseries.ThingUuid.String()
		t.ThingUuid = &v
	}

	return t, nil
}

type AddDataToTimeseriesParams struct {
	Uuid      uuid.UUID
	Points    []DataPoint
	CreatedBy uuid.UUID
	Unit      *string
}

func (svc *TimeseriesService) AddDataToTimeseries(ctx context.Context, p AddDataToTimeseriesParams) (int64, error) {
	series, err := svc.q.GetTimeseriesByUUID(ctx, p.Uuid)
	if err != nil {
		return 0, err
	}

	filteredPoints := make([]*DataPoint, 0)

	var from_unit units.Unit
	var to_unit units.Unit

	if p.Unit != nil {
		ts_unit, err := svc.q.GetUnitFromTimeseries(ctx, p.Uuid)
		if err != nil {
			return 0, err
		}

		if ts_unit == *p.Unit {
			p.Unit = nil
		} else {

			from_unit, err = units.Find(*p.Unit)
			if err != nil {
				return 0, ie.ErrorInvalidUnit
			}

			to_unit, err = units.Find(ts_unit)
			if err != nil {
				// This should never error out, as there should be no incompatible units in the DB
				return 0, ie.ErrorInvalidUnit
			}
		}
	}

	for _, item := range p.Points {
		// Do not use a pointer to the item variable as this is a known gotcha.
		p_item := item

		if p.Unit != nil {
			v := units.NewValue(p_item.Value, from_unit)
			conv, err := v.Convert(to_unit)
			if err != nil {
				return 0, ie.ErrorInvalidUnitConversion
			}
			p_item.Value = float64(conv.Float())
		}

		if series.LowerBound.Valid {
			// Should we skip this value
			if p_item.Value < series.LowerBound.Float64 {
				continue
			}
		}
		if series.UpperBound.Valid {
			// Should we skip this value
			if p_item.Value > series.UpperBound.Float64 {
				continue
			}
		}

		filteredPoints = append(filteredPoints, &p_item)
	}

	data, err := json.Marshal(filteredPoints)
	if err != nil {
		return 0, err
	}

	result, err := svc.db.ExecContext(ctx, insertDataToTimeseries, p.Uuid, p.CreatedBy, data)
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (svc *TimeseriesService) FindByTags(ctx context.Context, p FindByTagsParams) ([]*rest.Timeseries, error) {
	timeseries := make([]*rest.Timeseries, 0)

	params := pg.FindTimeseriesByTagsParams{
		Tags:  p.Tags,
		Token: p.Token,
	}
	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	ts_list, err := svc.q.FindTimeseriesByTags(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, item := range ts_list {
			var lBound *float64
			var uBound *float64

			if item.LowerBound.Valid {
				lBound = &item.LowerBound.Float64
			}
			if item.UpperBound.Valid {
				uBound = &item.UpperBound.Float64
			}

			t := &rest.Timeseries{
				CreatedBy:  item.CreatedBy.String(),
				LowerBound: lBound,
				Name:       item.Name,
				SiUnit:     item.SiUnit,
				Tags:       item.Tags,
				UpperBound: uBound,
				Uuid:       item.Uuid.String(),
			}

			if item.ThingUuid != NilUUID {
				v := item.ThingUuid.String()
				t.ThingUuid = &v
			}

			timeseries = append(timeseries, t)
		}
	}

	return timeseries, nil
}

func (svc *TimeseriesService) FindByThing(ctx context.Context, thing uuid.UUID) ([]*rest.Timeseries, error) {
	timeseries := make([]*rest.Timeseries, 0)

	count, err := svc.q.ExistsThing(ctx, thing)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, ie.ErrorNotFound
	}

	ts_list, err := svc.q.FindTimeseriesByThing(ctx, thing)
	if err != nil {
		return nil, err
	} else {
		for _, item := range ts_list {
			var lBound *float64
			var uBound *float64

			if item.LowerBound.Valid {
				lBound = &item.LowerBound.Float64
			}
			if item.UpperBound.Valid {
				uBound = &item.UpperBound.Float64
			}

			t := &rest.Timeseries{
				CreatedBy:  item.CreatedBy.String(),
				LowerBound: lBound,
				Name:       item.Name,
				SiUnit:     item.SiUnit,
				Tags:       item.Tags,
				UpperBound: uBound,
				Uuid:       item.Uuid.String(),
			}

			if item.ThingUuid != NilUUID {
				v := item.ThingUuid.String()
				t.ThingUuid = &v
			}

			timeseries = append(timeseries, t)
		}
	}

	return timeseries, nil
}

func (svc *TimeseriesService) FindByUuid(ctx context.Context, id uuid.UUID) (*rest.Timeseries, error) {
	t, err := svc.q.FindTimeseriesByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	var lBound *float64
	var uBound *float64

	if t.LowerBound.Valid {
		lBound = &t.LowerBound.Float64
	}
	if t.UpperBound.Valid {
		uBound = &t.UpperBound.Float64
	}

	timeseries := &rest.Timeseries{
		Uuid:       t.Uuid.String(),
		Name:       t.Name,
		SiUnit:     t.SiUnit,
		Tags:       t.Tags,
		LowerBound: lBound,
		UpperBound: uBound,
		CreatedBy:  t.CreatedBy.String(),
	}

	if t.ThingUuid != NilUUID {
		v := t.ThingUuid.String()
		timeseries.ThingUuid = &v
	}

	return timeseries, nil
}

func (svc *TimeseriesService) FindAll(ctx context.Context, p FindAllParams) ([]*rest.Timeseries, error) {
	timeseries := make([]*rest.Timeseries, 0)

	params := pg.FindTimeseriesParams{
		Token: p.Token,
	}
	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	ts_list, err := svc.q.FindTimeseries(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, item := range ts_list {
			var lBound *float64
			var uBound *float64

			if item.LowerBound.Valid {
				lBound = &item.LowerBound.Float64
			}
			if item.UpperBound.Valid {
				uBound = &item.UpperBound.Float64
			}

			t := &rest.Timeseries{
				Uuid:       item.Uuid.String(),
				Name:       item.Name,
				SiUnit:     item.SiUnit,
				UpperBound: uBound,
				LowerBound: lBound,
				Tags:       item.Tags,
				CreatedBy:  item.CreatedBy.String(),
			}

			if item.ThingUuid != NilUUID {
				v := item.ThingUuid.String()
				t.ThingUuid = &v
			}

			timeseries = append(timeseries, t)
		}
	}

	return timeseries, nil
}

type QueryDataParams struct {
	Uuid        uuid.UUID
	Start       time.Time
	End         time.Time
	GreaterOrEq *float32
	LessOrEq    *float32
	Unit        *string
}

func (svc *TimeseriesService) QueryData(ctx context.Context, p QueryDataParams) ([]*rest.TsRow, error) {
	tsdata := make([]*rest.TsRow, 0)

	// GetTsDataRange expects a list of time series
	tsuuids := []uuid.UUID{
		p.Uuid,
	}

	var from_unit units.Unit
	var to_unit units.Unit

	if p.Unit != nil {
		ts_unit, err := svc.q.GetUnitFromTimeseries(ctx, p.Uuid)
		if err != nil {
			return nil, err
		}

		if ts_unit == *p.Unit {
			p.Unit = nil
		} else {

			to_unit, err = units.Find(*p.Unit)
			if err != nil {
				return nil, ie.ErrorInvalidUnit
			}

			from_unit, err = units.Find(ts_unit)
			if err != nil {
				// This should never error out, as there should be no incompatible units in the DB
				return nil, ie.ErrorInvalidUnit
			}
		}
	}

	params := pg.GetTsDataRangeParams{
		TsUuids: tsuuids,
		Start:   p.Start,
		Stop:    p.End,
	}

	data_list, err := svc.q.GetTsDataRange(ctx, params)
	if err != nil {
		return nil, err
	} else {
		/*
			if p.GreaterOrEq != nil {
				params.Ge = float64(*p.GreaterOrEq)
			}
			if p.LessOrEq != nil {
				params.Le = float64(*p.LessOrEq)
			}
		*/

		for _, item := range data_list {
			var f float32
			if p.Unit != nil {
				v := units.NewValue(item.Value, from_unit)
				conv, err := v.Convert(to_unit)
				if err != nil {
					return nil, ie.ErrorInvalidUnitConversion
				}
				f = float32(conv.Float())
			} else {
				f = float32(item.Value)
			}

			// Value is less than ge limit
			if p.GreaterOrEq != nil && f < *p.GreaterOrEq {
				continue
			}

			// Value is more than le limit
			if p.LessOrEq != nil && f < *p.LessOrEq {
				continue
			}

			d := rest.TsRow{
				V:  f,
				Ts: item.Ts,
			}
			tsdata = append(tsdata, &d)
		}
	}

	return tsdata, nil
}

type UpdateTimeseriesParams struct {
	Uuid       uuid.UUID
	ThingUuid  *uuid.UUID
	LowerBound *sql.NullFloat64
	UpperBound *sql.NullFloat64
	Name       *string
	SiUnit     *string
	Tags       *[]string
}

func (svc *TimeseriesService) UpdateTimeseries(ctx context.Context, p UpdateTimeseriesParams) (int64, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	var count int64

	q := svc.q.WithTx(tx)

	if p.Name != nil {
		params := pg.SetTimeseriesNameParams{
			Uuid: p.Uuid,
			Name: *p.Name,
		}
		c, err := q.SetTimeseriesName(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.SiUnit != nil {
		// FIXME: Check SI unit against Gonum

		params := pg.SetTimeseriesSiUnitParams{
			Uuid:   p.Uuid,
			SiUnit: *p.SiUnit,
		}
		c, err := q.SetTimeseriesSiUnit(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.ThingUuid != nil {
		params := pg.SetTimeseriesThingParams{
			Uuid:      p.Uuid,
			ThingUuid: *p.ThingUuid,
		}
		c, err := q.SetTimeseriesThing(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.LowerBound != nil {
		params := pg.SetTimeseriesLowerBoundParams{
			Uuid:       p.Uuid,
			LowerBound: *p.LowerBound,
		}
		c, err := q.SetTimeseriesLowerBound(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.UpperBound != nil {
		params := pg.SetTimeseriesUpperBoundParams{
			Uuid:       p.Uuid,
			UpperBound: *p.UpperBound,
		}
		c, err := q.SetTimeseriesUpperBound(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if p.Tags != nil {
		params := pg.SetTimeseriesTagsParams{
			Uuid: p.Uuid,
			Tags: *p.Tags,
		}
		c, err := q.SetTimeseriesTags(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	tx.Commit()

	return count, nil
}

func (svc *TimeseriesService) DeleteTimeseries(ctx context.Context, ts_uuid uuid.UUID) (int64, error) {
	count, err := svc.q.DeleteTimeseries(ctx, ts_uuid)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type DeleteTsDataParams struct {
	Uuid        uuid.UUID
	Start       time.Time
	End         time.Time
	GreaterOrEq *float32
	LessOrEq    *float32
}

func (svc *TimeseriesService) DeleteTsData(ctx context.Context, p DeleteTsDataParams) (int64, error) {
	// DeleteTsDataRange expects a list of time series
	tsuuids := []uuid.UUID{
		p.Uuid,
	}

	params := pg.DeleteTsDataRangeParams{
		TsUuids: tsuuids,
		Start:   p.Start,
		Stop:    p.End,
		GeNull:  p.GreaterOrEq == nil,
		LeNull:  p.LessOrEq == nil,
	}

	if p.GreaterOrEq != nil {
		params.Ge = float64(*p.GreaterOrEq)
	}
	if p.LessOrEq != nil {
		params.Le = float64(*p.LessOrEq)
	}

	count, err := svc.q.DeleteTsDataRange(ctx, params)
	if err != nil {
		return 0, err
	}

	return count, nil
}
