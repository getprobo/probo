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

package dnsclient

import (
	"context"
	"fmt"

	"codeberg.org/miekg/dns"
)

type (
	// Client performs DNS lookups used to verify domain ownership and
	// certificate prerequisites.
	Client struct {
		ResolverAddr string
		exchange     exchangeFunc
	}

	exchangeFunc func(ctx context.Context, msg *dns.Msg, network string) (*dns.Msg, error)
)

// NewClient returns a client that resolves names through resolverAddr.
func NewClient(resolverAddr string) *Client {
	return &Client{ResolverAddr: resolverAddr}
}

func (c *Client) exchangeUDP(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	if c.exchange != nil {
		return c.exchange(ctx, msg, "udp")
	}

	client := dns.NewClient()

	resp, _, err := client.Exchange(ctx, msg, "udp", c.ResolverAddr)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) exchangeTCP(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	if c.exchange != nil {
		return c.exchange(ctx, msg, "tcp")
	}

	client := dns.NewClient()

	resp, _, err := client.Exchange(ctx, msg, "tcp", c.ResolverAddr)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) query(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	resp, err := c.exchangeUDP(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("cannot exchange dns message: %w", err)
	}

	return resp, nil
}
