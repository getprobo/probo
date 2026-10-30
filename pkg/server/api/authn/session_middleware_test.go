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

package authn

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/puddle/v2"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreSessionDataUpdateError(t *testing.T) {
	t.Parallel()

	closedPool := fmt.Errorf("cannot acquire connection: %w", puddle.ErrClosedPool)
	database := errors.New("cannot update session")

	t.Run("closed pool during shutdown", func(t *testing.T) {
		t.Parallel()

		assert.True(t, ignoreSessionDataUpdateError(false, closedPool))
	})

	t.Run("context canceled before the write", func(t *testing.T) {
		t.Parallel()

		assert.True(t, ignoreSessionDataUpdateError(true, database))
	})

	t.Run("context-caused write error", func(t *testing.T) {
		t.Parallel()

		wrapped := fmt.Errorf("cannot update session: %w", context.Canceled)

		assert.True(t, ignoreSessionDataUpdateError(false, wrapped))
	})

	t.Run("database error on a live request", func(t *testing.T) {
		t.Parallel()

		assert.False(t, ignoreSessionDataUpdateError(false, database))
	})
}
