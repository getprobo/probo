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

package oauth2_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam/oauth2"
)

func TestCreateManualAccessToken_ExpiryLimit(t *testing.T) {
	t.Parallel()

	svc := oauth2.NewService(nil, nil, "", log.NewLogger())
	ctx := context.Background()

	t.Run(
		"rejects a century expiry",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := svc.CreateManualAccessToken(
				ctx,
				&oauth2.CreateManualAccessTokenRequest{
					Name:      "century",
					ExpiresAt: time.Now().Add(100 * 365 * 24 * time.Hour),
				},
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, oauth2.ErrInvalidRequest)
			assert.ErrorContains(t, err, "expires_at must not be more than 180 days in the future")
		},
	)

	t.Run(
		"rejects an expiry past 180 days",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := svc.CreateManualAccessToken(
				ctx,
				&oauth2.CreateManualAccessTokenRequest{
					Name:      "too-long",
					ExpiresAt: time.Now().Add(181 * 24 * time.Hour),
				},
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, oauth2.ErrInvalidRequest)
			assert.ErrorContains(t, err, "expires_at must not be more than 180 days in the future")
		},
	)

	t.Run(
		"accepts an expiry of 180 days",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := svc.CreateManualAccessToken(
				ctx,
				&oauth2.CreateManualAccessTokenRequest{
					Name:      "max",
					ExpiresAt: time.Now().Add(180 * 24 * time.Hour),
				},
			)

			require.Error(t, err)
			assert.ErrorContains(t, err, "scopes are required")
			assert.NotContains(t, err.Error(), "180 days")
		},
	)

	t.Run(
		"rejects an expiry in the past",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := svc.CreateManualAccessToken(
				ctx,
				&oauth2.CreateManualAccessTokenRequest{
					Name:      "past",
					ExpiresAt: time.Now().Add(-time.Hour),
				},
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, oauth2.ErrInvalidRequest)
			assert.ErrorContains(t, err, "expires_at must be in the future")
		},
	)
}
