// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package resourcealias

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

const aliasMaxLength = 100

type (
	Service struct {
		pg *pg.Client
	}

	CreateRequest struct {
		ResourceID gid.GID
		Alias      string
	}
)

func NewService(pgClient *pg.Client) *Service {
	return &Service{
		pg: pgClient,
	}
}

func (req *CreateRequest) Validate() error {
	v := validator.New()

	v.Check(req.ResourceID, "resource_id", validator.Required(), validator.GID())
	v.Check(req.Alias, "alias", validator.Required(), validator.Slug(aliasMaxLength))

	return v.Error()
}

func (s *Service) ResolveAlias(
	ctx context.Context,
	scope coredata.Scoper,
	alias string,
) (gid.GID, error) {
	record := &coredata.ResourceAlias{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := record.LoadByAlias(ctx, conn, scope, alias); err != nil {
				return fmt.Errorf("cannot load resource alias: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return gid.Nil, err
	}

	return record.ResourceID, nil
}

func (s *Service) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateRequest,
) (*coredata.ResourceAlias, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	alias := &coredata.ResourceAlias{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := alias.Upsert(ctx, conn, scope, req.ResourceID, req.Alias); err != nil {
				return fmt.Errorf("cannot create resource alias: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return alias, nil
}

func (s *Service) Remove(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) error {
	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			alias := &coredata.ResourceAlias{ResourceID: resourceID}
			if err := alias.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot remove resource alias: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) GetByResourceID(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) (*string, error) {
	record := &coredata.ResourceAlias{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := record.LoadByResourceID(ctx, conn, scope, resourceID); err != nil {
				return fmt.Errorf("cannot load resource alias: %w", err)
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

func (s *Service) LoadByResourceIDs(
	ctx context.Context,
	scope coredata.Scoper,
	resourceIDs []gid.GID,
) (map[gid.GID]string, error) {
	if len(resourceIDs) == 0 {
		return map[gid.GID]string{}, nil
	}

	var aliases coredata.ResourceAliases

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := aliases.LoadByResourceIDs(ctx, conn, scope, resourceIDs); err != nil {
				return fmt.Errorf("cannot load resource aliases: %w", err)
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
