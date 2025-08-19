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

type SnapshotService struct {
	svc *TenantService
}

type (
	CreateSnapshotRequest struct {
		OrganizationID gid.GID
		Name           string
		Description    *string
		Type           coredata.SnapshotsType
	}

	UpdateSnapshotRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
		Type        *coredata.SnapshotsType
	}
)

func (s *SnapshotService) Get(
	ctx context.Context,
	snapshotID gid.GID,
) (*coredata.Snapshot, error) {
	snapshot := &coredata.Snapshot{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return snapshot.LoadByID(ctx, conn, s.svc.scope, snapshotID)
		},
	)

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (s *SnapshotService) Create(
	ctx context.Context,
	req *CreateSnapshotRequest,
) (*coredata.Snapshot, error) {
	now := time.Now()

	snapshot := &coredata.Snapshot{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.SnapshotEntityType),
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		CreatedAt:      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := snapshot.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert snapshot: %w", err)
			}

			snapshottable, err := coredata.GetSnapshottable(req.Type)
			if err != nil {
				return err
			}

			if err := snapshottable.Snapshot(ctx, conn, s.svc.scope, req.OrganizationID, snapshot.ID); err != nil {
				return fmt.Errorf("cannot create snapshot: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (s *SnapshotService) Delete(
	ctx context.Context,
	snapshotID gid.GID,
) error {
	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			snapshot := &coredata.Snapshot{}
			if err := snapshot.LoadByID(ctx, conn, s.svc.scope, snapshotID); err != nil {
				return fmt.Errorf("cannot load snapshot: %w", err)
			}

			if err := snapshot.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete snapshot: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s *SnapshotService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.SnapshotOrderField],
) (*page.Page[*coredata.Snapshot, coredata.SnapshotOrderField], error) {
	snapshots := coredata.Snapshots{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := snapshots.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot load snapshots: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(snapshots, cursor), nil
}

func (s *SnapshotService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			snapshots := coredata.Snapshots{}
			count, err = snapshots.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count snapshots: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
