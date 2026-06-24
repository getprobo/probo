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

package drivers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleWorkspaceDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/google_workspace", "GOOGLE_WORKSPACE_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("GOOGLE_WORKSPACE_TOKEN")))

	driver := NewGoogleWorkspaceDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.Roles)
}

func TestGoogleWorkspaceNameResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{
			name:   "200 returns customer domain",
			status: http.StatusOK,
			body:   `{"kind":"admin#directory#customer","customerDomain":"example.com"}`,
			want:   "example.com",
		},
		{
			name:   "403 is terminal (no error, no name)",
			status: http.StatusForbidden,
			body:   `{"error":{"code":403,"message":"Not Authorized to access this resource/api","status":"FORBIDDEN"}}`,
			want:   "",
		},
		{
			name:    "500 is retryable",
			status:  http.StatusInternalServerError,
			body:    `{"error":{"code":500,"message":"Internal Server Error","status":"INTERNAL"}}`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/admin/directory/v1/customers/my_customer", r.URL.Path)
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(tc.status)
						_, _ = w.Write([]byte(tc.body))
					},
				),
			)
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewGoogleWorkspaceNameResolver(client).ResolveInstanceName(t.Context())
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
