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

package provider

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

func TestProbeCloudflare(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderCloudflare}
	endpoints := Endpoints{Probe: "https://api.cloudflare.com/client/v4/user/tokens/verify"}

	t.Run("active token is connected", func(t *testing.T) {
		t.Parallel()

		var seen []*http.Request

		client := probeStubClient(&seen, http.StatusOK, `{"success":true,"result":{"status":"active"}}`)

		err := probeCloudflare(t.Context(), client, conn, endpoints)

		require.NoError(t, err)
		require.Len(t, seen, 1)
		assert.Equal(t, endpoints.Probe, seen[0].URL.String())
	})

	t.Run("malformed token is a 400 rejection", func(t *testing.T) {
		t.Parallel()

		client := probeStubClient(new([]*http.Request), http.StatusBadRequest, `{"success":false}`)

		err := probeCloudflare(t.Context(), client, conn, endpoints)

		rejected, ok := errors.AsType[*CredentialRejectedError](err)
		require.True(t, ok, "400 should reject the credential, got %v", err)
		assert.Equal(t, http.StatusBadRequest, rejected.StatusCode)
		assert.False(t, rejected.OperationRefused)
	})

	t.Run("dead token is a 401 rejection", func(t *testing.T) {
		t.Parallel()

		client := probeStubClient(new([]*http.Request), http.StatusUnauthorized, `{"success":false}`)

		err := probeCloudflare(t.Context(), client, conn, endpoints)

		rejected, ok := errors.AsType[*CredentialRejectedError](err)
		require.True(t, ok, "401 should reject the credential, got %v", err)
		assert.Equal(t, http.StatusUnauthorized, rejected.StatusCode)
		assert.False(t, rejected.OperationRefused)
	})

	t.Run("refused call is a 403 rejection", func(t *testing.T) {
		t.Parallel()

		client := probeStubClient(new([]*http.Request), http.StatusForbidden, `{"success":false}`)

		err := probeCloudflare(t.Context(), client, conn, endpoints)

		rejected, ok := errors.AsType[*CredentialRejectedError](err)
		require.True(t, ok, "403 should reject the credential, got %v", err)
		assert.Equal(t, http.StatusForbidden, rejected.StatusCode)
		assert.True(t, rejected.OperationRefused)
	})

	for _, status := range []string{"expired", "disabled"} {
		t.Run(status+" token on 200 is a credential rejection", func(t *testing.T) {
			t.Parallel()

			client := probeStubClient(
				new([]*http.Request),
				http.StatusOK,
				`{"success":true,"result":{"status":"`+status+`"}}`,
			)

			err := probeCloudflare(t.Context(), client, conn, endpoints)

			rejected, ok := errors.AsType[*CredentialRejectedError](err)
			require.True(t, ok, "status %s should reject the credential, got %v", status, err)
			assert.Equal(t, http.StatusUnauthorized, rejected.StatusCode)
			assert.False(t, rejected.OperationRefused)
		})
	}
}
