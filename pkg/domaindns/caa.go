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

package domaindns

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/miekg/dns"
)

// CheckCAA verifies that CAA policy from hostname up through each parent to
// the registrable apex permits issuance by permittedIssuer (RFC 6844).
func (c *Client) CheckCAA(ctx context.Context, hostname, permittedIssuer string) error {
	checkNames, err := HostnamesForCAA(hostname)
	if err != nil {
		return err
	}

	for _, checkName := range checkNames {
		fqdn := ToFQDN(checkName)

		msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
		msg.Question = []dns.RR{&dns.CAA{Hdr: dns.Header{Name: fqdn, Class: dns.ClassINET}}}

		resp, err := c.query(ctx, msg)
		if err != nil {
			return fmt.Errorf("cannot exchange dns message for caa records: %w", err)
		}

		caaRecords := caaRecordsOwnedBy(resp, fqdn)
		if len(caaRecords) == 0 {
			continue
		}

		if caaPermitsIssuer(caaRecords, permittedIssuer) {
			return nil
		}

		return fmt.Errorf("%w: domain %q by %q", ErrCAADenied, hostname, permittedIssuer)
	}

	return nil
}

func caaRecordsOwnedBy(resp *dns.Msg, owner string) []*dns.CAA {
	var records []*dns.CAA

	for _, rr := range resp.Answer {
		caa, ok := rr.(*dns.CAA)
		if !ok || !EqualNames(caa.Hdr.Name, owner) {
			continue
		}

		records = append(records, caa)
	}

	return records
}

func caaPermitsIssuer(records []*dns.CAA, permittedIssuer string) bool {
	for _, caa := range records {
		if caa.Tag != "issue" {
			continue
		}

		issuer, _, _ := strings.Cut(caa.Value, ";")
		if strings.EqualFold(strings.TrimSpace(issuer), permittedIssuer) {
			return true
		}
	}

	return false
}
