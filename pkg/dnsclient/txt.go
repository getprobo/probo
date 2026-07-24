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

package dnsclient

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/miekg/dns"
)

// CheckTXT verifies that domain has a TXT record owned by that exact name whose
// value equals expectedValue. Parent apex records are ignored.
func (c *Client) CheckTXT(ctx context.Context, domain, expectedValue string) error {
	msg := dns.NewMsg(domain, dns.TypeTXT)

	resp, err := c.query(ctx, msg)
	if err != nil {
		return fmt.Errorf("cannot query TXT record for %q: %w", domain, err)
	}

	if resp.Rcode == dns.RcodeNameError {
		return fmt.Errorf("%w for %q", ErrTXTNotFound, domain)
	}

	if resp.Rcode != dns.RcodeSuccess {
		return fmt.Errorf(
			"cannot query TXT record for %q: %s",
			domain,
			dns.RcodeToString[resp.Rcode],
		)
	}

	if len(resp.Answer) == 0 {
		return fmt.Errorf("%w for %q", ErrTXTNotFound, domain)
	}

	for _, answer := range resp.Answer {
		txt, ok := answer.(*dns.TXT)
		if !ok || !EqualNames(txt.Hdr.Name, domain) {
			continue
		}

		if strings.Join(txt.Txt, "") == expectedValue {
			return nil
		}
	}

	return fmt.Errorf("%w for %q", ErrTXTMismatch, domain)
}
