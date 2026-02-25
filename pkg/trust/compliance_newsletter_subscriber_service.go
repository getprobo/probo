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

package trust

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type ComplianceNewsletterSubscriberService struct {
	svc *TenantService
}

func (s *ComplianceNewsletterSubscriberService) IsSubscribed(
	ctx context.Context,
	trustCenterID gid.GID,
	email mail.Addr,
) (bool, error) {
	var subscriber coredata.ComplianceNewsletterSubscriber

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return subscriber.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("cannot check newsletter subscription: %w", err)
	}

	return true, nil
}

func (s *ComplianceNewsletterSubscriberService) Subscribe(
	ctx context.Context,
	trustCenterID gid.GID,
	email mail.Addr,
) (*coredata.ComplianceNewsletterSubscriber, error) {
	now := time.Now()

	trustCenter := &coredata.TrustCenter{}
	var subscriber coredata.ComplianceNewsletterSubscriber

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			existing := &coredata.ComplianceNewsletterSubscriber{}
			err := existing.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
			if err == nil {
				subscriber = *existing
				return nil
			}

			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot check existing subscription: %w", err)
			}

			subscriber = coredata.ComplianceNewsletterSubscriber{
				ID:             gid.New(trustCenter.TenantID, coredata.ComplianceNewsletterSubscriberEntityType),
				TenantID:       trustCenter.TenantID,
				OrganizationID: trustCenter.OrganizationID,
				TrustCenterID:  trustCenterID,
				Email:          email,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return subscriber.Insert(ctx, conn, s.svc.scope)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot subscribe to newsletter: %w", err)
	}

	return &subscriber, nil
}

func (s *ComplianceNewsletterSubscriberService) Unsubscribe(
	ctx context.Context,
	trustCenterID gid.GID,
	email mail.Addr,
) error {
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			subscriber := &coredata.ComplianceNewsletterSubscriber{}

			if err := subscriber.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load subscription: %w", err)
			}

			return subscriber.Delete(ctx, conn, s.svc.scope)
		},
	)
}
