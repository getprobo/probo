// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
		svc           *Service
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

func (s *CustomDomainService) CreateCustomDomain(
	ctx context.Context, scope coredata.Scoper,
	req CreateCustomDomainRequest,
) (*coredata.CustomDomain, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var domain *coredata.CustomDomain

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			domain = coredata.NewCustomDomain(scope.GetTenantID(), req.Domain)
			domain.OrganizationID = req.OrganizationID

			if err := domain.Insert(ctx, tx, scope, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot insert custom domain: %w", err)
			}

			var org coredata.Organization
			if err := org.LoadByID(ctx, tx, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			org.CustomDomainID = &domain.ID
			if err := org.Update(ctx, scope, tx); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var org coredata.Organization
			if err := org.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if org.CustomDomainID == nil {
				return fmt.Errorf("organization has no custom domain")
			}

			domain := &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, tx, scope, *org.CustomDomainID); err != nil {
				return fmt.Errorf("cannot load domain: %w", err)
			}

			if err := domain.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete domain: %w", err)
			}

			org.CustomDomainID = nil
			if err := org.Update(ctx, scope, tx); err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
			}

			return nil
		},
	)
}

func (s *CustomDomainService) GetOrganizationCustomDomain(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.CustomDomain, error) {
	var domain *coredata.CustomDomain

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var org coredata.Organization
			if err := org.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if org.CustomDomainID == nil {
				return nil
			}

			domain = &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, conn, scope, *org.CustomDomainID); err != nil {
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
