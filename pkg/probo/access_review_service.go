// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
)

type (
	AccessReviewService struct {
		svc *TenantService
	}

	UpdateAccessReviewRequest struct {
		AccessReviewID   gid.GID
		IdentitySourceID **gid.GID
	}
)

// GetOrCreate returns the existing AccessReview for the organization, or creates
// one if it doesn't exist yet. Each organization has at most one AccessReview.
func (s AccessReviewService) GetOrCreate(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.AccessReview, error) {
	var review coredata.AccessReview

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			err := review.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err == nil {
				return nil
			}

			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load access review: %w", err)
			}

			now := time.Now()
			review = coredata.AccessReview{
				ID:             gid.New(s.svc.scope.GetTenantID(), coredata.AccessReviewEntityType),
				OrganizationID: organizationID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := review.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert access review: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get or create access review: %w", err)
	}

	return &review, nil
}

func (s AccessReviewService) Get(
	ctx context.Context,
	accessReviewID gid.GID,
) (*coredata.AccessReview, error) {
	review := &coredata.AccessReview{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return review.LoadByID(ctx, conn, s.svc.scope, accessReviewID)
		},
	)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func (s AccessReviewService) GetByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.AccessReview, error) {
	review := &coredata.AccessReview{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return review.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID)
		},
	)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func (s AccessReviewService) Update(
	ctx context.Context,
	req UpdateAccessReviewRequest,
) (*coredata.AccessReview, error) {
	review := &coredata.AccessReview{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := review.LoadByID(ctx, conn, s.svc.scope, req.AccessReviewID); err != nil {
				return fmt.Errorf("cannot load access review: %w", err)
			}

			if req.IdentitySourceID != nil {
				if *req.IdentitySourceID != nil {
					// Validate the source exists
					source := &coredata.AccessSource{}
					if err := source.LoadByID(ctx, conn, s.svc.scope, **req.IdentitySourceID); err != nil {
						return fmt.Errorf("cannot load identity source: %w", err)
					}
				}
				review.IdentitySourceID = *req.IdentitySourceID
			}

			review.UpdatedAt = time.Now()

			if err := review.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update access review: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return review, nil
}
