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

package accessreview

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) GetInventoryChangeEvent(
	ctx context.Context,
	scope coredata.Scoper,
	eventID gid.GID,
) (*coredata.AccessInventoryChangeEvent, error) {
	event := &coredata.AccessInventoryChangeEvent{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := event.LoadByID(ctx, conn, scope, eventID); err != nil {
				return fmt.Errorf("cannot load access inventory change event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *Service) ListInventoryChangeEventsForGrantID(
	ctx context.Context,
	scope coredata.Scoper,
	grantID gid.GID,
	cursor *page.Cursor[coredata.AccessInventoryChangeEventOrderField],
	filter *coredata.AccessInventoryChangeEventFilter,
) (*page.Page[*coredata.AccessInventoryChangeEvent, coredata.AccessInventoryChangeEventOrderField], error) {
	var events coredata.AccessInventoryChangeEvents

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := events.LoadByGrantID(ctx, conn, scope, grantID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load access inventory change events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(events, cursor), nil
}

func (s *Service) CountInventoryChangeEventsForGrantID(
	ctx context.Context,
	scope coredata.Scoper,
	grantID gid.GID,
	filter *coredata.AccessInventoryChangeEventFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			events := coredata.AccessInventoryChangeEvents{}

			var err error

			count, err = events.CountByGrantID(ctx, conn, scope, grantID, filter)
			if err != nil {
				return fmt.Errorf("cannot count access inventory change events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListInventoryChangeEventsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessInventoryChangeEventOrderField],
	filter *coredata.AccessInventoryChangeEventFilter,
) (*page.Page[*coredata.AccessInventoryChangeEvent, coredata.AccessInventoryChangeEventOrderField], error) {
	var events coredata.AccessInventoryChangeEvents

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := events.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load access inventory change events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(events, cursor), nil
}

func (s *Service) CountInventoryChangeEventsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.AccessInventoryChangeEventFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			events := coredata.AccessInventoryChangeEvents{}

			var err error

			count, err = events.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count access inventory change events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
