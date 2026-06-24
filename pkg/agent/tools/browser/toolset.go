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
	"go.probo.inc/probo/pkg/agent"
)

// ReadOnlyToolset provides browser tools that only read page content.
type ReadOnlyToolset struct {
	browser *Browser
}

// NewReadOnlyToolset creates a read-only browser toolset.
func NewReadOnlyToolset(b *Browser) *ReadOnlyToolset {
	return &ReadOnlyToolset{browser: b}
}

func (t *ReadOnlyToolset) Tools() []agent.Tool {
	return []agent.Tool{
		NavigateToURLTool(t.browser),
		ExtractPageTextTool(t.browser),
		ExtractLinksTool(t.browser),
		FindLinksMatchingTool(t.browser),
		FetchRobotsTxtTool(),
		FetchSitemapTool(),
		DownloadPDFTool(),
	}
}

// InteractiveToolset provides all browser tools including click and select.
type InteractiveToolset struct {
	browser *Browser
}

// NewInteractiveToolset creates an interactive browser toolset.
func NewInteractiveToolset(b *Browser) *InteractiveToolset {
	return &InteractiveToolset{browser: b}
}

func (t *InteractiveToolset) Tools() []agent.Tool {
	return []agent.Tool{
		NavigateToURLTool(t.browser),
		ExtractPageTextTool(t.browser),
		ExtractLinksTool(t.browser),
		FindLinksMatchingTool(t.browser),
		ClickElementTool(t.browser),
		SelectOptionTool(t.browser),
		FetchRobotsTxtTool(),
		FetchSitemapTool(),
		DownloadPDFTool(),
	}
}
