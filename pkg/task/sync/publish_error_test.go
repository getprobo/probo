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

package tasksync_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	tasksync "go.probo.inc/probo/pkg/task/sync"
)

func TestLinearPublishClientMessage(t *testing.T) {
	t.Parallel()

	t.Run("known failure keeps its message", func(t *testing.T) {
		t.Parallel()

		message, known := tasksync.LinearPublishClientMessage(
			fmt.Errorf("cannot publish: %w", tasksync.ErrLinearTeamNotFound),
		)

		assert.True(t, known)
		assert.Equal(t, tasksync.ErrLinearTeamNotFound.Error(), message)
	})

	t.Run("conflict maps to already linked", func(t *testing.T) {
		t.Parallel()

		message, known := tasksync.LinearPublishClientMessage(
			fmt.Errorf("dial tcp: %w", coredata.ErrResourceAlreadyExists),
		)

		assert.True(t, known)
		assert.Equal(t, tasksync.ErrTaskAlreadyLinked.Error(), message)
	})

	t.Run("internal failure does not leak", func(t *testing.T) {
		t.Parallel()

		message, known := tasksync.LinearPublishClientMessage(
			fmt.Errorf("linear api 500: secret token"),
		)

		assert.False(t, known)
		assert.Equal(t, "could not publish the task to Linear", message)
		assert.NotContains(t, message, "secret")
	})
}
