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
	"encoding/json"
	"fmt"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

type (
	findLinksParams struct {
		URL     string `json:"url" jsonschema:"The URL to search for links"`
		Pattern string `json:"pattern" jsonschema:"Keyword to filter links by (case-insensitive match on href or text)"`
	}
)

func FindLinksMatchingTool(b *Browser) agent.Tool {
	return agent.FunctionTool(
		"find_links_matching",
		"Navigate to a URL and extract links whose href or text matches a keyword (case-insensitive).",
		func(ctx context.Context, p findLinksParams) (agent.ToolResult, error) {
			if r := b.checkAlive(); r != nil {
				return *r, nil
			}

			if r := b.checkURL(p.URL); r != nil {
				return *r, nil
			}

			if p.Pattern == "" {
				return agent.ResultError("pattern must not be empty"), nil
			}

			ctx, timeoutCancel := withToolTimeout(ctx)
			defer timeoutCancel()

			tabCtx, cancel := b.NewTab(ctx)
			defer cancel()

			var links []link

			patternJSON, err := json.Marshal(p.Pattern)
			if err != nil {
				return agent.ResultErrorf("cannot encode pattern: %s", err), nil
			}

			js := fmt.Sprintf(
				`(() => {
					const pattern = (%s).toLowerCase();
					const normalize = s => s.replace(/[-_\s]+/g, "");
					const normalizedPattern = normalize(pattern);
					return Array.from(document.querySelectorAll("a[href]"))
						.filter(a => {
							const href = a.href.toLowerCase();
							const text = a.innerText.toLowerCase();
							return href.includes(pattern) || text.includes(pattern)
								|| normalize(href).includes(normalizedPattern)
								|| normalize(text).includes(normalizedPattern);
						})
						.map(a => ({
							href: a.href,
							text: a.innerText.trim().substring(0, 200)
						}));
				})()`,
				string(patternJSON),
			)

			err = chromedp.Run(
				tabCtx,
				chromedp.Navigate(p.URL),
				waitForPage(),
				chromedp.Evaluate(js, &links),
			)
			if err != nil {
				return agent.ResultError(b.classifyError(ctx, p.URL, err)), nil
			}

			return agent.ResultJSON(links), nil
		},
	)
}
