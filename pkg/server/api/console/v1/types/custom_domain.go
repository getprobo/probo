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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
)

// NewCustomDomain builds the GraphQL CustomDomain type. The TLS lifecycle now
// lives on the linked certificate; when cert is nil (certificate not yet
// created) the domain reports a pending SSL status.
func NewCustomDomain(d *coredata.CustomDomain, cert *coredata.Certificate, cnameTarget string) *CustomDomain {
	result := &CustomDomain{
		ID: d.ID,
		Organization: &Organization{
			ID: d.OrganizationID,
		},
		Domain:    d.Domain,
		Managed:   d.Managed,
		SslStatus: coredata.CustomDomainSSLStatusPending,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}

	if cert != nil {
		result.SslStatus = coredata.CustomDomainSSLStatus(cert.Status)
		result.SslExpiresAt = cert.SSLExpiresAt
		result.ProvisioningError = cert.ProvisioningError
	}

	// Convert DNS records
	result.DNSRecords = convertDNSRecords(d, cnameTarget)

	return result
}

func convertDNSRecords(d *coredata.CustomDomain, cnameTarget string) []*DNSRecordInstruction {
	var records []*DNSRecordInstruction

	// For HTTP-01 challenges, we just need the domain to point to our servers via CNAME
	record := &DNSRecordInstruction{
		Type:    "CNAME",
		Name:    d.Domain,
		Value:   cnameTarget,
		TTL:     300,
		Purpose: "Point domain to Probo servers",
	}
	records = append(records, record)

	return records
}
