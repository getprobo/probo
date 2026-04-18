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

package search

import (
	"go.probo.inc/probo/pkg/agent"
)

// Toolset provides web search tools.
type Toolset struct {
	endpoint string
}

// NewToolset creates a search toolset with the given SearXNG endpoint.
func NewToolset(endpoint string) *Toolset {
	return &Toolset{endpoint: endpoint}
}

func (t *Toolset) Tools() []agent.Tool {
	return []agent.Tool{
		WebSearchTool(t.endpoint),
		CheckGovernmentDBTool(t.endpoint),
		CheckWaybackTool(),
		DiffDocumentsTool(),
	}
}
