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

package trust

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type TrustCenterAliasService struct {
	svc *Service
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

func (s TrustCenterAliasService) GetByStorageResourceID(
	ctx context.Context,
	scope coredata.Scoper,
	storageResourceID gid.GID,
) (*string, error) {
	record := &coredata.TrustCenterAlias{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := record.LoadByResourceID(ctx, conn, scope, storageResourceID); err != nil {
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

	return &record.Alias, nil
}

func (s TrustCenterAliasService) LoadByResourceIDs(
	ctx context.Context,
	scope coredata.Scoper,
	resourceIDs []gid.GID,
) (map[gid.GID]string, error) {
	if len(resourceIDs) == 0 {
		return map[gid.GID]string{}, nil
	}

	var aliases coredata.TrustCenterAliases

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := aliases.LoadByResourceIDs(ctx, conn, scope, resourceIDs); err != nil {
				return fmt.Errorf("cannot load trust center aliases: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	result := make(map[gid.GID]string, len(aliases))
	for _, alias := range aliases {
		result[alias.ResourceID] = alias.Alias
	}

	return result, nil
}
