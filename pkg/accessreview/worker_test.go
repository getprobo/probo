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

package accessreview

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.probo.inc/probo/pkg/accessreview/drivers"
)

func TestFetchFailureMessage(t *testing.T) {
	t.Parallel()

	t.Run("refused setting shows its own explanation", func(t *testing.T) {
		t.Parallel()

		rejected := &drivers.SettingRejectedError{
			Code:       drivers.BetterStackTeamNotFound,
			StatusCode: http.StatusUnprocessableEntity,
			Message:    "Better Stack has no team with this name for this API token.",
		}
		err := fmt.Errorf("cannot list accounts from source Better Stack / Acme: %w", rejected)

		assert.Equal(t, rejected.Message, fetchFailureMessage(err))
	})

	t.Run("anything else stays generic", func(t *testing.T) {
		t.Parallel()

		err := errors.New("cannot fetch supabase members: unexpected status 500")

		assert.Equal(t, sourceFetchFailureMessage, fetchFailureMessage(err))
	})
}
