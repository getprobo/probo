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

	"github.com/getprobo/probo/pkg/certmanager"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
)

type (
	CustomDomainService struct {
		svc           *TenantService
		acmeService   *certmanager.ACMEService
		encryptionKey cipher.EncryptionKey
		logger        *log.Logger
	}

	CreateCustomDomainRequest struct {
		OrganizationID gid.GID
		Domain         string
	}
)

func NewCustomDomainService(
	svc *TenantService,
	acmeService *certmanager.ACMEService,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
) *CustomDomainService {
	return &CustomDomainService{
		svc:           svc,
		acmeService:   acmeService,
		encryptionKey: encryptionKey,
		logger:        logger.Named("custom_domain"),
	}
}

func (s *CustomDomainService) CreateCustomDomain(
	ctx context.Context,
	req CreateCustomDomainRequest,
) (*coredata.CustomDomain, error) {
	var domain *coredata.CustomDomain

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			domain = coredata.NewCustomDomain(req.OrganizationID, req.Domain)

			if err := domain.Insert(ctx, conn, s.svc.scope, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot insert custom domain: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return domain, nil
}

func (s *CustomDomainService) DeleteCustomDomain(
	ctx context.Context,
	domainID gid.GID,
) error {
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			domain := &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, conn, s.svc.scope, s.encryptionKey, domainID); err != nil {
				return fmt.Errorf("cannot load domain: %w", err)
			}

			if err := domain.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete domain: %w", err)
			}

			return nil
		},
	)
}

func (s *CustomDomainService) ListOrganizationDomains(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CustomDomainOrderField],
) (*page.Page[*coredata.CustomDomain, coredata.CustomDomainOrderField], error) {
	var domains coredata.CustomDomains

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := domains.LoadByOrganizationID(ctx, conn, s.svc.scope, s.encryptionKey, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot list domains: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(domains, cursor), nil
}
