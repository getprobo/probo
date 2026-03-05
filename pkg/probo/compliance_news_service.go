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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

var ErrComplianceNewsAlreadySent = errors.New("compliance news already sent")

type ComplianceNewsService struct {
	svc *TenantService
}

func (s *ComplianceNewsService) Create(
	ctx context.Context,
	trustCenterID gid.GID,
	title string,
	body string,
) (*coredata.ComplianceNews, error) {
	now := time.Now()

	cn := &coredata.ComplianceNews{
		ID:            gid.New(s.svc.scope.GetTenantID(), coredata.ComplianceNewsEntityType),
		TrustCenterID: trustCenterID,
		Title:         title,
		Body:          body,
		Status:        coredata.ComplianceNewsStatusDraft,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var trustCenter coredata.TrustCenter
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			cn.OrganizationID = trustCenter.OrganizationID

			if err := cn.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert compliance news: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

func (s *ComplianceNewsService) Update(
	ctx context.Context,
	id gid.GID,
	title string,
	body string,
	status coredata.ComplianceNewsStatus,
) (*coredata.ComplianceNews, error) {
	var cn coredata.ComplianceNews

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := cn.LoadByID(ctx, conn, s.svc.scope, id); err != nil {
				return fmt.Errorf("cannot load compliance news: %w", err)
			}

			if cn.Status == coredata.ComplianceNewsStatusSent {
				return ErrComplianceNewsAlreadySent
			}

			cn.Title = title
			cn.Body = body
			cn.Status = status
			cn.UpdatedAt = time.Now()

			if err := cn.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update compliance news: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &cn, nil
}

func (s *ComplianceNewsService) Delete(
	ctx context.Context,
	id gid.GID,
) error {
	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			cn := coredata.ComplianceNews{ID: id}
			if err := cn.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete compliance news: %w", err)
			}

			return nil
		},
	)
	return err
}

func (s *ComplianceNewsService) Get(
	ctx context.Context,
	id gid.GID,
) (*coredata.ComplianceNews, error) {
	var cn coredata.ComplianceNews

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := cn.LoadByID(ctx, conn, s.svc.scope, id); err != nil {
				return fmt.Errorf("cannot load compliance news: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &cn, nil
}

func (s *ComplianceNewsService) List(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.ComplianceNewsOrderField],
) (*page.Page[*coredata.ComplianceNews, coredata.ComplianceNewsOrderField], error) {
	var items coredata.ComplianceNewsItems

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := items.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor); err != nil {
				return fmt.Errorf("cannot load compliance news list: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(items, cursor), nil
}

func (s *ComplianceNewsService) Count(
	ctx context.Context,
	trustCenterID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var items coredata.ComplianceNewsItems
			var err error
			count, err = items.CountByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID)
			if err != nil {
				return fmt.Errorf("cannot count compliance news: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
