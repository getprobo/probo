// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package management

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

var ErrCustomDomainSlotTaken = errors.New("compliance page already has a custom domain")

func (s *Service) AddCustomDomain(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	domain string,
) (*coredata.CustomDomain, error) {
	v := validator.New()
	v.Check(compliancePageID, "compliance_page_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(domain, "domain", validator.Required(), validator.NotEmpty(), validator.Domain())
	if err := v.Error(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var customDomain *coredata.CustomDomain

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, tx, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.CustomDomainID != nil {
				return ErrCustomDomainSlotTaken
			}

			certificate, err := s.certManager.EnsureCertificate(ctx, tx, scope, domain)
			if err != nil {
				return fmt.Errorf("cannot ensure certificate: %w", err)
			}

			customDomain = coredata.NewCustomDomain(
				scope.GetTenantID(),
				compliancePage.OrganizationID,
				domain,
				false,
			)
			customDomain.CertificateID = &certificate.ID

			if err := customDomain.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert custom domain: %w", err)
			}

			compliancePage.CustomDomainID = &customDomain.ID
			compliancePage.UpdatedAt = time.Now()

			if err := compliancePage.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update compliance page: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return customDomain, nil
}

func (s *Service) RemoveCustomDomain(
	ctx context.Context,
	scope coredata.Scoper,
	customDomainID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			domain := &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, tx, scope, customDomainID); err != nil {
				return fmt.Errorf("cannot load custom domain: %w", err)
			}

			if domain.Managed {
				return ErrCustomDomainManaged
			}

			compliancePage := &coredata.TrustCenter{}
			err := compliancePage.LoadByDomainID(ctx, tx, customDomainID)
			switch {
			case err == nil:
				if compliancePage.CustomDomainID != nil && *compliancePage.CustomDomainID == customDomainID {
					compliancePage.CustomDomainID = nil
					compliancePage.UpdatedAt = time.Now()

					if err := compliancePage.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot update compliance page: %w", err)
					}
				}
			case errors.Is(err, coredata.ErrResourceNotFound):
			default:
				return fmt.Errorf("cannot load compliance page by domain id: %w", err)
			}

			if err := domain.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete custom domain: %w", err)
			}

			if domain.CertificateID != nil {
				if err := s.certManager.Delete(ctx, tx, scope, *domain.CertificateID); err != nil {
					return fmt.Errorf("cannot delete certificate: %w", err)
				}
			}

			return nil
		},
	)
}

// GetDefaultDomain returns the compliance page's default probopage subdomain,
// or nil when it has not been provisioned yet.
func (s *Service) GetDefaultDomain(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (*coredata.CustomDomain, error) {
	var domain *coredata.CustomDomain

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.DefaultDomainID == nil {
				return nil
			}

			domain = &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, conn, scope, *compliancePage.DefaultDomainID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					domain = nil
					return nil
				}

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

func (s *Service) GetCustomDomain(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (*coredata.CustomDomain, error) {
	var domain *coredata.CustomDomain

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.CustomDomainID == nil {
				return nil
			}

			domain = &coredata.CustomDomain{}
			if err := domain.LoadByID(ctx, conn, scope, *compliancePage.CustomDomainID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					domain = nil
					return nil
				}

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

// PublicURL returns the canonical public URL of a compliance page on its
// dedicated domain.
func (s *Service) PublicURL(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (string, error) {
	var publicURL string

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			var err error
			publicURL, err = s.PublicURLForCompliancePage(ctx, conn, scope, compliancePage)
			if err != nil {
				return fmt.Errorf("cannot resolve public url: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return publicURL, nil
}
