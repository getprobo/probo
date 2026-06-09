// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package thirdparty

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/file"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	pg             *pg.Client
	file           *file.Service
	vetter         Vetter
	vettingEnabled bool
}

func NewService(pgClient *pg.Client, fileSvc *file.Service, vetter Vetter) *Service {
	_, disabled := vetter.(DisabledVetter)

	return &Service{
		pg:             pgClient,
		file:           fileSvc,
		vetter:         vetter,
		vettingEnabled: !disabled,
	}
}

func (s *Service) GenerateLogoURL(
	ctx context.Context,
	logoFileID gid.GID,
) (*string, error) {
	url, err := s.file.GenerateFileURL(ctx, logoFileID)
	if err != nil {
		return nil, fmt.Errorf("cannot generate logo URL: %w", err)
	}

	return &url, nil
}

func (s *Service) GetCommonThirdPartiesByIDs(
	ctx context.Context,
	ids ...gid.GID,
) (coredata.CommonThirdParties, error) {
	var parties coredata.CommonThirdParties

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := parties.LoadByIDs(ctx, conn, ids); err != nil {
				return fmt.Errorf("cannot load common third parties by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return parties, nil
}

func (s *Service) Search(ctx context.Context, name string) ([]*coredata.CommonThirdParty, error) {
	var parties coredata.CommonThirdParties

	filter := coredata.NewCommonThirdPartyFilter(&name)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return parties.LoadAll(ctx, conn, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot search common third parties: %w", err)
	}

	return parties, nil
}
