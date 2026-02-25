// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type ComplianceNewsletterSubscriberService struct {
	svc *TenantService
}

func (s *ComplianceNewsletterSubscriberService) Delete(
	ctx context.Context,
	id gid.GID,
) error {
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			subscriber := coredata.ComplianceNewsletterSubscriber{ID: id}
			return subscriber.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s *ComplianceNewsletterSubscriberService) List(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.ComplianceNewsletterSubscriberOrderField],
) (*page.Page[*coredata.ComplianceNewsletterSubscriber, coredata.ComplianceNewsletterSubscriberOrderField], error) {
	var subscribers coredata.ComplianceNewsletterSubscribers

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return subscribers.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list compliance newsletter subscribers: %w", err)
	}

	return page.NewPage(subscribers, cursor), nil
}
