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

package management

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func newTestCustomDomain(tenantID gid.TenantID, domain string, certificateID *gid.GID) *coredata.CustomDomain {
	return &coredata.CustomDomain{
		ID:            gid.New(tenantID, coredata.CustomDomainEntityType),
		Domain:        domain,
		CertificateID: certificateID,
	}
}

func newTestCertificate(tenantID gid.TenantID, status coredata.CertificateStatus) *coredata.Certificate {
	return &coredata.Certificate{
		ID:     gid.New(tenantID, coredata.CertificateEntityType),
		Status: status,
	}
}

func TestActiveDomains(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()

	t.Run(
		"external TLS mode treats every linked domain as active regardless of certificate status",
		func(t *testing.T) {
			t.Parallel()

			certificate := newTestCertificate(tenantID, coredata.CertificateStatusPending)
			domain := newTestCustomDomain(tenantID, "trust.acme.com", &certificate.ID)

			active := activeDomains(coredata.CustomDomains{domain}, nil, true)

			assert.True(t, active[domain.ID])
		},
	)

	t.Run(
		"external TLS mode treats a domain with no certificate at all as active",
		func(t *testing.T) {
			t.Parallel()

			domain := newTestCustomDomain(tenantID, "trust.acme.com", nil)

			active := activeDomains(coredata.CustomDomains{domain}, nil, true)

			assert.True(t, active[domain.ID])
		},
	)

	t.Run(
		"direct TLS mode requires the certificate to be active",
		func(t *testing.T) {
			t.Parallel()

			pendingCertificate := newTestCertificate(tenantID, coredata.CertificateStatusPending)
			pendingDomain := newTestCustomDomain(tenantID, "pending.acme.com", &pendingCertificate.ID)

			activeCertificate := newTestCertificate(tenantID, coredata.CertificateStatusActive)
			activeDomain := newTestCustomDomain(tenantID, "active.acme.com", &activeCertificate.ID)

			active := activeDomains(
				coredata.CustomDomains{pendingDomain, activeDomain},
				coredata.Certificates{pendingCertificate, activeCertificate},
				false,
			)

			assert.False(t, active[pendingDomain.ID])
			assert.True(t, active[activeDomain.ID])
		},
	)

	t.Run(
		"direct TLS mode treats a domain with no certificate as inactive",
		func(t *testing.T) {
			t.Parallel()

			domain := newTestCustomDomain(tenantID, "acme.com", nil)

			active := activeDomains(coredata.CustomDomains{domain}, nil, false)

			assert.False(t, active[domain.ID])
		},
	)
}

func TestPublicHostForCompliancePortal(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	baseDomain := "probo.example"

	t.Run(
		"uses the custom domain when active, even with a pending certificate in external TLS mode",
		func(t *testing.T) {
			t.Parallel()

			pendingCertificate := newTestCertificate(tenantID, coredata.CertificateStatusPending)
			customDomain := newTestCustomDomain(tenantID, "trust.acme.com", &pendingCertificate.ID)
			compliancePage := &coredata.CompliancePortal{
				Slug:           "acme",
				CustomDomainID: &customDomain.ID,
			}

			byID := map[gid.GID]*coredata.CustomDomain{customDomain.ID: customDomain}
			active := activeDomains(
				coredata.CustomDomains{customDomain},
				coredata.Certificates{pendingCertificate},
				true,
			)

			host := publicHostForCompliancePortal(compliancePage, byID, active, baseDomain)

			assert.Equal(t, "trust.acme.com", host)
		},
	)

	t.Run(
		"falls back to the slug host when the custom domain is not active",
		func(t *testing.T) {
			t.Parallel()

			pendingCertificate := newTestCertificate(tenantID, coredata.CertificateStatusPending)
			customDomain := newTestCustomDomain(tenantID, "trust.acme.com", &pendingCertificate.ID)
			compliancePage := &coredata.CompliancePortal{
				Slug:           "acme",
				CustomDomainID: &customDomain.ID,
			}

			byID := map[gid.GID]*coredata.CustomDomain{customDomain.ID: customDomain}
			active := activeDomains(
				coredata.CustomDomains{customDomain},
				coredata.Certificates{pendingCertificate},
				false,
			)

			host := publicHostForCompliancePortal(compliancePage, byID, active, baseDomain)

			assert.Equal(t, "acme."+baseDomain, host)
		},
	)

	t.Run(
		"falls back to the default domain when the custom domain is not active",
		func(t *testing.T) {
			t.Parallel()

			pendingCertificate := newTestCertificate(tenantID, coredata.CertificateStatusPending)
			customDomain := newTestCustomDomain(tenantID, "trust.acme.com", &pendingCertificate.ID)

			defaultDomain := newTestCustomDomain(tenantID, "acme."+baseDomain, nil)

			compliancePage := &coredata.CompliancePortal{
				Slug:            "acme",
				CustomDomainID:  &customDomain.ID,
				DefaultDomainID: &defaultDomain.ID,
			}

			byID := map[gid.GID]*coredata.CustomDomain{
				customDomain.ID:  customDomain,
				defaultDomain.ID: defaultDomain,
			}
			active := activeDomains(
				coredata.CustomDomains{customDomain, defaultDomain},
				coredata.Certificates{pendingCertificate},
				false,
			)
			// The default domain has no certificate of its own; simulate it
			// being provisioned and active, as Probo's own base domain is.
			active[defaultDomain.ID] = true

			host := publicHostForCompliancePortal(compliancePage, byID, active, baseDomain)

			assert.Equal(t, "acme."+baseDomain, host)
		},
	)

	t.Run(
		"falls back to the slug host when there is no custom or default domain",
		func(t *testing.T) {
			t.Parallel()

			compliancePage := &coredata.CompliancePortal{Slug: "acme"}

			host := publicHostForCompliancePortal(compliancePage, nil, nil, baseDomain)

			assert.Equal(t, "acme."+baseDomain, host)
		},
	)
}
