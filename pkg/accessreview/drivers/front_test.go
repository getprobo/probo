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
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

const frontTestBaseURL = "https://api2.frontapp.com"

func TestFrontDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/front", "FRONT_API_KEY")
	// Front accepts both an OAuth access token and a company API token as a
	// Bearer credential. The matcher ignores Authorization, so replay needs none.
	client := newVCRClient(rec, bearerAuth(os.Getenv("FRONT_API_KEY")))

	records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 6)

	seen := make(map[string]bool, len(records))

	for _, record := range records {
		require.NotEmpty(t, record.ExternalID, "every record must carry a stable external ID")
		require.False(t, seen[record.ExternalID], "external IDs must be unique: %q", record.ExternalID)
		seen[record.ExternalID] = true

		assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus, "front exposes no MFA state")
		assert.Nil(t, record.LastLogin, "front exposes no last-login")
		assert.Nil(t, record.CreatedAt, "front exposes no creation timestamp")
	}

	admin := records[0]
	assert.Equal(t, "tea_1", admin.ExternalID)
	assert.Equal(t, "alice@example.com", admin.Email)
	assert.Equal(t, "Alice Admin", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, admin.AuthMethod)

	member := records[1]
	assert.Equal(t, "tea_2", member.ExternalID)
	assert.Equal(t, []string{"Teammate"}, member.Roles)
	assert.False(t, member.IsAdmin)
	require.NotNil(t, member.Active)
	// is_available false is a presence toggle, not an account state.
	assert.True(t, *member.Active)

	// is_blocked true is Front's only account-status signal, and with no
	// first/last name the display name falls back to the mention username.
	blocked := records[2]
	assert.Equal(t, "tea_3", blocked.ExternalID)
	assert.Equal(t, "carol_offboarded", blocked.FullName)
	require.NotNil(t, blocked.Active)
	assert.False(t, *blocked.Active)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, blocked.AccountType)

	// A rule bot holds access too; its type qualifies the grant.
	rule := records[3]
	assert.Equal(t, "tea_4", rule.ExternalID)
	assert.Equal(t, "triage_rule", rule.FullName)
	assert.Empty(t, rule.Email)
	assert.Equal(t, []string{"Teammate", "Type: Rule"}, rule.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, rule.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodServiceAccount, rule.AuthMethod)

	// Page 2, reached through _pagination.next.
	oauthClient := records[4]
	assert.Equal(t, "tea_5", oauthClient.ExternalID)
	assert.Equal(t, []string{"Teammate", "Type: API"}, oauthClient.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, oauthClient.AccountType)

	application := records[5]
	assert.Equal(t, "tea_6", application.ExternalID)
	assert.Equal(t, []string{"Admin", "Type: Application"}, application.Roles)
	assert.True(t, application.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, application.AccountType)
}

// TestFrontDriverTransientFailureAborts pins the completeness guarantee: a
// short answer returned as a success would mark every missing teammate removed
// on the next campaign.
func TestFrontDriverTransientFailureAborts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		body   string
	}{
		{name: "server error", status: http.StatusInternalServerError, body: `{"_error":{"status":500}}`},
		{name: "malformed body", status: http.StatusOK, body: `{"_results":`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(strings.NewReader(tc.body)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			})}

			records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(context.Background())
			require.Error(t, err)
			assert.Nil(t, records)
		})
	}
}

// TestFrontDriverContextCancellation verifies a cancelled context surfaces as
// the cancellation error rather than a truncated success.
func TestFrontDriverContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		cancel()

		return nil, ctx.Err()
	})}

	records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Nil(t, records)
}

// TestFrontDriverEmptyPageEndsWalk covers a cursor that never clears: Front
// keeps _pagination on every response, so an empty page must end the walk
// instead of spinning until the guard trips.
func TestFrontDriverEmptyPageEndsWalk(t *testing.T) {
	t.Parallel()

	var requests int

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests++

		body := `{"_pagination":{"next":"https://api2.frontapp.com/teammates?page_token=always"},"_results":[]}`
		if requests == 1 {
			body = `{"_pagination":{"next":"https://api2.frontapp.com/teammates?page_token=always"},"_results":[{"id":"tea_1","email":"alice@example.com","first_name":"Alice","last_name":"Admin","type":"user"}]}`
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}

	records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, 2, requests, "the empty second page must end the walk")
}

// TestFrontDriverPaginationGuard covers a cursor that keeps returning results:
// the walk stops at maxPaginationPages with an explicit error rather than
// looping forever.
func TestFrontDriverPaginationGuard(t *testing.T) {
	t.Parallel()

	var requests int

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests++

		body := fmt.Sprintf(
			`{"_pagination":{"next":"https://api2.frontapp.com/teammates?page_token=p%d"},"_results":[{"id":"tea_%d","email":"user%d@example.com","type":"user"}]}`,
			requests, requests, requests,
		)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}

	records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPaginationLimitReached)
	assert.Nil(t, records)
	assert.Equal(t, maxPaginationPages, requests)
}

func TestFrontAccountTypeAndAuthMethod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		accountType string
		want        coredata.AccessReviewEntryAccountType
	}{
		{accountType: "user", want: coredata.AccessReviewEntryAccountTypeUser},
		{accountType: "visitor", want: coredata.AccessReviewEntryAccountTypeUser},
		// Unset and unrecognised types fall to USER, keeping a real person in
		// the human-review path.
		{accountType: "", want: coredata.AccessReviewEntryAccountTypeUser},
		{accountType: "something_new", want: coredata.AccessReviewEntryAccountTypeUser},
		{accountType: "ai", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "api", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "APPLICATION", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "bulk_reply", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "csat", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "integration", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "macro", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "rule", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{accountType: "smart_csat", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
	}

	for _, tc := range cases {
		t.Run(tc.accountType, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, frontAccountType(tc.accountType))

			wantAuth := coredata.AccessReviewEntryAuthMethodUnknown
			if tc.want == coredata.AccessReviewEntryAccountTypeServiceAccount {
				wantAuth = coredata.AccessReviewEntryAuthMethodServiceAccount
			}

			assert.Equal(t, wantAuth, frontAuthMethod(tc.accountType))
		})
	}
}

func TestFrontNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("resolves the company name", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "/me", req.URL.Path)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"id":"cmp_k30","name":"Dunder Mifflin, Inc."}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})}

		name, err := NewFrontNameResolver(client, frontTestBaseURL).ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "Dunder Mifflin, Inc.", name)
	})

	t.Run("unauthorized is terminal", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})}

		name, err := NewFrontNameResolver(client, frontTestBaseURL).ResolveInstanceName(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTerminalNameResolution)
		assert.Empty(t, name)
	})
}
