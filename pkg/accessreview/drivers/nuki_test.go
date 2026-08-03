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
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

// nukiTestBaseURL mirrors the APIBase declared in nukiRegistration, so the
// cassette's recorded URLs keep matching.
const nukiTestBaseURL = "https://api.nuki.io"

func TestNukiDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/nuki", "NUKI_API_KEY")
	// Nuki Web API tokens authenticate as a Bearer token. The matcher ignores
	// Authorization, so replay needs no auth.
	client := newVCRClient(rec, bearerAuth(os.Getenv("NUKI_API_KEY")))

	driver := NewNukiDriver(client, nukiTestBaseURL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	person := records[0]
	assert.Equal(t, "1859620430", person.ExternalID)
	assert.Equal(t, "alice.martin@example.com", person.Email)
	assert.Equal(t, "Alice Martin", person.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, person.AccountType)
	assert.Empty(t, person.Roles)
	assert.False(t, person.IsAdmin)
	// Nuki has no deactivated state for an account user, so Active carries no
	// signal rather than a fabricated true.
	assert.Nil(t, person.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, person.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, person.AuthMethod)
	assert.Nil(t, person.LastLogin)
	require.NotNil(t, person.CreatedAt)
	assert.Equal(t, "2025-05-12T14:45:40Z", person.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))

	// type=1 marks a company account user: a non-human key holder.
	company := records[1]
	assert.Equal(t, "718259920", company.ExternalID)
	assert.Equal(t, "Cleaning Co", company.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, company.AccountType)

	// No name on the row; the email is the fallback display name.
	nameless := records[2]
	assert.Equal(t, "568452711", nameless.ExternalID)
	assert.Equal(t, "nameless@example.com", nameless.Email)
	assert.Equal(t, "nameless@example.com", nameless.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, nameless.AccountType)
}

// TestNukiDriverPaginatesPastShortPage pins the pagination contract against the
// tempting "stop when the page is shorter than the requested limit" shortcut.
// Nuki clamps `limit` to an undocumented maximum, so a short page can still be
// a full page; stopping on one would silently truncate the review and report
// every dropped account as removed on the next campaign.
func TestNukiDriverPaginatesPastShortPage(t *testing.T) {
	t.Parallel()

	var offsets []string

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		offset := req.URL.Query().Get("offset")
		offsets = append(offsets, offset)

		// The server clamps the requested limit=100 down to 2 per page.
		body := `[]`

		switch offset {
		case "0":
			body = `[{"accountUserId":1,"email":"one@example.com","name":"One"},` +
				`{"accountUserId":2,"email":"two@example.com","name":"Two"}]`
		case "2":
			body = `[{"accountUserId":3,"email":"three@example.com","name":"Three"}]`
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}

	records, err := NewNukiDriver(client, nukiTestBaseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)
	assert.Equal(t, []string{"0", "2", "3"}, offsets)
	assert.Equal(t, "three@example.com", records[2].Email)
}

// TestNukiDriverPaginationLimit verifies that an API which never returns an
// empty page fails loudly instead of looping forever.
func TestNukiDriverPaginationLimit(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Always a full-looking page, and every row is distinct so the walk has
		// no other reason to stop.
		offset := req.URL.Query().Get("offset")
		id, _ := strconv.Atoi(offset)

		body := `[{"accountUserId":` + strconv.Itoa(id+1) + `,"email":"a` + offset + `@example.com","name":"A"}]`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}

	records, err := NewNukiDriver(client, nukiTestBaseURL).ListAccounts(context.Background())
	require.ErrorIs(t, err, ErrPaginationLimitReached)
	assert.Nil(t, records)
}

func TestNukiDriverErrorStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: status,
					Body:       io.NopCloser(strings.NewReader(``)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			})}

			records, err := NewNukiDriver(client, nukiTestBaseURL).ListAccounts(context.Background())
			require.Error(t, err)
			assert.Nil(t, records)
		})
	}
}

// TestNukiAccountRecord covers the identity guard and the field fallbacks
// without going through the HTTP layer.
func TestNukiAccountRecord(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		user           nukiAccountUser
		wantError      bool
		wantExternalID string
		wantFullName   string
		wantType       coredata.AccessReviewEntryAccountType
	}{
		{
			name:           "person with name",
			user:           nukiAccountUser{AccountUserID: 42, Email: "a@example.com", Name: "Ada"},
			wantExternalID: "42",
			wantFullName:   "Ada",
			wantType:       coredata.AccessReviewEntryAccountTypeUser,
		},
		{
			name:           "blank name falls back to email",
			user:           nukiAccountUser{AccountUserID: 42, Email: "a@example.com", Name: "   "},
			wantExternalID: "42",
			wantFullName:   "a@example.com",
			wantType:       coredata.AccessReviewEntryAccountTypeUser,
		},
		{
			name:           "no email falls back to the id for the display name",
			user:           nukiAccountUser{AccountUserID: 42},
			wantExternalID: "42",
			wantFullName:   "42",
			wantType:       coredata.AccessReviewEntryAccountTypeUser,
		},
		{
			name:           "company",
			user:           nukiAccountUser{AccountUserID: 7, Email: "c@example.com", Name: "Co", Type: 1},
			wantExternalID: "7",
			wantFullName:   "Co",
			wantType:       coredata.AccessReviewEntryAccountTypeServiceAccount,
		},
		{
			name:           "email-only row keeps its identity",
			user:           nukiAccountUser{Email: "a@example.com", Name: "Ada"},
			wantExternalID: "",
			wantFullName:   "Ada",
			wantType:       coredata.AccessReviewEntryAccountTypeUser,
		},
		{
			// Neither an email nor an id: the review would key this row the
			// same as every other identity-less row, so it must not be emitted.
			name:      "row without any identity is rejected",
			user:      nukiAccountUser{Name: "Ada"},
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			record, err := nukiAccountRecord(tc.user)
			if tc.wantError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantExternalID, record.ExternalID)
			assert.Equal(t, tc.wantFullName, record.FullName)
			assert.Equal(t, tc.wantType, record.AccountType)
		})
	}
}
