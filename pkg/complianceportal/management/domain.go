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
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *Service) EffectiveDomainForCompliancePortal(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePage *coredata.CompliancePortal,
) (*coredata.CustomDomain, error) {
	byID, active, err := s.loadDomains(ctx, conn, scope, compliancePage)
	if err != nil {
		return nil, err
	}

	if compliancePage.CustomDomainID != nil {
		if d := byID[*compliancePage.CustomDomainID]; d != nil && active[d.ID] {
			return d, nil
		}
	}

	if compliancePage.DefaultDomainID != nil {
		if d := byID[*compliancePage.DefaultDomainID]; d != nil && active[d.ID] {
			return d, nil
		}
	}

	return nil, nil
}

func (s *Service) PublicURLForCompliancePortal(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePage *coredata.CompliancePortal,
) (string, error) {
	byID, active, err := s.loadDomains(ctx, conn, scope, compliancePage)
	if err != nil {
		return "", err
	}

	host := publicHostForCompliancePortal(compliancePage, byID, active, s.baseDomain)

	return "https://" + host, nil
}

func (s *Service) loadDomains(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePage *coredata.CompliancePortal,
) (map[gid.GID]*coredata.CustomDomain, map[gid.GID]bool, error) {
	var ids []gid.GID
	if compliancePage.CustomDomainID != nil {
		ids = append(ids, *compliancePage.CustomDomainID)
	}

	if compliancePage.DefaultDomainID != nil {
		ids = append(ids, *compliancePage.DefaultDomainID)
	}

	byID := make(map[gid.GID]*coredata.CustomDomain)

	if len(ids) == 0 {
		return byID, make(map[gid.GID]bool), nil
	}

	var domains coredata.CustomDomains
	if err := domains.LoadByIDs(ctx, conn, scope, ids); err != nil {
		return nil, nil, fmt.Errorf("cannot load custom domains: %w", err)
	}

	for _, d := range domains {
		byID[d.ID] = d
	}

	// In external TLS mode, the customer's own reverse proxy terminates TLS
	// for these domains, so Probo never provisions a certificate for them and
	// certificate status can't gate whether a domain is usable.
	if s.externallyTerminatedTLS {
		return byID, activeDomains(domains, nil, true), nil
	}

	var certificateIDs []gid.GID

	for _, d := range domains {
		if d.CertificateID != nil {
			certificateIDs = append(certificateIDs, *d.CertificateID)
		}
	}

	if len(certificateIDs) == 0 {
		return byID, make(map[gid.GID]bool), nil
	}

	var certificates coredata.Certificates
	if err := certificates.LoadByIDs(ctx, conn, scope, certificateIDs); err != nil {
		return nil, nil, fmt.Errorf("cannot load certificates: %w", err)
	}

	return byID, activeDomains(domains, certificates, false), nil
}

// activeDomains reports, for each loaded custom domain, whether it is usable
// as a public hostname. In external TLS mode the customer's own reverse
// proxy terminates TLS, so Probo never provisions or tracks a certificate
// for these domains and every linked domain is usable regardless of
// certificate status. Otherwise a domain is usable only once its own
// certificate has reached ACTIVE.
func activeDomains(
	domains coredata.CustomDomains,
	certificates coredata.Certificates,
	externallyTerminatedTLS bool,
) map[gid.GID]bool {
	active := make(map[gid.GID]bool)

	if externallyTerminatedTLS {
		for _, d := range domains {
			active[d.ID] = true
		}

		return active
	}

	domainByCertificate := make(map[gid.GID]gid.GID)
	for _, d := range domains {
		if d.CertificateID != nil {
			domainByCertificate[*d.CertificateID] = d.ID
		}
	}

	for _, c := range certificates {
		if domainID, ok := domainByCertificate[c.ID]; ok {
			active[domainID] = c.Status == coredata.CertificateStatusActive
		}
	}

	return active
}

// publicHostForCompliancePortal picks the hostname a compliance portal
// should be publicly reached at: its custom domain if usable, else its
// default domain if usable, else the Probo-hosted slug fallback.
func publicHostForCompliancePortal(
	compliancePage *coredata.CompliancePortal,
	byID map[gid.GID]*coredata.CustomDomain,
	active map[gid.GID]bool,
	baseDomain string,
) string {
	switch {
	case compliancePage.CustomDomainID != nil && byID[*compliancePage.CustomDomainID] != nil && active[*compliancePage.CustomDomainID]:
		return byID[*compliancePage.CustomDomainID].Domain
	case compliancePage.DefaultDomainID != nil && byID[*compliancePage.DefaultDomainID] != nil && active[*compliancePage.DefaultDomainID]:
		return byID[*compliancePage.DefaultDomainID].Domain
	default:
		return compliancePage.Slug + "." + baseDomain
	}
}
