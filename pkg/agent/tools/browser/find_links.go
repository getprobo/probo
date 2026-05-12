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
					const pattern = JSON.parse(%s).toLowerCase();
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
