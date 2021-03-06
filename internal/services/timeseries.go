// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/postgres"

	units "github.com/ganehag/go-units"
)

const insertDataToTimeseries = `
INSERT INTO tsdata(ts_uuid, value, ts, created_by)
SELECT $1::uuid, x.v, x.ts, $2::uuid
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

func inValidRange(v float32, leLimit, geLimit *float32) bool {
	// leLimit: less or equal to (<=) this value
	// geLimit: greater or equal to (>=) this value
	//
	// When geLimit is more than leLimit, we have a range outside of the window
	// ge > le: ge <= x OR x <= le
	//
	// When leLimit is more than geLimit, we have a range inside of the window
	// le >= ge: ge <= x <= le

	if leLimit != nil && geLimit != nil {
		if *leLimit >= *geLimit {
			// Inside of the window
			if v >= *geLimit && v <= *leLimit {
				return true
			}
		} else {
			// Outside of the window
			if v >= *geLimit || v <= *leLimit {
				return true
			}
		}
	} else if leLimit != nil && v <= *leLimit {
		return true
	} else if geLimit != nil && v >= *geLimit {
		return true
	} else if leLimit == nil && geLimit == nil {
		return true
	}

	return false
}

// User represents the repository used for interacting with User records.
type TimeseriesService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewUser instantiates the User repository.
func NewTimeseriesService(db *sql.DB) *TimeseriesService {
	if db == nil {
		return nil
	}

	return &TimeseriesService{
		q:  postgres.New(db),
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

	params := postgres.CreateTimeseriesParams{
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
	}

	tx.Commit()

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

	var fromUnit units.Unit
	var toUnit units.Unit

	if p.Unit != nil {
		tsUnit, err := svc.q.GetUnitFromTimeseries(ctx, p.Uuid)
		if err != nil {
			return 0, err
		}

		if tsUnit == *p.Unit {
			p.Unit = nil
		} else {

			fromUnit, err = units.Find(*p.Unit)
			if err != nil {
				return 0, ie.ErrorInvalidUnit
			}

			toUnit, err = units.Find(tsUnit)
			if err != nil {
				// This should never error out, as there should be no incompatible units in the DB
				return 0, ie.ErrorInvalidUnit
			}
		}
	}

	for _, item := range p.Points {
		// Do not use a pointer to the item variable as this is a known gotcha.
		pItem := item

		if p.Unit != nil {
			v := units.NewValue(pItem.Value, fromUnit)
			conv, err := v.Convert(toUnit)
			if err != nil {
				return 0, ie.ErrorInvalidUnitConversion
			}
			pItem.Value = float64(conv.Float())
		}

		if series.LowerBound.Valid {
			// Should we skip this value
			if pItem.Value < series.LowerBound.Float64 {
				continue
			}
		}
		if series.UpperBound.Valid {
			// Should we skip this value
			if pItem.Value > series.UpperBound.Float64 {
				continue
			}
		}

		filteredPoints = append(filteredPoints, &pItem)
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

	params := postgres.FindTimeseriesByTagsParams{
		Tags:  p.Tags,
		Token: p.Token,
	}
	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	tsList, err := svc.q.FindTimeseriesByTags(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, item := range tsList {
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

	tsList, err := svc.q.FindTimeseriesByThing(ctx, thing)
	if err != nil {
		return nil, err
	}

	for _, item := range tsList {
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

	params := postgres.FindTimeseriesParams{
		Token: p.Token,
	}
	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	tsList, err := svc.q.FindTimeseries(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, item := range tsList {
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

	return timeseries, nil
}

type QuerySingleSourceDataParams struct {
	Uuid        uuid.UUID
	Start       time.Time
	End         time.Time
	GreaterOrEq *float32
	LessOrEq    *float32
	Unit        *string
	Aggregate   string
	Precision   string
	Timezone    string
}

func (svc *TimeseriesService) QuerySingleSourceData(ctx context.Context, p QuerySingleSourceDataParams) ([]*rest.TsRow, error) {
	tsdata := make([]*rest.TsRow, 0)

	var fromUnit units.Unit
	var toUnit units.Unit

	if p.Unit != nil {
		tsUnit, err := svc.q.GetUnitFromTimeseries(ctx, p.Uuid)
		if err != nil {
			return nil, err
		}

		if tsUnit == *p.Unit {
			p.Unit = nil
		} else {

			toUnit, err = units.Find(*p.Unit)
			if err != nil {
				return nil, ie.ErrorInvalidUnit
			}

			fromUnit, err = units.Find(tsUnit)
			if err != nil {
				// This should never error out, as there should be no incompatible units in the DB
				return nil, ie.ErrorInvalidUnit
			}
		}
	}

	tzloc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		return nil, err
	}

	params := postgres.GetTsDataRangeAggParams{
		Aggregate: p.Aggregate,
		Truncate:  p.Precision,
		Timezone:  p.Timezone,
		TsUuids: []uuid.UUID{
			p.Uuid, // Expects a list of time series
		},
		Start: p.Start,
		Stop:  p.End,
	}

	dataList, err := svc.q.GetTsDataRangeAgg(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, item := range dataList {
		var f float32
		if p.Unit != nil {
			v := units.NewValue(item.Value, fromUnit)
			conv, err := v.Convert(toUnit)
			if err != nil {
				return nil, ie.ErrorInvalidUnitConversion
			}
			f = float32(conv.Float())
		} else {
			f = float32(item.Value)
		}

		if inValidRange(f, p.LessOrEq, p.GreaterOrEq) == false {
			continue
		}

		d := rest.TsRow{
			V:  f,
			Ts: item.Ts.In(tzloc),
		}
		tsdata = append(tsdata, &d)
	}

	return tsdata, nil
}

type QueryMultiSourceDataParams struct {
	Uuids       []uuid.UUID
	Start       time.Time
	End         time.Time
	GreaterOrEq *float32
	LessOrEq    *float32
	Aggregate   string
	Precision   string
	Timezone    string
}

func (svc *TimeseriesService) QueryMultiSourceData(ctx context.Context, p QueryMultiSourceDataParams) ([]*rest.TsResults, error) {
	tzloc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		return nil, ie.NewInvalidRequestError(err)
	}

	params := postgres.GetTsDataRangeAggParams{
		Aggregate: p.Aggregate,
		Truncate:  p.Precision,
		Timezone:  p.Timezone,
		TsUuids:   p.Uuids,
		Start:     p.Start,
		Stop:      p.End,
	}

	mapping := make(map[uuid.UUID][]rest.TsRow, 0)
	dataList, err := svc.q.GetTsDataRangeAgg(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, item := range dataList {
		if _, ok := mapping[item.TsUuid]; ok == false {
			mapping[item.TsUuid] = make([]rest.TsRow, 0)
		}

		f := float32(item.Value)

		if inValidRange(f, p.LessOrEq, p.GreaterOrEq) == false {
			continue
		}

		mapping[item.TsUuid] = append(mapping[item.TsUuid], rest.TsRow{
			V:  f,
			Ts: item.Ts.In(tzloc),
		})
	}

	tsResult := make([]*rest.TsResults, 0)
	for key, data := range mapping {
		tsResult = append(tsResult, &rest.TsResults{
			Uuid: key.String(),
			Data: data,
		})
	}

	return tsResult, nil
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
		params := postgres.SetTimeseriesNameParams{
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

		params := postgres.SetTimeseriesSiUnitParams{
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
		params := postgres.SetTimeseriesThingParams{
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
		params := postgres.SetTimeseriesLowerBoundParams{
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
		params := postgres.SetTimeseriesUpperBoundParams{
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
		params := postgres.SetTimeseriesTagsParams{
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

func (svc *TimeseriesService) DeleteTimeseries(ctx context.Context, tsUUID uuid.UUID) (int64, error) {
	count, err := svc.q.DeleteTimeseries(ctx, tsUUID)
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

	params := postgres.DeleteTsDataRangeParams{
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
