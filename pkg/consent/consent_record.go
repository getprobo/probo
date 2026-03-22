// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package consent

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) ListConsentRecordsForCookieBannerID(
	ctx context.Context,
	cookieBannerID gid.GID,
	cursor *page.Cursor[coredata.ConsentRecordOrderField],
	filter *coredata.ConsentRecordFilter,
) (*page.Page[*coredata.ConsentRecord, coredata.ConsentRecordOrderField], error) {
	scope := coredata.NewScopeFromObjectID(cookieBannerID)
	var records coredata.ConsentRecords

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := records.LoadByCookieBannerID(
				ctx,
				conn,
				scope,
				cookieBannerID,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load consent records: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list consent records: %w", err)
	}

	return page.NewPage(records, cursor), nil
}

func (s *Service) CountConsentRecordsForCookieBannerID(
	ctx context.Context,
	cookieBannerID gid.GID,
	filter *coredata.ConsentRecordFilter,
) (int, error) {
	var (
		scope = coredata.NewScopeFromObjectID(cookieBannerID)
		count int
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			records := coredata.ConsentRecords{}
			count, err = records.CountByCookieBannerID(ctx, conn, scope, cookieBannerID, filter)
			if err != nil {
				return fmt.Errorf("cannot count consent records: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
