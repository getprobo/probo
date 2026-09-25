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

package console_v1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
)

func TestResolveTallySettingsWith(t *testing.T) {
	t.Parallel()

	newResolver := func() *Resolver {
		return &Resolver{
			providerRegistry: provider.NewBuiltinRegistry(),
			logger:           log.NewLogger(log.WithOutput(io.Discard)),
		}
	}

	fetchUser := func(user *drivers.TallyCurrentUser, err error) tallyUserFetcher {
		return func(context.Context, *http.Client, string) (*drivers.TallyCurrentUser, error) {
			return user, err
		}
	}

	assertCode := func(t *testing.T, err error, code string) {
		t.Helper()

		require.Error(t, err)

		gqlErr, ok := err.(*gqlerror.Error)
		require.True(t, ok, "expected *gqlerror.Error, got %T", err)
		assert.Equal(t, code, gqlErr.Extensions["code"])
	}

	t.Run("rejected key is invalid", func(t *testing.T) {
		t.Parallel()

		_, err := newResolver().resolveTallySettingsWith(
			context.Background(),
			"bad-key",
			fetchUser(nil, drivers.ErrTallyUnauthorized),
		)
		assertCode(t, err, "INVALID")
	})

	t.Run("transient fetch failure is internal", func(t *testing.T) {
		t.Parallel()

		_, err := newResolver().resolveTallySettingsWith(
			context.Background(),
			"key",
			fetchUser(nil, errors.New("boom")),
		)
		assertCode(t, err, "INTERNAL")
	})

	t.Run("missing organization id is internal", func(t *testing.T) {
		t.Parallel()

		_, err := newResolver().resolveTallySettingsWith(
			context.Background(),
			"key",
			fetchUser(&drivers.TallyCurrentUser{}, nil),
		)
		assertCode(t, err, "INTERNAL")
	})

	t.Run("derives the organization id from the key", func(t *testing.T) {
		t.Parallel()

		raw, err := newResolver().resolveTallySettingsWith(
			context.Background(),
			"key",
			fetchUser(&drivers.TallyCurrentUser{OrganizationID: "wvBzxD"}, nil),
		)
		require.NoError(t, err)
		assert.JSONEq(t, `{"organization_id":"wvBzxD"}`, string(raw))
	})
}
