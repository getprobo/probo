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
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

type Service struct {
	pgClient *pg.Client
}

func NewService(pgClient *pg.Client) *Service {
	return &Service{pgClient: pgClient}
}

func (s *Service) ImportFromDir(ctx context.Context, dataDir string) error {
	countryDir := filepath.Join(dataDir, "country")

	entries, err := os.ReadDir(countryDir)
	if err != nil {
		return fmt.Errorf("cannot read country directory: %w", err)
	}

	var blocks []coredata.IPCountryBlock

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		code := strings.ToUpper(entry.Name())

		var cc coredata.CountryCode
		if err := cc.UnmarshalText([]byte(code)); err != nil {
			continue
		}

		for _, filename := range []string{"ipv4-aggregated.txt", "ipv6-aggregated.txt"} {
			path := filepath.Join(countryDir, entry.Name(), filename)

			cidrs, err := parseCIDRFile(path)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			if err != nil {
				return fmt.Errorf("cannot parse CIDR file %s: %w", path, err)
			}

			for _, cidr := range cidrs {
				blocks = append(blocks, coredata.IPCountryBlock{
					CIDR:        cidr,
					CountryCode: cc,
				})
			}
		}
	}

	if err := s.pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.CreateIPCountryBlocksStaging(ctx, conn); err != nil {
				return fmt.Errorf("cannot create staging table: %w", err)
			}

			if err := coredata.CopyIPCountryBlocksStaging(ctx, conn, blocks); err != nil {
				return fmt.Errorf("cannot copy IP country blocks to staging: %w", err)
			}

			if err := coredata.FinalizeIPCountryBlocksStaging(ctx, conn); err != nil {
				return fmt.Errorf("cannot finalize staging table: %w", err)
			}

			return nil
		},
	); err != nil {
		return err
	}

	if err := s.pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.SwapIPCountryBlocksStaging(ctx, tx); err != nil {
				return fmt.Errorf("cannot swap staging table: %w", err)
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

func (s *Service) LookupCountry(ctx context.Context, ip string) (coredata.CountryCode, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", fmt.Errorf("cannot parse IP address: %q", ip)
	}

	var cc coredata.CountryCode

	err := s.pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			if cc, err = coredata.LookupCountryByIP(ctx, conn, ip); err != nil {
				return fmt.Errorf("cannot lookup country by IP: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return cc, nil
}

func (s *Service) IsPopulated(ctx context.Context) (bool, error) {
	var populated bool

	err := s.pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			if populated, err = coredata.IsIPCountryBlocksPopulated(ctx, conn); err != nil {
				return fmt.Errorf("cannot check if IP country blocks are populated: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return populated, nil
}

func parseCIDRFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	var cidrs []string

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		_, _, err := net.ParseCIDR(line)
		if err != nil {
			return nil, fmt.Errorf("cannot parse file: %w", err)
		}

		cidrs = append(cidrs, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot scan file: %w", err)
	}

	return cidrs, nil
}
