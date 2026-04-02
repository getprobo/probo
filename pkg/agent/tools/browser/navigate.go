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
	"encoding/json"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

func withToolTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, defaultToolTimeout)
}

type navigateParams struct {
	URL string `json:"url" jsonschema:"The URL to navigate to"`
}

type navigateResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	FinalURL    string `json:"final_url"`
}

func NavigateToURLTool(b *Browser) (agent.Tool, error) {
	return agent.FunctionTool[navigateParams](
		"navigate_to_url",
		"Navigate to a URL and return the page title, meta description, and final URL after redirects.",
		func(ctx context.Context, p navigateParams) (agent.ToolResult, error) {
			if r := b.checkAlive(); r != nil {
				return *r, nil
			}

			if r := b.checkURL(p.URL); r != nil {
				return *r, nil
			}

			if r := checkPDF(p.URL); r != nil {
				return *r, nil
			}

			ctx, timeoutCancel := withToolTimeout(ctx)
			defer timeoutCancel()

			tabCtx, cancel := b.NewTab(ctx)
			defer cancel()

			var (
				title       string
				description string
				finalURL    string
			)

			err := chromedp.Run(
				tabCtx,
				chromedp.Navigate(p.URL),
				waitForPage(),
				chromedp.Title(&title),
				chromedp.Evaluate(
					`(() => {
						const meta = document.querySelector('meta[name="description"]');
						return meta ? meta.getAttribute("content") : "";
					})()`,
					&description,
				),
				chromedp.Location(&finalURL),
			)
			if err != nil {
				return agent.ToolResult{
					Content: b.classifyError(ctx, p.URL, err),
					IsError: true,
				}, nil
			}

			data, _ := json.Marshal(navigateResult{
				Title:       title,
				Description: description,
				FinalURL:    finalURL,
			})

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
