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

package scanner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"go.probo.inc/probo/pkg/netcheck"
)

const hostLookupTimeout = 2 * time.Second

type hostVerdict int

const (
	hostPublic hostVerdict = iota
	hostPrivate
	hostUnknown
)

// requestPolicy decides which paused requests the browser may send.
//
// Requests from other targets, and WebSockets, are not paused here. The
// Chrome host's egress rules are what stops those from reaching a private
// address. Cross-site iframes are separate targets, so this policy does not
// see them.
type requestPolicy struct {
	site      origin
	api       origin
	mainFrame string
	skip      bool

	mu    sync.Mutex
	hosts map[string]hostVerdict
}

func newRequestPolicy(site, api origin, skip bool) *requestPolicy {
	return &requestPolicy{
		site:  site,
		api:   api,
		skip:  skip,
		hosts: map[string]hostVerdict{},
	}
}

func (p *requestPolicy) setMainFrame(id string) {
	p.mainFrame = id
}

// block reports whether the browser must fail this request.
func (p *requestPolicy) block(ctx context.Context, method, rawURL, kind, frameID string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || rawURL == "" {
		return true
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	if parsed.User != nil || parsed.Hostname() == "" {
		return true
	}

	offSiteDocument := kind == string(network.ResourceTypeDocument) &&
		frameID != "" &&
		frameID == p.mainFrame &&
		!p.site.contains(rawURL)
	if offSiteDocument {
		return true
	}

	if p.api.contains(rawURL) && !readMethod(method) {
		return true
	}

	return p.lookup(ctx, parsed.Hostname()) != hostPublic
}

func readMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

// resolveStart chooses the first page URL and checks that it is public.
func (p *requestPolicy) resolveStart(ctx context.Context, site origin) (string, error) {
	if p.skip {
		return site.root(), nil
	}

	apex := p.lookup(ctx, site.host)

	www := hostUnknown
	if apex == hostUnknown {
		www = p.lookup(ctx, "www."+site.host)
	}

	return chooseStart(site, apex, www)
}

func (p *requestPolicy) lookup(ctx context.Context, host string) hostVerdict {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" {
		return hostPrivate
	}

	if p.skip {
		return hostPublic
	}

	if ip := net.ParseIP(host); ip != nil {
		if netcheck.IsPublicIP(ip) {
			return hostPublic
		}

		return hostPrivate
	}

	p.mu.Lock()
	if verdict, ok := p.hosts[host]; ok {
		p.mu.Unlock()

		return verdict
	}

	p.mu.Unlock()

	verdict := lookupHost(ctx, host)
	if ctx.Err() != nil {
		return verdict
	}

	p.mu.Lock()
	p.hosts[host] = verdict
	p.mu.Unlock()

	return verdict
}

func lookupHost(ctx context.Context, host string) hostVerdict {
	lookupCtx, cancel := context.WithTimeout(ctx, hostLookupTimeout)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIPAddr(lookupCtx, host)
	if err != nil || len(ips) == 0 {
		return hostUnknown
	}

	for _, ip := range ips {
		if !netcheck.IsPublicIP(ip.IP) {
			return hostPrivate
		}
	}

	return hostPublic
}

func errPrivate(host string) error {
	return fmt.Errorf("cannot start scan: %w: host %q", netcheck.ErrBlockedAddress, host)
}

func errResolve(host string) error {
	return fmt.Errorf("cannot start scan: cannot resolve host %q", host)
}
