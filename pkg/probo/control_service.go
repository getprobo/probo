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

// Create creates a new control
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

// Get retrieves a control by ID
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

// Update updates an existing control
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

// Delete removes a control
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

// ListForFrameworkID retrieves all controls for a framework
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

func (s ControlService) ConnectToMitigation(
	ctx context.Context,
	req ConnectControlToMitigationRequest,
) error {
	now := time.Now()

	controlMitigation := &coredata.ControlMitigation{
		ControlID:    req.ControlID,
		MitigationID: req.MitigationID,
		TenantID:     s.svc.scope.GetTenantID(),
		CreatedAt:    now,
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controlMitigation.Insert(ctx, conn, s.svc.scope)
		},
	)
}

// DisconnectFromMitigation removes the link between a control and a mitigation
func (s ControlService) DisconnectFromMitigation(
	ctx context.Context,
	req DisconnectControlFromMitigationRequest,
) error {
	controlMitigation := &coredata.ControlMitigation{
		ControlID:    req.ControlID,
		MitigationID: req.MitigationID,
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return controlMitigation.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s ControlService) ListMitigationsForControlID(
	ctx context.Context,
	controlID gid.GID,
) ([]*coredata.Mitigation, error) {
	var controlMitigations coredata.ControlMitigations
	var mitigations []*coredata.Mitigation

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := controlMitigations.LoadByControlID(ctx, conn, s.svc.scope, controlID); err != nil {
				return fmt.Errorf("cannot load control mitigations: %w", err)
			}

			for _, cm := range controlMitigations {
				mitigation := &coredata.Mitigation{}
				if err := mitigation.LoadByID(ctx, conn, s.svc.scope, cm.MitigationID); err != nil {
					return fmt.Errorf("cannot load mitigation: %w", err)
				}
				mitigations = append(mitigations, mitigation)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return mitigations, nil
}

// ListControlsForMitigationID retrieves all controls linked to a mitigation
func (s ControlService) ListControlsForMitigationID(
	ctx context.Context,
	mitigationID gid.GID,
) ([]*coredata.Control, error) {
	var controlMitigations coredata.ControlMitigations
	var controls []*coredata.Control

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := controlMitigations.LoadByMitigationID(ctx, conn, s.svc.scope, mitigationID); err != nil {
				return fmt.Errorf("cannot load control mitigations: %w", err)
			}

			for _, cm := range controlMitigations {
				control := &coredata.Control{}
				if err := control.LoadByID(ctx, conn, s.svc.scope, cm.ControlID); err != nil {
					return fmt.Errorf("cannot load control: %w", err)
				}
				controls = append(controls, control)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return controls, nil
}
