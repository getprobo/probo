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

// CheckCNAME verifies that hostname has a single CNAME record owned by that
// name and pointing at expectedTarget.
func (c *Client) CheckCNAME(ctx context.Context, hostname, expectedTarget string) error {
	owner := ToFQDN(hostname)
	target := ToFQDN(expectedTarget)

	msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
	msg.Question = []dns.RR{&dns.CNAME{Hdr: dns.Header{Name: owner, Class: dns.ClassINET}}}

	resp, err := c.query(ctx, msg)
	if err != nil {
		return err
	}

	if len(resp.Answer) == 0 {
		return fmt.Errorf("no cname records found for domain %q", hostname)
	}

	if len(resp.Answer) > 1 {
		return fmt.Errorf("multiple cname records found for domain %q", hostname)
	}

	resolvedRecord, ok := resp.Answer[0].(*dns.CNAME)
	if !ok {
		return fmt.Errorf("first answer is not a cname record for domain %q", hostname)
	}

	if !EqualNames(resolvedRecord.Hdr.Name, owner) {
		return fmt.Errorf(
			"cname owner mismatch: domain %q has record owned by %q",
			hostname,
			strings.TrimSuffix(resolvedRecord.Hdr.Name, "."),
		)
	}

	if !EqualNames(resolvedRecord.Target, target) {
		return fmt.Errorf(
			"cname target mismatch: domain %q resolves to %q, expected %q",
			hostname,
			strings.TrimSuffix(resolvedRecord.Target, "."),
			strings.TrimSuffix(expectedTarget, "."),
		)
	}

	return nil
}
