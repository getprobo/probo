// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package browser

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

const (
	defaultToolTimeout = 60 * time.Second
)

type Browser struct {
	addr           string
	allocCtx       context.Context
	cancel         context.CancelFunc
	allowedDomains []string
}

func NewBrowser(ctx context.Context, addr string) *Browser {
	if !strings.HasPrefix(addr, "ws://") && !strings.HasPrefix(addr, "wss://") {
		addr = "ws://" + addr
	}

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, addr)

	return &Browser{
		addr:     addr,
		allocCtx: allocCtx,
		cancel:   cancel,
	}
}

// SetAllowedDomain restricts navigation to URLs under the given domain and
// its subdomains. For example, setting "getprobo.com" allows navigation to
// getprobo.com, www.getprobo.com, and compliance.getprobo.com.
// This replaces any previously set domains.
func (b *Browser) SetAllowedDomain(domain string) {
	domain = strings.ToLower(strings.TrimSpace(domain))

	// Strip "www." prefix so that setting either "www.example.com" or
	// "example.com" allows navigation to *.example.com.
	domain = strings.TrimPrefix(domain, "www.")

	b.allowedDomains = []string{domain}
}

// checkURL validates that the URL is allowed. It returns an error tool result
// if the URL uses a disallowed scheme or is outside the allowed domains.
func (b *Browser) checkURL(rawURL string) *agent.ToolResult {
	u, err := url.Parse(rawURL)
	if err != nil {
		return &agent.ToolResult{
			Content: fmt.Sprintf("invalid URL: %s", err),
			IsError: true,
		}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return &agent.ToolResult{
			Content: fmt.Sprintf("cannot navigate to URL with scheme %q: only http and https are allowed", u.Scheme),
			IsError: true,
		}
	}

	if len(b.allowedDomains) == 0 {
		return nil
	}

	host := strings.ToLower(u.Hostname())
	for _, allowed := range b.allowedDomains {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return nil
		}
	}

	return &agent.ToolResult{
		Content: fmt.Sprintf("navigation blocked: %s is outside the allowed domains", host),
		IsError: true,
	}
}

// checkAlive returns a tool error result if the browser connection has been
// lost. Call this at the start of every tool to fail fast with a clear
// message instead of waiting for the tool timeout.
func (b *Browser) checkAlive() *agent.ToolResult {
	if err := b.allocCtx.Err(); err != nil {
		return &agent.ToolResult{
			Content: "browser connection lost: the remote Chrome instance is no longer reachable",
			IsError: true,
		}
	}
	return nil
}

// classifyError inspects the caller's timeout context and the browser's
// allocator context to produce a human-readable error message. Without this,
// both a tool timeout and a dropped Chrome connection appear as the opaque
// "context canceled".
func (b *Browser) classifyError(timeoutCtx context.Context, rawURL string, err error) string {
	if b.allocCtx.Err() != nil {
		return fmt.Sprintf(
			"browser connection lost while loading %s: the remote Chrome instance is no longer reachable",
			rawURL,
		)
	}

	if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
		return fmt.Sprintf(
			"page load timed out after %s for %s: the page may be too slow or unresponsive",
			defaultToolTimeout,
			rawURL,
		)
	}

	return fmt.Sprintf("cannot load %s: %s", rawURL, err)
}

func (b *Browser) NewTab(ctx context.Context) (context.Context, context.CancelFunc) {
	tabCtx, tabCancel := chromedp.NewContext(b.allocCtx)

	// Propagate the caller's cancellation to the Chrome tab so that
	// tool-level timeouts and context deadlines actually stop the browser.
	go func() {
		select {
		case <-ctx.Done():
			tabCancel()
		case <-tabCtx.Done():
		}
	}()

	return tabCtx, tabCancel
}

func (b *Browser) Close() {
	b.cancel()
}
