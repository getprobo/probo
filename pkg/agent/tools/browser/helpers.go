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
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

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

func withToolTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, defaultToolTimeout)
}
