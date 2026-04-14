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

package agent

import "context"

type stopSignalKey struct{}

// WithStopSignal returns a derived context carrying a stop channel.
// The agent loop checks this channel at each turn boundary.
func WithStopSignal(ctx context.Context, ch <-chan struct{}) context.Context {
	return context.WithValue(ctx, stopSignalKey{}, ch)
}

// stopSignalFrom retrieves the stop channel from the context, or nil.
func stopSignalFrom(ctx context.Context) <-chan struct{} {
	ch, _ := ctx.Value(stopSignalKey{}).(<-chan struct{})
	return ch
}
