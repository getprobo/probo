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
	"go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

// The compliance portal service owns the relationship between a compliance
// page (trust center) and its domains. A page has two slots stored on the
// trust center row: a default {slug}.probopage.com domain provided by Probo and
// an optional custom domain. It provisions each domain's TLS certificate
// through the generic certmanager service within the trust center's transaction
// so slot changes stay atomic with the page.

// ErrCustomDomainSlotTaken is returned when a compliance page already has a
// custom domain and another one is added.
var ErrCustomDomainSlotTaken = errors.New("compliance page already has a custom domain")

// AddCustomDomain provisions the compliance page's custom domain. It fails
// when the page already has one. The default probopage subdomain, provisioned
// at page creation, keeps serving as a fallback while the new certificate
// provisions.
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
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.CustomDomainID != nil {
				return ErrCustomDomainSlotTaken
			}

			certificate, err := s.certManager.EnsureCertificate(ctx, tx, scope, domain)
			if err != nil {
				return fmt.Errorf("cannot ensure certificate: %w", err)
			}

			customDomain = coredata.NewCustomDomain(
				scope.GetTenantID(),
				trustCenter.OrganizationID,
				domain,
				false,
			)
			customDomain.CertificateID = &certificate.ID

			if err := customDomain.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert custom domain: %w", err)
			}

			trustCenter.CustomDomainID = &customDomain.ID
			trustCenter.UpdatedAt = time.Now()

			if err := trustCenter.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return customDomain, nil
}

// RemoveCustomDomain clears the compliance page's custom domain and deletes
// the underlying domain together with its certificate. The default domain
// cannot be removed.
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
				return complianceportal.ErrCustomDomainManaged
			}

			trustCenter := &coredata.TrustCenter{}
			err := trustCenter.LoadByDomainID(ctx, tx, customDomainID)
			switch {
			case err == nil:
				if trustCenter.CustomDomainID != nil && *trustCenter.CustomDomainID == customDomainID {
					trustCenter.CustomDomainID = nil
					trustCenter.UpdatedAt = time.Now()

					if err := trustCenter.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot update trust center: %w", err)
					}
				}
			case errors.Is(err, coredata.ErrResourceNotFound):
			default:
				return fmt.Errorf("cannot load trust center by domain id: %w", err)
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

// GetCertificate returns the certificate backing a custom domain, or nil when
// the domain has no certificate yet.
func (s *Service) GetCertificate(
	ctx context.Context,
	scope coredata.Scoper,
	domain *coredata.CustomDomain,
) (*coredata.Certificate, error) {
	if domain == nil || domain.CertificateID == nil {
		return nil, nil
	}

	return s.certManager.Get(ctx, scope, *domain.CertificateID)
}

// GetDefaultDomain returns the compliance page's default probopage subdomain,
// or nil when it has not been provisioned yet.
func (s *Service) GetDefaultDomain(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (*coredata.CustomDomain, error) {
	return s.domainSlot(ctx, scope, compliancePageID, func(tc *coredata.TrustCenter) *gid.GID {
		return tc.DefaultDomainID
	},
	)
}

// GetCustomDomain returns the compliance page's custom domain, or nil when
// none is configured.
func (s *Service) GetCustomDomain(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (*coredata.CustomDomain, error) {
	return s.domainSlot(ctx, scope, compliancePageID, func(tc *coredata.TrustCenter) *gid.GID {
		return tc.CustomDomainID
	},
	)
}

func (s *Service) domainSlot(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	slot func(*coredata.TrustCenter) *gid.GID,
) (*coredata.CustomDomain, error) {
	var domain *coredata.CustomDomain

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			domainID := slot(trustCenter)
			if domainID == nil {
				return nil
			}

			loaded := &coredata.CustomDomain{}
			if err := loaded.LoadByID(ctx, conn, scope, *domainID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load custom domain: %w", err)
			}

			domain = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

// EffectiveDomain returns the domain a compliance page is served under: the
// custom domain when it has an active certificate, otherwise the default
// subdomain when its certificate is active. It returns nil when no serving
// domain is available yet.
func (s *Service) EffectiveDomain(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (*coredata.CustomDomain, error) {
	var effective *coredata.CustomDomain

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			d, err := complianceportal.EffectiveDomainForTrustCenter(ctx, conn, scope, trustCenter)
			if err != nil {
				return err
			}

			effective = d

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return effective, nil
}

// EffectiveCanonicalHost returns the host a compliance page should be served
// under, or an empty string when no serving host is available yet.
func (s *Service) EffectiveCanonicalHost(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (string, error) {
	domain, err := s.EffectiveDomain(ctx, scope, compliancePageID)
	if err != nil {
		return "", err
	}

	if domain == nil {
		return "", nil
	}

	return domain.Domain, nil
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
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			url, err := complianceportal.PublicURLForTrustCenter(ctx, conn, scope, trustCenter, s.baseDomain)
			if err != nil {
				return err
			}

			publicURL = url

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return publicURL, nil
}
