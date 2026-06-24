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
	clickParams struct {
		URL      string `json:"url" jsonschema:"The URL to navigate to before clicking"`
		Selector string `json:"selector" jsonschema:"CSS selector of the element to click (e.g. button.next, a[href*=page])"`
	}
)

func ClickElementTool(b *Browser) agent.Tool {
	return agent.FunctionTool(
		"click_element",
		"Navigate to a URL, click an element matching a CSS selector, and return the page text after the click. Useful for pagination buttons, 'show all' links, tabs, and other interactive elements.",
		func(ctx context.Context, p clickParams) (agent.ToolResult, error) {
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

			var (
				text         string
				postClickURL string
			)

			err := chromedp.Run(
				tabCtx,
				chromedp.Navigate(p.URL),
				waitForPage(),
				chromedp.WaitVisible(p.Selector),
				chromedp.Click(p.Selector),
				waitForPage(),
				chromedp.Location(&postClickURL),
				chromedp.Evaluate(`document.body.innerText`, &text),
			)
			if err != nil {
				return agent.ResultError(b.classifyError(ctx, p.URL, err)), nil
			}

			// Revalidate the post-click URL: a click may navigate
			// the page to a different host (redirect, JS navigation,
			// <a href>), bypassing the initial checkURL. Reject the
			// result if the new URL is outside the allowed scope or
			// resolves to a non-public IP.
			if postClickURL != "" && postClickURL != p.URL {
				if r := b.checkURL(postClickURL); r != nil {
					return *r, nil
				}
			}

			runes := []rune(text)
			if len(runes) > maxTextLength {
				text = string(runes[:maxTextLength])
			}

			return agent.ToolResult{Content: text}, nil
		},
	)
}
