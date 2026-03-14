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
	"strings"

	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/agent"
)

type Browser struct {
	addr     string
	allocCtx context.Context
	cancel   context.CancelFunc
}

func NewBrowser(ctx context.Context, addr string) *Browser {
	if !strings.HasPrefix(addr, "ws://") {
		addr = "ws://" + addr
	}

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, addr)

	return &Browser{
		addr:     addr,
		allocCtx: allocCtx,
		cancel:   cancel,
	}
}

func (b *Browser) NewTab(ctx context.Context) (context.Context, context.CancelFunc) {
	tabCtx, tabCancel := chromedp.NewContext(b.allocCtx)

	// Propagate the caller's cancellation to the Chrome tab so that
	// tool-level timeouts and context deadlines actually stop the browser.
	go func() {
		select {
		case <-ctx.Done():
			tabCancel()
		case <-tabCtx.Done():
		}
	}()

	return tabCtx, tabCancel
}

func (b *Browser) Close() {
	b.cancel()
}

func BuildTools(b *Browser) ([]agent.Tool, error) {
	navigateTool, err := NavigateToURLTool(b)
	if err != nil {
		return nil, err
	}

	extractTextTool, err := ExtractPageTextTool(b)
	if err != nil {
		return nil, err
	}

	extractLinksTool, err := ExtractLinksTool(b)
	if err != nil {
		return nil, err
	}

	findLinksTool, err := FindLinksMatchingTool(b)
	if err != nil {
		return nil, err
	}

	return []agent.Tool{
		navigateTool,
		extractTextTool,
		extractLinksTool,
		findLinksTool,
	}, nil
}
