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

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
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

func (ccdr *CreateCustomDomainRequest) Validate() error {
	v := validator.New()

	v.Check(ccdr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(ccdr.Domain, "domain", validator.Required(), validator.NotEmpty(), validator.Domain())

	return v.Error()
}

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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var domain *coredata.CustomDomain

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			domain = coredata.NewCustomDomain(s.svc.scope.GetTenantID(), req.Domain)
			domain.OrganizationID = req.OrganizationID

			if err := domain.Insert(ctx, tx, s.svc.scope, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot insert custom domain: %w", err)
			}

			var org coredata.Organization
			if err := org.LoadByID(ctx, tx, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			org.CustomDomainID = &domain.ID
			if err := org.Update(ctx, s.svc.scope, tx); err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
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
	organizationID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var org coredata.Organization
			if err := org.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if org.CustomDomainID == nil {
				return fmt.Errorf("organization has no custom domain")
			}

			domain := &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, tx, s.svc.scope, s.encryptionKey, *org.CustomDomainID); err != nil {
				return fmt.Errorf("cannot load domain: %w", err)
			}

			if err := domain.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete domain: %w", err)
			}

			org.CustomDomainID = nil
			if err := org.Update(ctx, s.svc.scope, tx); err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
			}

			return nil
		},
	)
}

func (s *CustomDomainService) GetOrganizationCustomDomain(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.CustomDomain, error) {
	var domain *coredata.CustomDomain

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var org coredata.Organization
			if err := org.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if org.CustomDomainID == nil {
				return nil
			}

			domain = &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, conn, s.svc.scope, s.encryptionKey, *org.CustomDomainID); err != nil {
				return fmt.Errorf("cannot load custom domain: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return domain, nil
}
