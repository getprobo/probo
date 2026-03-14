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

type findLinksParams struct {
	URL     string `json:"url" jsonschema:"description=The URL to search for links"`
	Pattern string `json:"pattern" jsonschema:"description=Keyword to filter links by (case-insensitive match on href or text)"`
}

func FindLinksMatchingTool(b *Browser) (agent.Tool, error) {
	return agent.FunctionTool[findLinksParams](
		"find_links_matching",
		"Navigate to a URL and extract links whose href or text matches a keyword (case-insensitive).",
		func(ctx context.Context, p findLinksParams) (agent.ToolResult, error) {
			tabCtx, cancel := b.NewTab(ctx)
			defer cancel()

			var links []link

			js := fmt.Sprintf(
				`(() => {
					const pattern = %q.toLowerCase();
					return Array.from(document.querySelectorAll("a[href]"))
						.filter(a => {
							const href = a.href.toLowerCase();
							const text = a.innerText.toLowerCase();
							return href.includes(pattern) || text.includes(pattern);
						})
						.map(a => ({
							href: a.href,
							text: a.innerText.trim().substring(0, 200)
						}));
				})()`,
				p.Pattern,
			)

			err := chromedp.Run(
				tabCtx,
				chromedp.Navigate(p.URL),
				chromedp.WaitReady("body"),
				chromedp.Evaluate(js, &links),
			)
			if err != nil {
				return agent.ToolResult{
					Content: fmt.Sprintf("cannot find links on %s: %s", p.URL, err),
					IsError: true,
				}, nil
			}

			data, _ := json.Marshal(links)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
