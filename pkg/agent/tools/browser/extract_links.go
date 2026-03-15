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
	"net/url"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

type extractLinksParams struct {
	URL string `json:"url" jsonschema:"description=The URL to extract links from"`
}

type link struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

func ExtractLinksTool(b *Browser) (agent.Tool, error) {
	return agent.FunctionTool[extractLinksParams](
		"extract_links",
		"Navigate to a URL and extract all links (<a> elements) with their href and text.",
		func(ctx context.Context, p extractLinksParams) (agent.ToolResult, error) {
			u, err := url.Parse(p.URL)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
				return agent.ToolResult{
					Content: "invalid URL scheme: only http and https are allowed",
					IsError: true,
				}, nil
			}

			ctx, timeoutCancel := withToolTimeout(ctx)
			defer timeoutCancel()

			tabCtx, cancel := b.NewTab(ctx)
			defer cancel()

			var links []link

			err = chromedp.Run(
				tabCtx,
				chromedp.Navigate(p.URL),
				chromedp.WaitReady("body"),
				chromedp.Evaluate(
					`Array.from(document.querySelectorAll("a[href]")).map(a => ({
						href: a.href,
						text: a.innerText.trim().substring(0, 200)
					}))`,
					&links,
				),
			)
			if err != nil {
				return agent.ToolResult{
					Content: fmt.Sprintf("cannot extract links from %s: %s", p.URL, err),
					IsError: true,
				}, nil
			}

			data, _ := json.Marshal(links)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
