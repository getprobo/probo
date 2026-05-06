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

package geoloc

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
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

	type row struct {
		cidr        string
		countryCode coredata.CountryCode
	}

	var rows []row

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		code := strings.ToUpper(entry.Name())
		var cc coredata.CountryCode
		if err := cc.Scan(code); err != nil {
			continue
		}

		for _, filename := range []string{"ipv4-aggregated.txt", "ipv6-aggregated.txt"} {
			path := filepath.Join(countryDir, entry.Name(), filename)

			cidrs, err := parseCIDRFile(path)
			if err != nil {
				continue
			}

			for _, cidr := range cidrs {
				rows = append(rows, row{cidr: cidr, countryCode: cc})
			}
		}
	}

	return s.pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, "TRUNCATE common_ip_country_blocks")
			if err != nil {
				return fmt.Errorf("cannot truncate common_ip_country_blocks: %w", err)
			}

			pgxRows := make([][]any, len(rows))
			for i, r := range rows {
				pgxRows[i] = []any{r.cidr, r.countryCode.String()}
			}

			_, err = tx.CopyFrom(
				ctx,
				pgx.Identifier{"common_ip_country_blocks"},
				[]string{"cidr", "country_code"},
				pgx.CopyFromRows(pgxRows),
			)
			if err != nil {
				return fmt.Errorf("cannot copy rows into common_ip_country_blocks: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) LookupCountry(ctx context.Context, conn pg.Querier, ip string) (coredata.CountryCode, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", fmt.Errorf("cannot parse IP address: %q", ip)
	}

	q := `
SELECT country_code
FROM common_ip_country_blocks
WHERE cidr >>= @ip::inet
ORDER BY masklen(cidr) DESC
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"ip": ip}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return "", fmt.Errorf("cannot query ip country blocks: %w", err)
	}

	cc, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[coredata.CountryCode])
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("cannot collect ip country block row: %w", err)
	}

	return cc, nil
}

func (s *Service) IsPopulated(ctx context.Context, conn pg.Querier) (bool, error) {
	q := `SELECT EXISTS (SELECT 1 FROM common_ip_country_blocks);`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return false, fmt.Errorf("cannot check if ip country blocks is populated: %w", err)
	}

	populated, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect populated check: %w", err)
	}

	return populated, nil
}

func parseCIDRFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cidrs []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		_, _, err := net.ParseCIDR(line)
		if err != nil {
			continue
		}

		cidrs = append(cidrs, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot scan file: %w", err)
	}

	return cidrs, nil
}
