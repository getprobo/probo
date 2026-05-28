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

package coredata

import "testing"

func TestNormalizeSCIMBridgeState(t *testing.T) {
	t.Parallel()

	t.Run("keeps valid state", func(t *testing.T) {
		t.Parallel()

		state := NormalizeSCIMBridgeState(SCIMBridgeStateActive, nil)
		if state != SCIMBridgeStateActive {
			t.Fatalf("expected %q, got %q", SCIMBridgeStateActive, state)
		}
	})

	t.Run("maps invalid state with sync error to failed", func(t *testing.T) {
		t.Parallel()

		syncErr := "sync failed"
		state := NormalizeSCIMBridgeState(SCIMBridgeState(":"), &syncErr)
		if state != SCIMBridgeStateFailed {
			t.Fatalf("expected %q, got %q", SCIMBridgeStateFailed, state)
		}
	})

	t.Run("maps invalid state without sync error to pending", func(t *testing.T) {
		t.Parallel()

		state := NormalizeSCIMBridgeState(SCIMBridgeState(":"), nil)
		if state != SCIMBridgeStatePending {
			t.Fatalf("expected %q, got %q", SCIMBridgeStatePending, state)
		}
	})
}
