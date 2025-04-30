// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package probo

import (
	"context"
	"fmt"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/kit/pg"
)

type (
	ControlService struct {
		svc *TenantService
	}

	CreateControlRequest struct {
		ID          gid.GID
		FrameworkID gid.GID
		Name        string
		Description string
	}

	UpdateControlRequest struct {
		ID              gid.GID
		ExpectedVersion int
		Name            *string
		Description     *string
	}

	ConnectControlToMitigationRequest struct {
		ControlID    gid.GID
		MitigationID gid.GID
	}

	DisconnectControlFromMitigationRequest struct {
		ControlID    gid.GID
		MitigationID gid.GID
	}
)

func (s ControlService) ListForPolicyID(
	ctx context.Context,
	policyID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controls.LoadByPolicyID(ctx, conn, s.svc.scope, policyID, cursor)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) ListForMeasureID(
	ctx context.Context,
	measureID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controls.LoadByMeasureID(ctx, conn, s.svc.scope, measureID, cursor)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) CreateMeasureMapping(
	ctx context.Context,
	controlID gid.GID,
	measureID gid.GID,
) error {
	controlMeasure := &coredata.ControlMeasure{
		ControlID: controlID,
		MeasureID: measureID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controlMeasure.Insert(ctx, conn, s.svc.scope)
		},
	)
}

func (s ControlService) DeleteMeasureMapping(
	ctx context.Context,
	controlID gid.GID,
	measureID gid.GID,
) error {
	controlMeasure := &coredata.ControlMeasure{
		ControlID: controlID,
		MeasureID: measureID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controlMeasure.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s ControlService) CreatePolicyMapping(
	ctx context.Context,
	controlID gid.GID,
	policyID gid.GID,
) error {
	controlPolicy := &coredata.ControlPolicy{
		ControlID: controlID,
		PolicyID:  policyID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controlPolicy.Insert(ctx, conn, s.svc.scope)
		},
	)
}

func (s ControlService) DeletePolicyMapping(
	ctx context.Context,
	controlID gid.GID,
	policyID gid.GID,
) error {
	controlPolicy := &coredata.ControlPolicy{
		ControlID: controlID,
		PolicyID:  policyID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controlPolicy.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s ControlService) Create(
	ctx context.Context,
	req CreateControlRequest,
) (*coredata.Control, error) {
	now := time.Now()

	control := &coredata.Control{
		ID:          req.ID,
		FrameworkID: req.FrameworkID,
		TenantID:    s.svc.scope.GetTenantID(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return control.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot create control: %w", err)
	}

	return control, nil
}

func (s ControlService) Get(
	ctx context.Context,
	controlID gid.GID,
) (*coredata.Control, error) {
	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return control.LoadByID(ctx, conn, s.svc.scope, controlID)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot get control: %w", err)
	}

	return control, nil
}

func (s ControlService) Update(
	ctx context.Context,
	req UpdateControlRequest,
) (*coredata.Control, error) {
	params := coredata.UpdateControlParams{
		ExpectedVersion: req.ExpectedVersion,
		Name:            req.Name,
		Description:     req.Description,
	}

	control := &coredata.Control{ID: req.ID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return control.Update(ctx, conn, s.svc.scope, params)
		})
	if err != nil {
		return nil, fmt.Errorf("cannot update control: %w", err)
	}

	return control, nil
}

func (s ControlService) Delete(
	ctx context.Context,
	controlID gid.GID,
) error {
	control := &coredata.Control{ID: controlID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return control.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s ControlService) ListForFrameworkID(
	ctx context.Context,
	frameworkID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controls.LoadByFrameworkID(
				ctx,
				conn,
				s.svc.scope,
				frameworkID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) ListForRiskID(
	ctx context.Context,
	riskID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controls.LoadByRiskID(ctx, conn, s.svc.scope, riskID, cursor)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}
