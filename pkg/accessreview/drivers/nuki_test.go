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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

const nukiTestBaseURL = "https://api.nuki.io"

func TestNukiDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/nuki", "NUKI_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("NUKI_API_KEY")))

	driver := NewNukiDriver(client, nukiTestBaseURL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	person := records[0]
	assert.Equal(t, "1859620430", person.ExternalID)
	assert.Equal(t, "alice.martin@example.com", person.Email)
	assert.Equal(t, "Alice Martin", person.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, person.AccountType)
	assert.Equal(t, []string{"Office front door", "Remote access", "Server room"}, person.Roles)
	assert.False(t, person.IsAdmin)
	require.NotNil(t, person.Active)
	assert.True(t, *person.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, person.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, person.AuthMethod)
	require.NotNil(t, person.LastLogin)
	assert.Equal(t, "2025-07-01T09:15:00Z", person.LastLogin.UTC().Format("2006-01-02T15:04:05Z"))
	require.NotNil(t, person.CreatedAt)
	assert.Equal(t, "2025-05-12T14:45:40Z", person.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))

	company := records[1]
	assert.Equal(t, "718259920", company.ExternalID)
	assert.Equal(t, "Cleaning Co", company.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, company.AccountType)
	assert.Empty(t, company.Roles)
	assert.Nil(t, company.Active)
	assert.Nil(t, company.LastLogin)

	nameless := records[2]
	assert.Equal(t, "568452711", nameless.ExternalID)
	assert.Equal(t, "nameless@example.com", nameless.Email)
	assert.Equal(t, "nameless@example.com", nameless.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, nameless.AccountType)

	keypad := records[3]
	assert.Equal(t, "auth-keypad-office", keypad.ExternalID)
	assert.Equal(t, "Cleaning keypad", keypad.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, keypad.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodServiceAccount, keypad.AuthMethod)
	assert.Equal(t, []string{"Office front door"}, keypad.Roles)
	require.NotNil(t, keypad.Active)
	assert.True(t, *keypad.Active)
	require.NotNil(t, keypad.LastLogin)
	assert.Equal(t, "2025-07-07T12:00:00Z", keypad.LastLogin.UTC().Format("2006-01-02T15:04:05Z"))
	assert.Empty(t, keypad.Email)
}

func nukiMockTransport(accountUsers roundTripFunc) http.RoundTripper {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(req.URL.Path, "/smartlock") && req.URL.Path != "" && !strings.Contains(req.URL.Path, "/auth"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`[]`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case strings.Contains(req.URL.Path, "/smartlock/auth/paged"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"results":[],"pagination":{"totalPages":0,"currentPage":0}}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case strings.Contains(req.URL.Path, "/account/user"):
			return accountUsers(req)
		default:
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(``)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}
	})
}

func TestNukiDriverPaginatesPastShortPage(t *testing.T) {
	t.Parallel()

	var offsets []string

	client := &http.Client{Transport: nukiMockTransport(func(req *http.Request) (*http.Response, error) {
		offset := req.URL.Query().Get("offset")
		offsets = append(offsets, offset)

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

func TestNukiDriverPaginationLimit(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: nukiMockTransport(func(req *http.Request) (*http.Response, error) {
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

			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if strings.HasSuffix(req.URL.Path, "/smartlock") && !strings.Contains(req.URL.Path, "/auth") {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`[]`)),
						Header:     http.Header{"Content-Type": []string{"application/json"}},
					}, nil
				}

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
			name:      "row without any identity is rejected",
			user:      nukiAccountUser{Name: "Ada"},
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			record, err := nukiAccountRecord(tc.user, nil, nil)
			if tc.wantError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantExternalID, record.ExternalID)
			assert.Equal(t, tc.wantFullName, record.FullName)
			assert.Equal(t, tc.wantType, record.AccountType)
			assert.Nil(t, record.Active)
			assert.Empty(t, record.Roles)
		})
	}
}

func TestNukiRolesAndActive(t *testing.T) {
	t.Parallel()

	enabled := true
	disabled := false
	remote := true
	lockNames := map[int64]string{10: "Front door"}

	auths := []nukiSmartlockAuth{
		{
			SmartlockID:      10,
			Enabled:          &enabled,
			RemoteAllowed:    &remote,
			LastActiveDate:   "2025-06-01T12:00:00Z",
			AllowedUntilDate: "2099-01-01T00:00:00Z",
		},
		{
			SmartlockID:    99,
			Enabled:        &disabled,
			LastActiveDate: "2025-01-01T00:00:00Z",
		},
	}

	assert.Equal(t, []string{"Front door", "Remote access", "Smartlock 99"}, nukiRoles(auths, lockNames))

	active := nukiActiveFromAuths(auths, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	require.NotNil(t, active)
	assert.True(t, *active)

	expiredOnly := []nukiSmartlockAuth{{
		SmartlockID:      10,
		Enabled:          &enabled,
		AllowedUntilDate: "2020-01-01T00:00:00Z",
	}}
	inactive := nukiActiveFromAuths(expiredOnly, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	require.NotNil(t, inactive)
	assert.False(t, *inactive)

	last := nukiLastActive(auths)
	require.NotNil(t, last)
	assert.Equal(t, "2025-06-01T12:00:00Z", last.UTC().Format("2006-01-02T15:04:05Z"))
}

func TestNukiServiceAccountRecord(t *testing.T) {
	t.Parallel()

	enabled := true
	auth := nukiSmartlockAuth{
		ID:             "auth-1",
		SmartlockID:    10,
		AuthID:         7,
		Type:           nukiAuthTypeKeypadCode,
		Name:           "Guest code",
		Enabled:        &enabled,
		CreationDate:   "2025-05-01T00:00:00Z",
		LastActiveDate: "2025-06-01T00:00:00Z",
	}

	record, err := nukiServiceAccountRecord(auth, map[int64]string{10: "Front door"})
	require.NoError(t, err)
	assert.Equal(t, "auth-1", record.ExternalID)
	assert.Equal(t, "Guest code", record.FullName)
	assert.Equal(t, []string{"Front door"}, record.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, record.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodServiceAccount, record.AuthMethod)
	require.NotNil(t, record.Active)
	assert.True(t, *record.Active)

	_, err = nukiServiceAccountRecord(nukiSmartlockAuth{SmartlockID: 1}, nil)
	require.Error(t, err)
}
