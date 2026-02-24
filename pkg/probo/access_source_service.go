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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	AccessSourceService struct {
		svc *TenantService
	}

	CreateAccessSourceRequest struct {
		AccessReviewID gid.GID
		ConnectorID    *gid.GID
		Name           string
		Category       coredata.AccessSourceCategory
		CsvData        *string
	}

	UpdateAccessSourceRequest struct {
		AccessSourceID gid.GID
		Name           *string
		Category       *coredata.AccessSourceCategory
		ConnectorID    **gid.GID
		CsvData        **string
	}
)

func (r *CreateAccessSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.AccessReviewID, "access_review_id", validator.Required(), validator.GID(coredata.AccessReviewEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Category, "category", validator.OneOfSlice(coredata.AccessSourceCategories()))

	return v.Error()
}

func (r *UpdateAccessSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.AccessSourceID, "access_source_id", validator.Required(), validator.GID(coredata.AccessSourceEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Category, "category", validator.OneOfSlice(coredata.AccessSourceCategories()))

	return v.Error()
}

func (s AccessSourceService) Create(
	ctx context.Context,
	req CreateAccessSourceRequest,
) (*coredata.AccessSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	source := &coredata.AccessSource{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.AccessSourceEntityType),
		AccessReviewID: req.AccessReviewID,
		ConnectorID:    req.ConnectorID,
		Name:           req.Name,
		Category:       req.Category,
		CsvData:        req.CsvData,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			// Validate parent access review exists
			review := &coredata.AccessReview{}
			if err := review.LoadByID(ctx, conn, s.svc.scope, req.AccessReviewID); err != nil {
				return fmt.Errorf("cannot load access review: %w", err)
			}

			// Validate connector exists if provided
			if req.ConnectorID != nil {
				connector := &coredata.Connector{}
				if err := connector.LoadMetadataByID(ctx, conn, s.svc.scope, *req.ConnectorID); err != nil {
					return fmt.Errorf("cannot load connector: %w", err)
				}
			}

			if err := source.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert access source: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create access source: %w", err)
	}

	return source, nil
}

func (s AccessSourceService) Get(
	ctx context.Context,
	accessSourceID gid.GID,
) (*coredata.AccessSource, error) {
	source := &coredata.AccessSource{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return source.LoadByID(ctx, conn, s.svc.scope, accessSourceID)
		},
	)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (s AccessSourceService) Update(
	ctx context.Context,
	req UpdateAccessSourceRequest,
) (*coredata.AccessSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	source := &coredata.AccessSource{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := source.LoadByID(ctx, conn, s.svc.scope, req.AccessSourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

			if req.Name != nil {
				source.Name = *req.Name
			}

			if req.Category != nil {
				source.Category = *req.Category
			}

			if req.ConnectorID != nil {
				if *req.ConnectorID != nil {
					connector := &coredata.Connector{}
					if err := connector.LoadMetadataByID(ctx, conn, s.svc.scope, **req.ConnectorID); err != nil {
						return fmt.Errorf("cannot load connector: %w", err)
					}
				}
				source.ConnectorID = *req.ConnectorID
			}

			if req.CsvData != nil {
				source.CsvData = *req.CsvData
			}

			source.UpdatedAt = time.Now()

			if err := source.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update access source: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (s AccessSourceService) Delete(
	ctx context.Context,
	accessSourceID gid.GID,
) error {
	source := &coredata.AccessSource{ID: accessSourceID}

	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return source.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s AccessSourceService) ListForAccessReviewID(
	ctx context.Context,
	accessReviewID gid.GID,
	cursor *page.Cursor[coredata.AccessSourceOrderField],
) (*page.Page[*coredata.AccessSource, coredata.AccessSourceOrderField], error) {
	var sources coredata.AccessSources

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return sources.LoadByAccessReviewID(ctx, conn, s.svc.scope, accessReviewID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(sources, cursor), nil
}

func (s AccessSourceService) CountForAccessReviewID(
	ctx context.Context,
	accessReviewID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			sources := coredata.AccessSources{}
			count, err = sources.CountByAccessReviewID(ctx, conn, s.svc.scope, accessReviewID)
			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s AccessSourceService) ListScopeSourcesForCampaignID(
	ctx context.Context,
	campaignID gid.GID,
) ([]*coredata.AccessSource, error) {
	var sources coredata.AccessSources

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return sources.LoadScopeSourcesByCampaignID(ctx, conn, s.svc.scope, campaignID)
		},
	)
	if err != nil {
		return nil, err
	}

	return sources, nil
}
