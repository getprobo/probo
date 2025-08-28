// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package llmgw

import (
	"context"
	"fmt"
)

type (
	Gateway struct {
		providers map[string]Provider
	}

	ErrProviderNotFound struct {
		ProviderName string
	}
)

func (e *ErrProviderNotFound) Error() string {
	return fmt.Sprintf("provider %s not found", e.ProviderName)
}

func NewGateway(providers map[string]Provider) *Gateway {
	return &Gateway{providers: providers}
}

func (g *Gateway) Generate(ctx context.Context, providerName string, req GenerateRequest) (*GenerateResponse, error) {
	provider, ok := g.providers[providerName]
	if !ok {
		return nil, &ErrProviderNotFound{ProviderName: providerName}
	}
	return provider.Generate(ctx, req)
}

func (g *Gateway) Chat(ctx context.Context, providerName string, req ChatRequest) (*ChatResponse, error) {
	provider, ok := g.providers[providerName]
	if !ok {
		return nil, &ErrProviderNotFound{ProviderName: providerName}
	}
	return provider.Chat(ctx, req)
}
