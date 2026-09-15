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
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// crispCassetteWebsiteID is the website the sanitized cassette is keyed on.
// Recording talks to the real workspace named by CRISP_WEBSITE_ID and
// sanitizeCrispOperators rewrites it to this value in both the URL and the
// body, so replay matches without the real identifier reaching the repo.
const crispCassetteWebsiteID = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

func TestCrispDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/crisp", "CRISP_API_KEY", sanitizeCrispOperators)
	// Crisp authenticates with HTTP Basic over the "identifier:key" plugin
	// token. The matcher ignores Authorization, so replay needs no credential.
	client := newVCRClient(rec, basicAuthUserPass(os.Getenv("CRISP_API_KEY")))

	websiteID := os.Getenv("CRISP_WEBSITE_ID")
	if websiteID == "" {
		websiteID = crispCassetteWebsiteID
	}

	driver := NewCrispDriver(client, websiteID, "https://api.crisp.chat/v1")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// The cassette holds an operator and a sandbox member; only the operator is
	// a human with access, and dropping the other is the point of this fixture.
	// No "invite" member is covered: the recorded workspace had no pending
	// invitation and inventing one is what made the previous cassette useless.
	require.Len(t, records, 1)

	owner := records[0]
	assert.Equal(t, "crisp-user-1", owner.ExternalID)
	assert.Equal(t, "operator1@example.com", owner.Email)
	assert.Equal(t, "Operator1 Number1", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	// The recorded operator has has_token:false, which Crisp documents as
	// two-factor being off. An absent flag would map to Unknown instead.
	assert.Equal(t, coredata.MFAStatusDisabled, owner.MFAStatus)
	// operators/list carries no account-status signal, so Active stays nil.
	assert.Nil(t, owner.Active)

	// Regression: "sandbox" is the Marketplace plugin developer (Probo itself),
	// not a workspace member. It repeats the operator's email under its own
	// user_id, so before the type filter it landed as a second account for the
	// same human in every Crisp review.
	for _, r := range records {
		assert.NotEqual(t, "crisp-user-2", r.ExternalID, "sandbox member must not be listed")
	}
}

// sanitizeCrispOperators rewrites the recorded operators/list response so the
// committed cassette carries no real identity, while preserving the member
// "type" of every entry: that field is exactly what the driver now filters on,
// so a sanitizer that dropped or normalised it would erase the regression this
// fixture exists to catch.
func sanitizeCrispOperators(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize crisp response with status %d", i.Response.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded crisp response: %w", err)
	}

	raw, ok := body["data"]
	if !ok {
		return fmt.Errorf("recorded crisp response has no data field")
	}

	var members []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return fmt.Errorf("cannot decode recorded crisp members: %w", err)
	}

	if len(members) == 0 {
		return fmt.Errorf("recorded crisp response lists no members")
	}

	// Every identity value seen, so the sanitized body can be checked for
	// leftovers below.
	var recorded []string

	for idx, member := range members {
		// Decoded rather than merely present: a type that came back as null or
		// absent would silently disable the filter this cassette pins down.
		rawType, ok := member["type"]
		if !ok {
			return fmt.Errorf("recorded crisp member %d has no type field", idx)
		}

		var memberType string
		if err := json.Unmarshal(rawType, &memberType); err != nil {
			return fmt.Errorf("recorded crisp member %d has a non-string type: %w", idx, err)
		}

		if strings.TrimSpace(memberType) == "" {
			return fmt.Errorf("recorded crisp member %d has an empty type", idx)
		}

		rawDetails, ok := member["details"]
		if !ok {
			return fmt.Errorf("recorded crisp member %d has no details field", idx)
		}

		var details map[string]json.RawMessage
		if err := json.Unmarshal(rawDetails, &details); err != nil {
			return fmt.Errorf("cannot decode recorded crisp member %d details: %w", idx, err)
		}

		replacements := map[string]string{
			"user_id":    fmt.Sprintf("crisp-user-%d", idx+1),
			"email":      fmt.Sprintf("operator%d@example.com", idx+1),
			"first_name": fmt.Sprintf("Operator%d", idx+1),
			"last_name":  fmt.Sprintf("Number%d", idx+1),
		}

		for field, replacement := range replacements {
			// Absent fields are left absent: "user_id" is documented as
			// operator/sandbox only and "role" as operator/invite only, so
			// materialising them would record a shape Crisp never sends.
			value, present, err := crispFieldCarriesIdentity(details, field, idx)
			if err != nil {
				return err
			}

			if !present {
				continue
			}

			recorded = append(recorded, value)

			encoded, err := json.Marshal(replacement)
			if err != nil {
				return fmt.Errorf("cannot encode sanitized %s: %w", field, err)
			}

			details[field] = encoded
		}

		encodedDetails, err := json.Marshal(details)
		if err != nil {
			return fmt.Errorf("cannot re-encode crisp member %d details: %w", idx, err)
		}

		member["details"] = encodedDetails
	}

	sanitizedMembers, err := json.Marshal(members)
	if err != nil {
		return fmt.Errorf("cannot re-encode crisp members: %w", err)
	}

	body["data"] = sanitizedMembers

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode crisp response: %w", err)
	}

	// Aborting here writes no cassette, which is the right way to fail: a
	// spurious abort costs a re-record, a miss commits a real address.
	for _, value := range recorded {
		if strings.Contains(string(sanitized), value) {
			return fmt.Errorf("sanitized crisp response still contains a recorded identity value")
		}
	}

	i.Response.Body = string(sanitized)
	i.Response.ContentLength = int64(len(sanitized))
	i.Response.Headers.Set("Content-Length", fmt.Sprintf("%d", len(sanitized)))

	// The real website id keys the request URL, so replay would miss it unless
	// it is rewritten to match the id the test uses when the env var is unset.
	// Crisp also echoes the id back in X-Crisp-Ray, so the headers are swept
	// too rather than just the URL.
	if websiteID := strings.TrimSpace(os.Getenv("CRISP_WEBSITE_ID")); websiteID != "" {
		i.Request.URL = strings.ReplaceAll(i.Request.URL, websiteID, crispCassetteWebsiteID)

		for _, headers := range []http.Header{i.Request.Headers, i.Response.Headers} {
			for name, values := range headers {
				for idx, value := range values {
					values[idx] = strings.ReplaceAll(value, websiteID, crispCassetteWebsiteID)
				}

				headers[name] = values
			}
		}
	}

	return nil
}

// crispFieldCarriesIdentity reports whether details holds field as a non-empty
// string. Crisp omits most detail fields for non-operator members, so absence
// is normal and only a present-but-wrong-shape value is an error.
func crispFieldCarriesIdentity(
	details map[string]json.RawMessage,
	field string,
	idx int,
) (string, bool, error) {
	raw, ok := details[field]
	if !ok {
		return "", false, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		// null is legitimate (Crisp sends it for avatar/title); anything else
		// present but non-string is a shape change worth failing on.
		if string(raw) == "null" {
			return "", false, nil
		}

		return "", false, fmt.Errorf("recorded crisp member %d has a non-string %s: %w", idx, field, err)
	}

	if strings.TrimSpace(value) == "" {
		return "", false, nil
	}

	return value, true, nil
}
