// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TrustCenterAliasService struct {
		svc *Service
	}

	CreateTrustCenterAliasRequest struct {
		ResourceID gid.GID
		Alias      string
	}

	ErrTrustCenterAliasResourceInvalid struct {
		ResourceID gid.GID
	}

	ErrTrustCenterAliasAuditReportMissing struct {
		AuditID gid.GID
	}
)

func (e ErrTrustCenterAliasResourceInvalid) Error() string {
	return fmt.Sprintf("resource %q cannot have a trust center alias", e.ResourceID)
}

func (e ErrTrustCenterAliasAuditReportMissing) Error() string {
	return fmt.Sprintf("audit %q has no report file", e.AuditID)
}

func (req *CreateTrustCenterAliasRequest) Validate() error {
	v := validator.New()

	v.Check(req.ResourceID, "resource_id", validator.Required(), validator.GID())
	v.Check(req.Alias, "alias", validator.Required(), validator.Slug(NameMaxLength))

	return v.Error()
}

func (s TrustCenterAliasService) ResolveAlias(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	alias string,
) (gid.GID, error) {
	record := &coredata.TrustCenterAlias{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := record.LoadByAlias(ctx, conn, scope, organizationID, alias); err != nil {
				return fmt.Errorf("cannot load trust center alias: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return gid.Nil, err
	}

	return record.ResourceID, nil
}

func (s TrustCenterAliasService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateTrustCenterAliasRequest,
) (*coredata.TrustCenterAlias, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	aliasResourceID, err := s.aliasResourceID(ctx, scope, req.ResourceID)
	if err != nil {
		return nil, err
	}

	alias := &coredata.TrustCenterAlias{}

	err = s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := alias.Upsert(ctx, conn, scope, aliasResourceID, req.Alias); err != nil {
				return fmt.Errorf("cannot create trust center alias: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return alias, nil
}

func (s TrustCenterAliasService) Remove(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) (gid.GID, error) {
	aliasResourceID, err := s.aliasResourceID(ctx, scope, resourceID)
	if err != nil {
		return gid.Nil, err
	}

	err = s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			alias := &coredata.TrustCenterAlias{ResourceID: aliasResourceID}
			if err := alias.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot remove trust center alias: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return gid.Nil, err
	}

	return aliasResourceID, nil
}

func (s TrustCenterAliasService) GetByResourceID(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) (*string, error) {
	aliasResourceID, err := s.aliasResourceID(ctx, scope, resourceID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, err
	}

	alias := &coredata.TrustCenterAlias{}

	err = s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := alias.LoadByResourceID(ctx, conn, scope, aliasResourceID); err != nil {
				return fmt.Errorf("cannot load trust center alias: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &alias.Alias, nil
}

func (s TrustCenterAliasService) aliasResourceID(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) (gid.GID, error) {
	switch resourceID.EntityType() {
	case coredata.DocumentEntityType, coredata.TrustCenterFileEntityType:
		return resourceID, nil

	case coredata.AuditEntityType:
		audit, err := s.svc.Audits.Get(ctx, scope, resourceID)
		if err != nil {
			return gid.Nil, err
		}

		if audit.ReportFileID == nil {
			return gid.Nil, &ErrTrustCenterAliasAuditReportMissing{AuditID: audit.ID}
		}

		return *audit.ReportFileID, nil

	default:
		return gid.Nil, &ErrTrustCenterAliasResourceInvalid{ResourceID: resourceID}
	}
}
