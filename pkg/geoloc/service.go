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

package geoloc

import (
	"context"
	"fmt"
	"net"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

type Service struct {
	pgClient *pg.Client
}

func NewService(pgClient *pg.Client) *Service {
	return &Service{pgClient: pgClient}
}

func (s *Service) LookupLocation(ctx context.Context, ip string) (coredata.IPLocationBlock, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return coredata.IPLocationBlock{}, fmt.Errorf("cannot parse IP address: %q", ip)
	}

	var location coredata.IPLocationBlock

	err := s.pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			if location, err = coredata.LookupLocationByIP(ctx, conn, ip); err != nil {
				return fmt.Errorf("cannot lookup location by IP: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return coredata.IPLocationBlock{}, err
	}

	return location, nil
}

func (s *Service) IsPopulated(ctx context.Context) (bool, error) {
	var populated bool

	err := s.pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			if populated, err = coredata.IsIPLocationBlocksPopulated(ctx, conn); err != nil {
				return fmt.Errorf("cannot check if IP location blocks are populated: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return populated, nil
}
