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

// AlertService represents the repository used for interacting with Alert records.
type AlertService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewAlertService instantiates the AlertService repository.
func NewAlertService(db *sql.DB) *AlertService {
	if db == nil {
		return nil
	}

	return &AlertService{
		q:  postgres.New(db),
		db: db,
	}
}

func (svc *AlertService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := svc.q.ExistsAlert(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

type CreateAlertParams struct {
	Resource    string
	Environment string
	Event       string
	Severity    rest.AlertSeverity
	Status      rest.AlertStatus
	Value       string
	Description string
	Origin      string
	Service     []string
	Tags        []string
	Timeout     int32
	Rawdata     []byte
}

func (p *CreateAlertParams) ToPgParams() postgres.CreateAlertParams {
	return postgres.CreateAlertParams{
		Resource:    p.Resource,
		Environment: p.Environment,
		Event:       p.Event,
		Severity:    postgres.AlertSeverity(p.Severity),
		Status:      postgres.AlertStatus(p.Status),
		Value:       p.Value,
		Description: p.Description,
		Origin:      p.Origin,
		Service:     p.Service,
		Tags:        p.Tags,
		Timeout:     p.Timeout,
		Rawdata:     p.Rawdata,
	}
}

func (svc *AlertService) CreateAlert(ctx context.Context, p *CreateAlertParams) (*rest.NewAlertReply, error) {
	tags := make([]string, 0)
	if p.Tags != nil {
		for _, tag := range p.Tags {
			tags = append(tags, tag)
		}
	}

	params := p.ToPgParams()

	alert_uuid, err := svc.q.CreateAlert(ctx, params)
	if err != nil {
		return nil, err
	}

	v := &rest.NewAlertReply{
		Uuid: alert_uuid.String(),
	}

	return v, nil
}

type FindAllAlertParams struct {
	Resource    string
	Environment string
	Event       string
	Origin      string
	Status      *rest.AlertStatus
	SeverityLe  *rest.AlertSeverity
	SeverityGe  *rest.AlertSeverity
	SeverityEq  *rest.AlertSeverity
	Service     []string
	Tags        []string
	ArgOffset   int64
	ArgLimit    int64
}

func (svc *AlertService) FindAll(ctx context.Context, p FindAllAlertParams) ([]*rest.Alert, error) {
	alerts := make([]*rest.Alert, 0)

	params := postgres.FindAlertsParams{
		Resource:    p.Resource,
		Environment: p.Environment,
		Event:       p.Event,
		Origin:      p.Origin,
		Service:     p.Service,
		Tags:        p.Tags,
		ArgOffset:   p.ArgOffset,
		ArgLimit:    p.ArgLimit,
	}

	// Can't use the Enum type here as we compare to "" in SQL.
	// and "" is not one of the Enum elements.
	if p.SeverityEq != nil {
		params.SeverityEq = string(*p.SeverityEq)
	} else if p.SeverityLe != nil {
		params.SeverityLe = string(*p.SeverityLe)
	} else if p.SeverityGe != nil {
		params.SeverityGe = string(*p.SeverityGe)
	}
	if p.Status != nil {
		params.Status = string(*p.Status)
	}

	alertList, err := svc.q.FindAlerts(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, t := range alertList {
		alert := &rest.Alert{
			Uuid:             t.Uuid.String(),
			Resource:         t.Resource,
			Environment:      t.Environment,
			Event:            t.Event,
			Severity:         rest.AlertSeverity(t.Severity),
			Status:           rest.AlertStatus(t.Status),
			Service:          t.Service,
			Value:            t.Value,
			Description:      t.Description,
			Origin:           t.Origin,
			Tags:             t.Tags,
			Created:          t.Created,
			Timeout:          t.Timeout,
			Rawdata:          t.Rawdata,
			Duplicate:        t.Duplicate,
			PreviousSeverity: rest.AlertSeverity(t.PreviousSeverity),
		}

		if t.LastReceiveTime.Valid == true {
			alert.LastReceiveTime = &t.LastReceiveTime.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (svc *AlertService) FindAlertByUuid(ctx context.Context, id uuid.UUID) (*rest.Alert, error) {
	alert, err := svc.q.FindAlertByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	v := &rest.Alert{
		Uuid:             alert.Uuid.String(),
		Resource:         alert.Resource,
		Environment:      alert.Environment,
		Event:            alert.Event,
		Severity:         rest.AlertSeverity(alert.Severity),
		Status:           rest.AlertStatus(alert.Status),
		Service:          alert.Service,
		Value:            alert.Value,
		Description:      alert.Description,
		Origin:           alert.Origin,
		Tags:             alert.Tags,
		Created:          alert.Created,
		Timeout:          alert.Timeout,
		Rawdata:          alert.Rawdata,
		Duplicate:        alert.Duplicate,
		PreviousSeverity: rest.AlertSeverity(alert.PreviousSeverity),
	}

	if alert.LastReceiveTime.Valid == true {
		v.LastReceiveTime = &alert.LastReceiveTime.Time
	}

	return v, nil
}

type UpdateAlertByUuidParams struct {
	Resource    *string
	Environment *string
	Event       *string
	Severity    *rest.AlertSeverity
	Status      *rest.AlertStatus
	Value       *string
	Description *string
	Origin      *string
	Service     *[]string
	Tags        *[]string
	Timeout     *int32
	Rawdata     *[]byte
}

func (svc *AlertService) UpdateAlertByUuid(ctx context.Context, id uuid.UUID, p UpdateAlertByUuidParams) (int64, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	var count int64

	q := svc.q.WithTx(tx)

	if p.Resource != nil {
		c, err := q.UpdateAlertSetResource(ctx, postgres.UpdateAlertSetResourceParams{
			Uuid:     id,
			Resource: *p.Resource,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Environment != nil {
		c, err := q.UpdateAlertSetEnvironment(ctx, postgres.UpdateAlertSetEnvironmentParams{
			Uuid:        id,
			Environment: *p.Environment,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Event != nil {
		c, err := q.UpdateAlertSetEvent(ctx, postgres.UpdateAlertSetEventParams{
			Uuid:  id,
			Event: *p.Event,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Origin != nil {
		c, err := q.UpdateAlertSetOrigin(ctx, postgres.UpdateAlertSetOriginParams{
			Uuid:   id,
			Origin: *p.Origin,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Severity != nil {
		c, err := q.UpdateAlertSetSeverity(ctx, postgres.UpdateAlertSetSeverityParams{
			Uuid:     id,
			Severity: postgres.AlertSeverity(*p.Severity),
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Status != nil {
		c, err := q.UpdateAlertSetStatus(ctx, postgres.UpdateAlertSetStatusParams{
			Uuid:   id,
			Status: postgres.AlertStatus(*p.Status),
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Service != nil {
		c, err := q.UpdateAlertSetService(ctx, postgres.UpdateAlertSetServiceParams{
			Uuid:    id,
			Service: *p.Service,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Value != nil {
		c, err := q.UpdateAlertSetValue(ctx, postgres.UpdateAlertSetValueParams{
			Uuid:  id,
			Value: *p.Value,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Description != nil {
		c, err := q.UpdateAlertSetDescription(ctx, postgres.UpdateAlertSetDescriptionParams{
			Uuid:        id,
			Description: *p.Description,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Tags != nil {
		c, err := q.UpdateAlertSetTags(ctx, postgres.UpdateAlertSetTagsParams{
			Uuid: id,
			Tags: *p.Tags,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Timeout != nil {
		c, err := q.UpdateAlertSetTimeout(ctx, postgres.UpdateAlertSetTimeoutParams{
			Uuid:    id,
			Timeout: *p.Timeout,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Rawdata != nil {
		c, err := q.UpdateAlertSetRawdata(ctx, postgres.UpdateAlertSetRawdataParams{
			Uuid:    id,
			Rawdata: *p.Rawdata,
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

func (svc *AlertService) DeleteAlert(ctx context.Context, id uuid.UUID) (int64, error) {
	count, err := svc.q.DeleteAlert(ctx, id)
	if err != nil {
		return 0, err
	}

	return count, nil
}
