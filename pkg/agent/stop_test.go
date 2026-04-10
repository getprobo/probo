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

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopSignal(t *testing.T) {
	t.Parallel()

	t.Run("absent returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, stopSignalFrom(context.Background()))
	})

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()
		ch := make(chan struct{})
		ctx := WithStopSignal(context.Background(), ch)
		assert.Equal(t, (<-chan struct{})(ch), stopSignalFrom(ctx))
	})

	t.Run("non-blocking when open", func(t *testing.T) {
		t.Parallel()
		ch := make(chan struct{})
		ctx := WithStopSignal(context.Background(), ch)
		sig := stopSignalFrom(ctx)
		select {
		case <-sig:
			t.Fatal("should not fire")
		default:
		}
	})

	t.Run("fires when closed", func(t *testing.T) {
		t.Parallel()
		ch := make(chan struct{})
		ctx := WithStopSignal(context.Background(), ch)
		close(ch)
		sig := stopSignalFrom(ctx)
		select {
		case <-sig:
		default:
			t.Fatal("should have fired")
		}
	})
}
