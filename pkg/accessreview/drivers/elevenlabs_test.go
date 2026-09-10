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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// sanitizeElevenLabsMembers replaces member identity in a recorded response
// with synthetic values.
//
// It edits the decoded JSON in place rather than re-marshalling
// []elevenLabsMember, which would rewrite the body as a serialization of the
// type under test and drop whatever the driver does not model. An
// unrecognised response errors, which aborts the save.
//
// The hook runs on save, so a recording run still sees live data and its
// assertions fail. Re-record, then run again without ELEVENLABS_API_KEY.
func sanitizeElevenLabsMembers(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize elevenlabs response with status %d", i.Response.Code)
	}

	// The endpoint answers with a bare array, not an envelope.
	var members []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &members); err != nil {
		return fmt.Errorf("cannot decode recorded elevenlabs response: %w", err)
	}

	if len(members) == 0 {
		return fmt.Errorf("recorded elevenlabs response lists no members")
	}

	for idx, member := range members {
		// Decoded rather than merely present: replacing a user_id that came
		// back as null or a number would paper over the struct mismatch this
		// cassette exists to expose.
		for _, field := range []string{"user_id", "email"} {
			raw, ok := member[field]
			if !ok {
				return fmt.Errorf("recorded elevenlabs member %d has no %s field", idx, field)
			}

			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("recorded elevenlabs member %d has a non-string %s: %w", idx, field, err)
			}
		}

		// first_name is nullable, so it is checked separately: a member who
		// never set one records as null, and demanding a string there would
		// abort the save on the very workspace most worth re-recording against.
		raw, ok := member["first_name"]
		if !ok {
			return fmt.Errorf("recorded elevenlabs member %d has no first_name field", idx)
		}

		if !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			var name string
			if err := json.Unmarshal(raw, &name); err != nil {
				return fmt.Errorf("recorded elevenlabs member %d has a non-string first_name: %w", idx, err)
			}
		}

		member["user_id"] = json.RawMessage(fmt.Sprintf(`"user_%08d"`, idx+1))
		member["email"] = json.RawMessage(fmt.Sprintf(`"member%d@example.com"`, idx+1))
		member["first_name"] = json.RawMessage(fmt.Sprintf(`"Member%d"`, idx+1))
	}

	sanitized, err := json.Marshal(members)
	if err != nil {
		return fmt.Errorf("cannot re-encode sanitized elevenlabs response: %w", err)
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

func TestElevenLabsDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/elevenlabs", "ELEVENLABS_API_KEY", sanitizeElevenLabsMembers)
	// ElevenLabs authenticates via the xi-api-key header, not Authorization.
	client := newVCRClientWithHeader(rec, "xi-api-key", os.Getenv("ELEVENLABS_API_KEY"))

	driver := NewElevenLabsDriver(client, "https://api.elevenlabs.io/v1")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	// Cassette recorded live, then anonymized: one workspace owner holding
	// the workspace_admin seat.
	owner := records[0]
	assert.Equal(t, "user_00000001", owner.ExternalID)
	assert.Equal(t, "member1@example.com", owner.Email)
	assert.Equal(t, "Member1", owner.FullName)
	assert.Equal(t, []string{"Owner", "Workspace Admin"}, owner.Roles)
	assert.Equal(t, new(true), owner.IsAdmin)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
}

func TestElevenLabsRoles(t *testing.T) {
	t.Parallel()

	// Ownership is a privilege held on top of a seat, so it is its own entry.
	assert.Equal(
		t,
		[]string{"Owner", "Workspace Admin"},
		elevenLabsRoles(elevenLabsMember{IsOwner: true, SeatType: elevenLabsSeatAdmin}),
	)
	assert.Equal(
		t,
		[]string{"Workspace Member"},
		elevenLabsRoles(elevenLabsMember{SeatType: elevenLabsSeatMember}),
	)
	assert.Equal(
		t,
		[]string{"Workspace Lite Member"},
		elevenLabsRoles(elevenLabsMember{SeatType: elevenLabsSeatLiteMember}),
	)
	// An unrecognised future seat is preserved rather than dropped.
	assert.Equal(t, []string{"workspace_owner_plus"}, elevenLabsRoles(elevenLabsMember{SeatType: "workspace_owner_plus"}))
	// No seat at all is no role, never a nil slice.
	assert.Equal(t, []string{}, elevenLabsRoles(elevenLabsMember{}))
}

func TestElevenLabsIsAdmin(t *testing.T) {
	t.Parallel()

	assert.True(t, elevenLabsIsAdmin(elevenLabsMember{IsOwner: true, SeatType: elevenLabsSeatMember}))
	assert.True(t, elevenLabsIsAdmin(elevenLabsMember{SeatType: elevenLabsSeatAdmin}))
	assert.False(t, elevenLabsIsAdmin(elevenLabsMember{SeatType: elevenLabsSeatMember}))
	assert.False(t, elevenLabsIsAdmin(elevenLabsMember{SeatType: elevenLabsSeatLiteMember}))
	// A seat type that merely contains "admin" is not the admin seat.
	assert.False(t, elevenLabsIsAdmin(elevenLabsMember{SeatType: "workspace_admin_readonly"}))
}

func TestElevenLabsFullName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ada", elevenLabsFullName(elevenLabsMember{FirstName: " Ada "}, "ada@example.com"))
	// ElevenLabs stores no last name, and a blank first name falls back to
	// the email so a record is never nameless.
	assert.Equal(t, "ada@example.com", elevenLabsFullName(elevenLabsMember{FirstName: "  "}, "ada@example.com"))
	assert.Equal(t, "ada@example.com", elevenLabsFullName(elevenLabsMember{}, "ada@example.com"))
}

func TestElevenLabsRecord(t *testing.T) {
	t.Parallel()

	// A locked member keeps their seat and their history, so they stay in the
	// review marked inactive. Dropping them would hide exactly the access a
	// campaign exists to revoke.
	locked := elevenLabsRecord(elevenLabsMember{
		UserID:   "user_1",
		SeatType: elevenLabsSeatMember,
		IsLocked: true,
	}, "locked@example.com")

	require.NotNil(t, locked.Active)
	assert.False(t, *locked.Active)
	assert.Equal(t, "user_1", locked.ExternalID)
	assert.Equal(t, new(false), locked.IsAdmin)

	live := elevenLabsRecord(elevenLabsMember{
		UserID:   "user_2",
		SeatType: elevenLabsSeatAdmin,
	}, "admin@example.com")

	require.NotNil(t, live.Active)
	assert.True(t, *live.Active)
	assert.Equal(t, new(true), live.IsAdmin)
}
