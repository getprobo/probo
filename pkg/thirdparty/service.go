// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package thirdparty

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	pg             *pg.Client
	file           *filemanager.Service
	vetter         Vetter
	vettingEnabled bool
}

func NewService(pgClient *pg.Client, fileSvc *filemanager.Service, vetter Vetter) *Service {
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
	file, err := s.file.GetPublicFile(ctx, logoFileID)
	if err != nil {
		return nil, fmt.Errorf("cannot load logo file: %w", err)
	}

	url := s.file.GenerateFileURL(file)

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
