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
