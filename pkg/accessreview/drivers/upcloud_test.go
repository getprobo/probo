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
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

func TestUpCloudDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/upcloud", "UPCLOUD_API_KEY")
	// UpCloud authenticates with a Bearer API token. The matcher ignores
	// Authorization, so replay needs no auth.
	client := newVCRClient(rec, bearerAuth(os.Getenv("UPCLOUD_API_KEY")))

	driver := NewUpCloudDriver(client, log.NewLogger(log.WithName("test")), "https://api.upcloud.com/1.3")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	main := records[0]
	assert.Equal(t, "test", main.ExternalID)
	assert.Equal(t, "Main Account", main.FullName)
	assert.Equal(t, "main@example.com", main.Email)
	assert.Equal(t, []string{"technical"}, main.Roles)
	assert.Equal(t, new(true), main.IsAdmin)
	assert.Nil(t, main.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, main.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, main.AccountType)

	sub := records[1]
	assert.Equal(t, "my_sub_account", sub.ExternalID)
	assert.Equal(t, "Sub Account", sub.FullName)
	assert.Equal(t, "sub@example.com", sub.Email)
	assert.Equal(t, []string{"technical"}, sub.Roles)
	assert.Equal(t, new(false), sub.IsAdmin)

	// no roles assigned; details answers 403, which is a stable "no details"
	// rather than an error, so the record keeps its list-only fields.
	temp := records[2]
	assert.Equal(t, "my_temp_account", temp.ExternalID)
	assert.Equal(t, "my_temp_account", temp.FullName)
	assert.Empty(t, temp.Email)
	assert.Equal(t, []string{}, temp.Roles)
	assert.Equal(t, new(false), temp.IsAdmin)

	billing := records[3]
	assert.Equal(t, "my_billing_account", billing.ExternalID)
	assert.Equal(t, "Billing Account", billing.FullName)
	assert.Equal(t, "billing@example.com", billing.Email)
	assert.Equal(t, []string{"billing"}, billing.Roles)
	assert.Equal(t, new(false), billing.IsAdmin)
}

// TestUpCloudDriverContextCancellation verifies that a context canceled
// mid-run aborts ListAccounts with the cancellation error instead of being
// swallowed as a best-effort per-account detail failure, which would let a
// caller mistake a truncated run for a complete, successful sync.
func TestUpCloudDriverContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "/account/details/") {
				cancel()

				return nil, ctx.Err()
			}

			body := `{"accounts":{"account":[{"labels":[],"roles":{"role":["technical"]},"type":"mymain","username":"test"}]}}`

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}

	driver := NewUpCloudDriver(client, log.NewLogger(log.WithName("test")), "https://api.upcloud.com/1.3")
	records, err := driver.ListAccounts(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Nil(t, records)
}

func TestUpCloudIsMainAccount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		accountType string
		want        bool
	}{
		{accountType: "main", want: true},   // live API
		{accountType: "mymain", want: true}, // published docs example
		{accountType: "MAIN", want: true},
		{accountType: "sub", want: false},
		{accountType: "SUB", want: false},
		{accountType: " sub ", want: false},
		{accountType: "", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.accountType, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, upcloudIsMainAccount(tc.accountType))
		})
	}
}

func TestUpCloudFetchAccountDetails(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		status    int
		wantNil   bool
		wantError bool
	}{
		{name: "forbidden yields no details", status: http.StatusForbidden, wantNil: true},
		{name: "not found yields no details", status: http.StatusNotFound, wantNil: true},
		{name: "server error is fatal", status: http.StatusInternalServerError, wantError: true},
		{name: "rate limited is fatal", status: http.StatusTooManyRequests, wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(strings.NewReader(`{"error":{"error_code":"X"}}`)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			})}

			driver := NewUpCloudDriver(client, log.NewLogger(log.WithName("test")), "https://api.upcloud.com/1.3")

			details, err := driver.fetchAccountDetails(context.Background(), "someone")
			if tc.wantError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantNil, details == nil)
		})
	}
}

// TestUpCloudDriverTransientDetailFailureAborts pins the identity guarantee:
// the review keys accounts on email plus external ID, so a record emitted
// with a blank email after a transient failure would read as one account
// removed and another added.
func TestUpCloudDriverTransientDetailFailureAborts(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"accounts":{"account":[{"roles":{"role":["technical"]},"type":"main","username":"test"}]}}`
		status := http.StatusOK

		if strings.Contains(req.URL.Path, "/account/details/") {
			body = `{"error":{"error_code":"INTERNAL"}}`
			status = http.StatusInternalServerError
		}

		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}

	records, err := NewUpCloudDriver(client, log.NewLogger(log.WithName("test")), "https://api.upcloud.com/1.3").ListAccounts(context.Background())
	require.Error(t, err)
	assert.Nil(t, records)
}

// TestUpCloudDriverBlankUsernameAborts pins the malformed-list guarantee:
// username is the only stable identifier, so silently dropping a blank row
// would hide an account the source still exposes and mark it removed on the
// next review.
func TestUpCloudDriverBlankUsernameAborts(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "/account/details/") {
			t.Errorf("driver must not fetch details after a malformed list row")
		}

		body := `{"accounts":{"account":[{"roles":{"role":[]},"type":"sub","username":"  "}]}}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}

	records, err := NewUpCloudDriver(client, log.NewLogger(log.WithName("test")), "https://api.upcloud.com/1.3").ListAccounts(context.Background())
	require.Error(t, err)
	assert.Nil(t, records)
}
