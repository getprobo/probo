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

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

type (
	selectParams struct {
		URL      string `json:"url" jsonschema:"The URL to navigate to before selecting"`
		Selector string `json:"selector" jsonschema:"CSS selector of the select element"`
		Value    string `json:"value" jsonschema:"The option value to select"`
	}
)

func SelectOptionTool(b *Browser) (agent.Tool, error) {
	return agent.FunctionTool(
		"select_option",
		"Navigate to a URL, select an option from a <select> dropdown, and return the page text after selection. Useful for changing page size dropdowns (e.g. 'show 100 per page').",
		func(ctx context.Context, p selectParams) (agent.ToolResult, error) {
			if r := b.checkAlive(); r != nil {
				return *r, nil
			}

			if r := b.checkURL(p.URL); r != nil {
				return *r, nil
			}

			ctx, timeoutCancel := withToolTimeout(ctx)
			defer timeoutCancel()

			tabCtx, cancel := b.NewTab(ctx)
			defer cancel()

			var text string

			err := chromedp.Run(
				tabCtx,
				chromedp.Navigate(p.URL),
				waitForPage(),
				chromedp.WaitVisible(p.Selector),
				chromedp.SetValue(p.Selector, p.Value),
				chromedp.Evaluate(
					fmt.Sprintf(
						`document.querySelector(%q).dispatchEvent(new Event('change', {bubbles: true}))`,
						p.Selector,
					),
					nil,
				),
				waitForPage(),
				chromedp.Evaluate(`document.body.innerText`, &text),
			)
			if err != nil {
				return agent.ResultError(b.classifyError(ctx, p.URL, err)), nil
			}

			runes := []rune(text)
			if len(runes) > maxTextLength {
				text = string(runes[:maxTextLength])
			}

			return agent.ToolResult{Content: text}, nil
		},
	)
}
