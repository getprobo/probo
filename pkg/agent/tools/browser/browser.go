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

const defaultToolTimeout = 60 * time.Second

// waitForPage returns chromedp actions that wait for the page to fully load,
// including SPA content rendered by JavaScript. It first waits for the body to
// be ready, then polls until the page content stabilizes (innerText stops
// changing) with a short debounce. After stabilization, it attempts to dismiss
// common cookie consent banners so they don't interfere with content
// extraction.
func waitForPage() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if err := chromedp.WaitReady("body").Do(ctx); err != nil {
			return err
		}

		// Wait for SPA content to stabilize by checking if innerText
		// length stops changing over a 500ms window. Gives up after 5s.
		// EvaluateAsDevTools is required to await the Promise.
		if err := chromedp.EvaluateAsDevTools(`
			new Promise((resolve) => {
				let lastLen = -1;
				let stableCount = 0;
				const interval = setInterval(() => {
					const curLen = document.body.innerText.length;
					if (curLen === lastLen && curLen > 0) {
						stableCount++;
					} else {
						stableCount = 0;
					}
					lastLen = curLen;
					if (stableCount >= 2) {
						clearInterval(interval);
						resolve(true);
					}
				}, 250);
				setTimeout(() => {
					clearInterval(interval);
					resolve(true);
				}, 5000);
			})
		`, nil).Do(ctx); err != nil {
			return err
		}

		// Dismiss common cookie consent banners. This is best-effort;
		// failures are silently ignored because not every page has a
		// banner and the selectors may not match.
		return chromedp.Evaluate(`
			(() => {
				const selectors = [
					"#onetrust-accept-btn-handler",
					"#CybotCookiebotDialogBodyLevelButtonLevelOptinAllowAll",
					"#CybotCookiebotDialogBodyButtonAccept",
					".cky-btn-accept",
					"[data-testid='cookie-policy-dialog-accept-button']",
					"button.accept-cookies",
					"#cookie-accept",
					"#accept-cookies",
					".cc-accept",
					".cc-btn.cc-dismiss",
				];
				for (const sel of selectors) {
					const btn = document.querySelector(sel);
					if (btn) { btn.click(); return; }
				}
				const buttons = document.querySelectorAll(
					"button, a[role='button'], [role='button']"
				);
				const patterns = /^(accept all|accept|agree|i agree|allow all|allow|got it|ok|okay|consent)$/i;
				for (const btn of buttons) {
					if (patterns.test(btn.innerText.trim())) {
						btn.click();
						return;
					}
				}
			})()
		`, nil).Do(ctx)
	})
}

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

// checkPDF returns an error tool result if the URL points to a PDF file,
// which cannot be rendered by the headless browser.
func checkPDF(rawURL string) *agent.ToolResult {
	if strings.HasSuffix(strings.ToLower(rawURL), ".pdf") {
		return &agent.ToolResult{
			Content: fmt.Sprintf("cannot load %s: PDF files are not supported by the browser", rawURL),
			IsError: true,
		}
	}

	return nil
}

// checkURL validates that the URL is allowed. It returns an error tool result
// if the URL is outside the allowed domains.
func (b *Browser) checkURL(rawURL string) *agent.ToolResult {
	if len(b.allowedDomains) == 0 {
		return nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return &agent.ToolResult{
			Content: fmt.Sprintf("invalid URL: %s", err),
			IsError: true,
		}
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

// BuildReadOnlyTools returns browser tools that only read page content:
// navigate, extract text, extract links, and find links. It excludes
// interactive tools (click, select) that modify page state. Also includes
// standalone HTTP tools (robots.txt, sitemap, PDF download) that do not
// require the browser.
func BuildReadOnlyTools(b *Browser) ([]agent.Tool, error) {
	navigateTool, err := NavigateToURLTool(b)
	if err != nil {
		return nil, err
	}

	extractTextTool, err := ExtractPageTextTool(b)
	if err != nil {
		return nil, err
	}

	extractLinksTool, err := ExtractLinksTool(b)
	if err != nil {
		return nil, err
	}

	findLinksTool, err := FindLinksMatchingTool(b)
	if err != nil {
		return nil, err
	}

	robotsTool, err := FetchRobotsTxtTool()
	if err != nil {
		return nil, err
	}

	sitemapTool, err := FetchSitemapTool()
	if err != nil {
		return nil, err
	}

	pdfTool, err := DownloadPDFTool()
	if err != nil {
		return nil, err
	}

	return []agent.Tool{
		navigateTool,
		extractTextTool,
		extractLinksTool,
		findLinksTool,
		robotsTool,
		sitemapTool,
		pdfTool,
	}, nil
}

func BuildTools(b *Browser) ([]agent.Tool, error) {
	navigateTool, err := NavigateToURLTool(b)
	if err != nil {
		return nil, err
	}

	extractTextTool, err := ExtractPageTextTool(b)
	if err != nil {
		return nil, err
	}

	extractLinksTool, err := ExtractLinksTool(b)
	if err != nil {
		return nil, err
	}

	findLinksTool, err := FindLinksMatchingTool(b)
	if err != nil {
		return nil, err
	}

	clickTool, err := ClickElementTool(b)
	if err != nil {
		return nil, err
	}

	selectTool, err := SelectOptionTool(b)
	if err != nil {
		return nil, err
	}

	robotsTool, err := FetchRobotsTxtTool()
	if err != nil {
		return nil, err
	}

	sitemapTool, err := FetchSitemapTool()
	if err != nil {
		return nil, err
	}

	pdfTool, err := DownloadPDFTool()
	if err != nil {
		return nil, err
	}

	return []agent.Tool{
		navigateTool,
		extractTextTool,
		extractLinksTool,
		findLinksTool,
		clickTool,
		selectTool,
		robotsTool,
		sitemapTool,
		pdfTool,
	}, nil
}
