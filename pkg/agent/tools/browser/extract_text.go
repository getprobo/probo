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
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

const (
	maxTextLength = 32000
)

type (
	extractTextParams struct {
		URL string `json:"url" jsonschema:"The URL to extract text from"`
	}
)

func ExtractPageTextTool(b *Browser) agent.Tool {
	return agent.FunctionTool(
		"extract_page_text",
		"Navigate to a URL and extract the visible text content of the page, truncated to 32000 characters.",
		func(ctx context.Context, p extractTextParams) (agent.ToolResult, error) {
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

			var text string

			// Cap the JS-side slice at 4 code units per rune so the
			// DevTools transfer stays bounded even for huge pages;
			// the Go-side rune truncation below then produces the
			// final exact-length output.
			jsMaxLen := maxTextLength * 4
			extractJS := fmt.Sprintf(
				`String(document.body?.innerText ?? '').slice(0, %d)`,
				jsMaxLen,
			)

			err := chromedp.Run(
				tabCtx,
				chromedp.Navigate(p.URL),
				waitForPage(),
				// Scroll to bottom to trigger lazy-loaded content,
				// then back to top and wait briefly for rendering.
				chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
				chromedp.Sleep(500*time.Millisecond),
				chromedp.Evaluate(`window.scrollTo(0, 0)`, nil),
				chromedp.Sleep(200*time.Millisecond),
				chromedp.Evaluate(extractJS, &text),
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
