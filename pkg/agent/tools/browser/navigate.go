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

package browser

import (
	"context"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

type (
	navigateParams struct {
		URL string `json:"url" jsonschema:"The URL to navigate to"`
	}

	navigateResult struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		FinalURL    string `json:"final_url"`
	}
)

func NavigateToURLTool(b *Browser) agent.Tool {
	return agent.FunctionTool(
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
				return agent.ResultError(b.classifyError(ctx, p.URL, err)), nil
			}

			return agent.ResultJSON(
				navigateResult{
					Title:       title,
					Description: description,
					FinalURL:    finalURL,
				},
			), nil
		},
	)
}
