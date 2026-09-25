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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// sanitizeLangfuseMemberships replaces member identity in a recorded
// response with synthetic values.
//
// It edits the decoded JSON in place rather than re-marshalling
// langfuseMembershipsResponse, which would rewrite the body as a
// serialization of the type under test and drop whatever the driver does not
// model. An unrecognised response errors, which aborts the save.
//
// The hook runs on save, so a recording run still sees live data and its
// assertions fail. Re-record, then run again without LANGFUSE_API_KEY.
// langfuseFieldCarriesIdentity reports whether an optional membership field
// holds a value that has to be replaced, and rejects one that is not a string.
//
// Absent or blank is left exactly as recorded, because the driver acts on it:
// it skips a member with no email, and falls back to the email when the name
// is blank. Substituting a value would change what the cassette replays.
func langfuseFieldCarriesIdentity(
	membership map[string]json.RawMessage,
	field string,
	idx int,
) (string, bool, error) {
	raw, ok := membership[field]
	if !ok {
		return "", false, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false, fmt.Errorf("recorded langfuse membership %d has a non-string %s: %w", idx, field, err)
	}

	return value, strings.TrimSpace(value) != "", nil
}

func sanitizeLangfuseMemberships(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize langfuse response with status %d", i.Response.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded langfuse response: %w", err)
	}

	raw, ok := body["memberships"]
	if !ok {
		return fmt.Errorf("recorded langfuse response has no memberships field")
	}

	var memberships []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &memberships); err != nil {
		return fmt.Errorf("cannot decode recorded langfuse memberships: %w", err)
	}

	if len(memberships) == 0 {
		return fmt.Errorf("recorded langfuse response lists no members")
	}

	// Every identity value seen, so the sanitized body can be checked for
	// leftovers below.
	var recorded []string

	for idx, membership := range memberships {
		// Decoded rather than merely present: replacing a userId that came
		// back as null or a number would paper over the struct mismatch this
		// cassette exists to expose.
		raw, ok := membership["userId"]
		if !ok {
			return fmt.Errorf("recorded langfuse membership %d has no userId field", idx)
		}

		var userID string
		if err := json.Unmarshal(raw, &userID); err != nil {
			return fmt.Errorf("recorded langfuse membership %d has a non-string userId: %w", idx, err)
		}

		// Unmarshalling JSON null into a string is a no-op in Go rather than
		// an error, so emptiness is the check that catches both null and "".
		if strings.TrimSpace(userID) == "" {
			return fmt.Errorf("recorded langfuse membership %d has an empty userId", idx)
		}

		recorded = append(recorded, userID)

		replacements := map[string]string{
			"userId": fmt.Sprintf("lf-user-%d", idx+1),
		}

		for field, replacement := range map[string]string{
			"email": fmt.Sprintf("member%d@example.com", idx+1),
			"name":  fmt.Sprintf("Member %d", idx+1),
		} {
			value, ok, err := langfuseFieldCarriesIdentity(membership, field, idx)
			if err != nil {
				return err
			}

			if ok {
				recorded = append(recorded, value)
				replacements[field] = replacement
			}
		}

		for field, value := range replacements {
			encoded, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("cannot encode sanitized %s: %w", field, err)
			}

			membership[field] = encoded
		}
	}

	sanitizedMemberships, err := json.Marshal(memberships)
	if err != nil {
		return fmt.Errorf("cannot re-encode langfuse memberships: %w", err)
	}

	body["memberships"] = sanitizedMemberships

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode langfuse response: %w", err)
	}

	// Unknown fields are preserved verbatim to keep the recorded shape, so
	// identity Langfuse returns outside the three known keys (a top-level
	// user object, an invited-by record) would otherwise survive. Aborting
	// here writes no cassette, which is the right way to fail: a short name
	// may match by coincidence, and a spurious abort costs a re-record while
	// a miss would commit a real address.
	for _, value := range recorded {
		if strings.Contains(string(sanitized), value) {
			return fmt.Errorf("sanitized langfuse response still contains a recorded identity value")
		}
	}

	i.Response.Body = string(sanitized)
	i.Response.ContentLength = int64(len(sanitized))
	i.Response.Headers.Set("Content-Length", fmt.Sprintf("%d", len(sanitized)))

	return nil
}

func TestLangfuseDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/langfuse", "LANGFUSE_API_KEY", sanitizeLangfuseMemberships)
	// Langfuse authenticates with HTTP Basic (publicKey:secretKey). The
	// matcher ignores Authorization, so replay needs no auth.
	client := newVCRClient(rec, basicAuthUserPass(os.Getenv("LANGFUSE_API_KEY")))

	driver := NewLangfuseDriver(client, "https://cloud.langfuse.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// The recorded organization has a single owner. Role mapping is covered
	// by TestLangfuseRoles and the loop by TestLangfuseListAccountsMapping.
	require.Len(t, records, 1)

	owner := records[0]
	assert.Equal(t, "lf-user-1", owner.ExternalID)
	assert.Equal(t, "member1@example.com", owner.Email)
	assert.Equal(t, "Member 1", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, owner.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
}

// TestLangfuseListAccountsMapping covers what a single-member cassette
// cannot. It stubs the transport rather than padding the cassette, which
// would make the recording a fixture again.
func TestSanitizeLangfuseMemberships(t *testing.T) {
	t.Parallel()

	interaction := &cassette.Interaction{}
	interaction.Response.Code = http.StatusOK
	interaction.Response.Headers = http.Header{}
	interaction.Response.Body = `{"memberships":[` +
		`{"userId":"real-id","role":"OWNER","email":"real@corp.example","name":"Real Person","joinedAt":"2026-01-01"}` +
		`],"nextPage":null}`

	require.NoError(t, sanitizeLangfuseMemberships(interaction))

	assert.NotContains(t, interaction.Response.Body, "real@corp.example")
	assert.NotContains(t, interaction.Response.Body, "Real Person")
	assert.NotContains(t, interaction.Response.Body, "real-id")
	assert.Contains(t, interaction.Response.Body, "member1@example.com")

	// Identity outside the three known keys is exactly what the verbatim
	// preservation of unknown fields would otherwise leak, so the save aborts.
	leaky := &cassette.Interaction{}
	leaky.Response.Code = http.StatusOK
	leaky.Response.Headers = http.Header{}
	leaky.Response.Body = `{"memberships":[{"userId":"u-1","role":"OWNER","email":"real@corp.example","name":"Real Person",` +
		`"invitedBy":{"email":"real@corp.example"}}]}`

	require.Error(t, sanitizeLangfuseMemberships(leaky))

	// A blank name is what the driver falls back on, so like a blank email it
	// must survive intact rather than gaining a synthetic value.
	blankName := &cassette.Interaction{}
	blankName.Response.Code = http.StatusOK
	blankName.Response.Headers = http.Header{}
	blankName.Response.Body = `{"memberships":[{"userId":"real-9","role":"MEMBER","email":"real@corp.example","name":""}]}`

	require.NoError(t, sanitizeLangfuseMemberships(blankName))
	assert.Contains(t, blankName.Response.Body, `"name":""`)
	assert.NotContains(t, blankName.Response.Body, "real@corp.example")
	assert.NotContains(t, blankName.Response.Body, "real-9")

	// A blank email is what the driver skips on, so it must survive intact:
	// substituting one would keep a member the live call dropped.
	blank := &cassette.Interaction{}
	blank.Response.Code = http.StatusOK
	blank.Response.Headers = http.Header{}
	blank.Response.Body = `{"memberships":[{"userId":"real-1","role":"MEMBER","email":"","name":"Real Name"}]}`

	require.NoError(t, sanitizeLangfuseMemberships(blank))
	assert.Contains(t, blank.Response.Body, `"email":""`)
	assert.NotContains(t, blank.Response.Body, "Real Name")
	assert.NotContains(t, blank.Response.Body, "real-1")

	// Fields the driver does not model must survive the rewrite.
	assert.Contains(t, interaction.Response.Body, "joinedAt")
	assert.Contains(t, interaction.Response.Body, "nextPage")
	assert.Contains(t, interaction.Response.Body, `"role":"OWNER"`)

	assert.Equal(t, int64(len(interaction.Response.Body)), interaction.Response.ContentLength)
}

func TestSanitizeLangfuseMembershipsRefusesUnexpectedResponses(t *testing.T) {
	t.Parallel()

	// A sanitizer error aborts the save, so an unrecognised response is
	// never committed as if it were a member list.
	for name, body := range map[string]string{
		"no memberships field":      `{"error":"forbidden"}`,
		"empty member list":         `{"memberships":[]}`,
		"membership missing userId": `{"memberships":[{"email":"e@example.com","name":"N"}]}`,
		"userId is null":            `{"memberships":[{"userId":null,"email":"e@example.com","name":"N"}]}`,
		"name is not a string":      `{"memberships":[{"userId":"u1","email":"e@example.com","name":{"first":"N"}}]}`,
		"email is not a string":     `{"memberships":[{"userId":"u1","email":42,"name":"N"}]}`,
		"userId is not a string":    `{"memberships":[{"userId":42,"email":"e@example.com","name":"N"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			interaction := &cassette.Interaction{}
			interaction.Response.Code = http.StatusOK
			interaction.Response.Headers = http.Header{}
			interaction.Response.Body = body

			require.Error(t, sanitizeLangfuseMemberships(interaction))
		})
	}

	interaction := &cassette.Interaction{}
	interaction.Response.Code = http.StatusForbidden
	interaction.Response.Headers = http.Header{}
	interaction.Response.Body = `{"memberships":[{"userId":"u1","email":"e@example.com","name":"N"}]}`

	require.Error(t, sanitizeLangfuseMemberships(interaction))
}

func TestLangfuseListAccountsMapping(t *testing.T) {
	t.Parallel()

	const body = `{"memberships":[` +
		`{"userId":"u1","role":"OWNER","email":"owner@example.com","name":"Olivia Owner"},` +
		`{"userId":"u2","role":"MEMBER","email":"member@example.com","name":""},` +
		`{"userId":"u3","role":"ADMIN","email":"","name":"No Email"},` +
		`{"userId":"u4","role":"SUPERVISOR","email":"future@example.com","name":"Future Role"}` +
		`]}`

	client := &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}

	records, err := NewLangfuseDriver(client, "https://cloud.langfuse.com").ListAccounts(context.Background())
	require.NoError(t, err)
	// Drops the member with no email.
	require.Len(t, records, 3)

	assert.Equal(t, "u1", records[0].ExternalID)
	assert.Equal(t, "Olivia Owner", records[0].FullName)
	assert.Equal(t, new(true), records[0].IsAdmin)

	// Empty name falls back to the email, per membership.
	assert.Equal(t, "u2", records[1].ExternalID)
	assert.Equal(t, "member@example.com", records[1].FullName)
	assert.Equal(t, []string{"Member"}, records[1].Roles)
	assert.Equal(t, new(false), records[1].IsAdmin)

	// An unmapped role survives verbatim and is not admin.
	assert.Equal(t, "u4", records[2].ExternalID)
	assert.Equal(t, []string{"SUPERVISOR"}, records[2].Roles)
	assert.Equal(t, new(false), records[2].IsAdmin)
}

func TestLangfuseFullName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ada Admin", langfuseFullName(langfuseMembership{Name: "Ada Admin"}, "ada@example.com"))
	// Langfuse leaves name empty for a member who never set one, and the
	// review table needs something to show, so the email stands in.
	assert.Equal(t, "ada@example.com", langfuseFullName(langfuseMembership{Name: ""}, "ada@example.com"))
	assert.Equal(t, "ada@example.com", langfuseFullName(langfuseMembership{Name: "   "}, "ada@example.com"))
}

func TestLangfuseRoles(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"Owner"}, langfuseRoles("OWNER"))
	assert.Equal(t, []string{"Admin"}, langfuseRoles("ADMIN"))
	assert.Equal(t, []string{"Member"}, langfuseRoles("MEMBER"))
	assert.Equal(t, []string{"Viewer"}, langfuseRoles("VIEWER"))
	assert.Equal(t, []string{"None"}, langfuseRoles("NONE"))
	assert.True(t, langfuseIsAdmin("OWNER"))
	assert.True(t, langfuseIsAdmin("ADMIN"))
	assert.False(t, langfuseIsAdmin("MEMBER"))
}
