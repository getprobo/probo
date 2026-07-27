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

const (
	maxCNAMELookups = 4
)

func (c *Client) CheckCNAME(ctx context.Context, hostname, expectedTarget string) error {
	owner := ToFQDN(hostname)
	target := ToFQDN(expectedTarget)

	current := owner
	visited := map[string]struct{}{owner: {}}

	for range maxCNAMELookups {
		queryCtx, cancel := c.withExchangeTimeout(ctx)
		resp, err := c.queryCNAME(queryCtx, current)

		cancel()

		if err != nil {
			return fmt.Errorf("cannot query cname for domain %q: %w", trimRootDot(current), err)
		}

		edges, err := cnameEdges(resp)
		if err != nil {
			return err
		}

		next, ok := edges[current]
		if !ok {
			return cnameChainStopError(resp, hostname, owner, current, target)
		}

		// Resolvers disagree on QTYPE=CNAME: some answer with the first hop
		// only, others chase and return every hop, so consume the hops this
		// answer already carries before spending another lookup.
		for {
			if _, seen := visited[next]; seen {
				return fmt.Errorf(
					"cname chain for domain %q loops back to %q",
					hostname,
					trimRootDot(next),
				)
			}

			visited[next] = struct{}{}

			if EqualNames(next, target) {
				return nil
			}

			following, ok := edges[next]
			if !ok {
				break
			}

			next = following
		}

		current = next
	}

	return fmt.Errorf(
		"cname chain for domain %q does not reach %q within %d lookups",
		hostname,
		trimRootDot(target),
		maxCNAMELookups,
	)
}

func (c *Client) queryCNAME(ctx context.Context, name string) (*dns.Msg, error) {
	msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
	msg.Question = []dns.RR{&dns.CNAME{Hdr: dns.Header{Name: name, Class: dns.ClassINET}}}

	return c.query(ctx, msg)
}

func cnameEdges(resp *dns.Msg) (map[string]string, error) {
	edges := make(map[string]string)

	for _, rr := range resp.Answer {
		cname, ok := rr.(*dns.CNAME)
		if !ok {
			continue
		}

		name := ToFQDN(cname.Hdr.Name)
		if _, duplicate := edges[name]; duplicate {
			return nil, fmt.Errorf("multiple cname records found for domain %q", trimRootDot(name))
		}

		edges[name] = ToFQDN(cname.Target)
	}

	return edges, nil
}

func cnameChainStopError(
	resp *dns.Msg,
	hostname string,
	owner string,
	current string,
	target string,
) error {
	if !EqualNames(current, owner) {
		return fmt.Errorf(
			"cname chain for domain %q stops at %q, expected %q",
			hostname,
			trimRootDot(current),
			trimRootDot(target),
		)
	}

	for _, rr := range resp.Answer {
		cname, ok := rr.(*dns.CNAME)
		if !ok {
			continue
		}

		return fmt.Errorf(
			"cname owner mismatch: domain %q has record owned by %q",
			hostname,
			trimRootDot(cname.Hdr.Name),
		)
	}

	return fmt.Errorf("no cname records found for domain %q", hostname)
}

func trimRootDot(name string) string {
	return strings.TrimSuffix(name, ".")
}
