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
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	elevenLabsWorkspaceSegment = "workspace"
	elevenLabsMembersSegment   = "members"

	elevenLabsSeatAdmin      = "workspace_admin"
	elevenLabsSeatMember     = "workspace_member"
	elevenLabsSeatLiteMember = "workspace_lite_member"
)

// ElevenLabsDriver lists the members of the one ElevenLabs workspace its API
// key belongs to. The key carries the workspace, so there is no tenant
// selector and nothing to pick.
type ElevenLabsDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*ElevenLabsDriver)(nil)

// elevenLabsMember is one entry of GET /v1/workspace/members, which answers
// with a bare JSON array rather than an envelope. The endpoint takes no
// pagination parameters and returns the whole workspace in one response,
// locked members included.
type elevenLabsMember struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	SeatType  string `json:"seat_type"`
	IsOwner   bool   `json:"is_owner"`
	IsLocked  bool   `json:"is_locked"`
}

// ElevenLabsMembersURL builds the workspace-members URL from an API origin. It
// is exported because the connection probe checks this same endpoint from
// another package: were it to re-derive the path, moving the driver's would
// leave the probe reporting a healthy connection against the old one.
func ElevenLabsMembersURL(baseURL string) (string, error) {
	endpoint, err := url.JoinPath(baseURL, elevenLabsWorkspaceSegment, elevenLabsMembersSegment)
	if err != nil {
		return "", fmt.Errorf("cannot build elevenlabs members URL: %w", err)
	}

	return endpoint, nil
}

func NewElevenLabsDriver(httpClient *http.Client, baseURL string) *ElevenLabsDriver {
	return &ElevenLabsDriver{httpClient: httpClient, baseURL: baseURL}
}

func (d *ElevenLabsDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	members, err := d.fetchMembers(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(members))

	for _, m := range members {
		email := strings.TrimSpace(m.Email)
		if email == "" {
			continue
		}

		records = append(records, elevenLabsRecord(m, email))
	}

	return records, nil
}

func (d *ElevenLabsDriver) fetchMembers(ctx context.Context) ([]elevenLabsMember, error) {
	endpoint, err := ElevenLabsMembersURL(d.baseURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create elevenlabs members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute elevenlabs members request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// ElevenLabs answers a rejected key with 400 and an authentication_error
	// body rather than 401, so a non-2xx here carries no more meaning than
	// "the call failed"; the connection probe is what tells the customer
	// their key is the problem.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch elevenlabs members: unexpected status %d", resp.StatusCode)
	}

	var members []elevenLabsMember
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("cannot decode elevenlabs members response: %w", err)
	}

	return members, nil
}

// elevenLabsRecord maps one member onto the review record. It is a function of
// its own so the locked-seat rule below can be asserted without a cassette.
func elevenLabsRecord(m elevenLabsMember, email string) AccountRecord {
	// is_locked is the workspace's own deactivation: a locked member keeps
	// their seat and their history but cannot use the workspace, so they are
	// reported inactive rather than dropped from the review.
	active := !m.IsLocked

	return AccountRecord{
		Email:       email,
		FullName:    elevenLabsFullName(m, email),
		Roles:       elevenLabsRoles(m),
		IsAdmin:     new(elevenLabsIsAdmin(m)),
		Active:      &active,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  strings.TrimSpace(m.UserID),
	}
}

func elevenLabsFullName(m elevenLabsMember, fallback string) string {
	// ElevenLabs stores only a first name, so FullName is that name when the
	// member has set one and the email otherwise.
	if name := strings.TrimSpace(m.FirstName); name != "" {
		return name
	}

	return fallback
}

// elevenLabsIsAdmin reports the two ways a member administers the workspace:
// the owner, and anyone holding the workspace_admin seat. Both are matched
// exactly — a future seat type reports no admin signal rather than a false
// one.
func elevenLabsIsAdmin(m elevenLabsMember) bool {
	return m.IsOwner || m.SeatType == elevenLabsSeatAdmin
}

// elevenLabsRoles maps the seat type to a display label and adds Owner
// alongside it, since ownership is a separate privilege a member holds on top
// of their seat. An unrecognised seat type is preserved verbatim.
func elevenLabsRoles(m elevenLabsMember) []string {
	roles := []string{}

	if m.IsOwner {
		roles = append(roles, "Owner")
	}

	switch m.SeatType {
	case elevenLabsSeatAdmin:
		roles = append(roles, "Workspace Admin")
	case elevenLabsSeatMember:
		roles = append(roles, "Workspace Member")
	case elevenLabsSeatLiteMember:
		roles = append(roles, "Workspace Lite Member")
	default:
		if seat := strings.TrimSpace(m.SeatType); seat != "" {
			roles = append(roles, seat)
		}
	}

	return roles
}
