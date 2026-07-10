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

package resolver

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// EffectiveDomainForTrustCenter returns the domain a compliance page is served
// under: the custom domain when it has an active certificate, otherwise the
// default subdomain when its certificate is active. It returns nil when no
// serving domain is available yet.
func EffectiveDomainForTrustCenter(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	trustCenter *coredata.TrustCenter,
) (*coredata.CustomDomain, error) {
	byID, active, err := loadDomains(ctx, conn, scope, trustCenter)
	if err != nil {
		return nil, err
	}

	if trustCenter.CustomDomainID != nil {
		if d := byID[*trustCenter.CustomDomainID]; d != nil && active[d.ID] {
			return d, nil
		}
	}

	if trustCenter.DefaultDomainID != nil {
		if d := byID[*trustCenter.DefaultDomainID]; d != nil && active[d.ID] {
			return d, nil
		}
	}

	return nil, nil
}

// PublicURLForTrustCenter returns the canonical public URL of a compliance
// page. Compliance pages are always served on a dedicated domain: the custom
// domain when its certificate is active, otherwise the default probopage
// subdomain (even while its certificate provisions), and finally the default
// subdomain hostname derived from the page slug when no domain row is loaded
// yet.
func PublicURLForTrustCenter(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	trustCenter *coredata.TrustCenter,
	baseDomain string,
) (string, error) {
	byID, active, err := loadDomains(ctx, conn, scope, trustCenter)
	if err != nil {
		return "", err
	}

	var host string

	switch {
	case trustCenter.CustomDomainID != nil && byID[*trustCenter.CustomDomainID] != nil && active[*trustCenter.CustomDomainID]:
		host = byID[*trustCenter.CustomDomainID].Domain
	case trustCenter.DefaultDomainID != nil && byID[*trustCenter.DefaultDomainID] != nil:
		host = byID[*trustCenter.DefaultDomainID].Domain
	case trustCenter.CustomDomainID != nil && byID[*trustCenter.CustomDomainID] != nil:
		host = byID[*trustCenter.CustomDomainID].Domain
	}

	if host == "" {
		host = trustCenter.Slug + "." + baseDomain
	}

	return "https://" + host, nil
}

func loadDomains(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	trustCenter *coredata.TrustCenter,
) (map[gid.GID]*coredata.CustomDomain, map[gid.GID]bool, error) {
	var ids []gid.GID
	if trustCenter.CustomDomainID != nil {
		ids = append(ids, *trustCenter.CustomDomainID)
	}

	if trustCenter.DefaultDomainID != nil {
		ids = append(ids, *trustCenter.DefaultDomainID)
	}

	byID := make(map[gid.GID]*coredata.CustomDomain)
	active := make(map[gid.GID]bool)

	if len(ids) == 0 {
		return byID, active, nil
	}

	var domains coredata.CustomDomains
	if err := domains.LoadByIDs(ctx, conn, scope, ids); err != nil {
		return nil, nil, fmt.Errorf("cannot load custom domains: %w", err)
	}

	var certificateIDs []gid.GID
	domainByCertificate := make(map[gid.GID]gid.GID)

	for _, d := range domains {
		byID[d.ID] = d
		if d.CertificateID != nil {
			certificateIDs = append(certificateIDs, *d.CertificateID)
			domainByCertificate[*d.CertificateID] = d.ID
		}
	}

	if len(certificateIDs) == 0 {
		return byID, active, nil
	}

	var certificates coredata.Certificates
	if err := certificates.LoadByIDs(ctx, conn, scope, certificateIDs); err != nil {
		return nil, nil, fmt.Errorf("cannot load certificates: %w", err)
	}

	for _, c := range certificates {
		if domainID, ok := domainByCertificate[c.ID]; ok {
			active[domainID] = c.Status == coredata.CertificateStatusActive
		}
	}

	return byID, active, nil
}
