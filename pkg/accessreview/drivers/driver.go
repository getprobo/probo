// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package drivers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// AccountRecord represents a single account from an access source or identity
// source. All fields are best-effort; sources populate what they can. Drivers
// must return ALL accounts the source exposes (including inactive / suspended
// / deleted); classification is the job of the reviewer or of an agent run
// against the campaign, not of the fetch pipeline.
//
// Active is three-valued: nil means the source API has no explicit
// account-status signal for this account (the driver cannot tell), a non-nil
// pointer means the driver observed an explicit signal (true = active at
// source, false = deactivated / suspended / deleted). Drivers whose API does
// not distinguish active from deactivated accounts must leave Active nil
// rather than fabricate a value.
type AccountRecord struct {
	Email       string
	FullName    string
	Roles       []string // system roles/permissions (e.g. "Admin", "Viewer")
	JobTitle    string   // HR job title / department (e.g. "Software Engineer")
	Active      *bool
	IsAdmin     bool
	MFAStatus   coredata.MFAStatus
	AuthMethod  coredata.AccessReviewEntryAuthMethod
	AccountType coredata.AccessReviewEntryAccountType
	LastLogin   *time.Time
	CreatedAt   *time.Time
	ExternalID  string // system-specific user ID
}

// maxPaginationPages is the upper bound on the number of pages a driver will
// fetch from an external API. This prevents infinite loops if an API returns
// a non-empty cursor on every response.
const maxPaginationPages = 500

// ErrPaginationLimitReached is returned when a driver exhausts the maximum
// number of pagination pages without reaching the end of the result set.
var ErrPaginationLimitReached = fmt.Errorf("pagination limit of %d pages reached", maxPaginationPages)

// Driver defines the interface for fetching accounts from an access or
// identity source. Each driver implementation corresponds to a specific
// system (e.g. Google Workspace, AWS IAM, Probo memberships, CSV).
//
// All sources in a campaign's scope return "who actually has access" data.
type Driver interface {
	// ListAccounts returns all accounts from the source system.
	ListAccounts(ctx context.Context) ([]AccountRecord, error)
}

// parseRFC3339Ptr parses an RFC 3339 timestamp into a *time.Time, returning
// nil for an empty or unparseable value. Drivers use it for best-effort
// timestamp fields (created_at, last_login_at) that an API may omit.
func parseRFC3339Ptr(s string) *time.Time {
	if s == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}

	return &t
}

// activeFromStatus maps a provider status string to the three-valued Active
// signal for providers whose only "live" state is the literal "active" and
// whose remaining status enum is not otherwise enumerated: "active" → active,
// an empty status → nil (no signal), and any other non-empty status →
// inactive. Used by drivers like Pylon and Brevo; a provider with a fully
// known status enum (e.g. Render's active/inactive) maps its own values
// explicitly instead, so an unrecognised value stays nil rather than false.
func activeFromStatus(status string) *bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active":
		active := true

		return &active
	case "":
		return nil
	default:
		inactive := false

		return &inactive
	}
}

// ownerMemberRoles maps a provider role to a display label for providers whose
// role model is exactly owner/member (e.g. Crisp operators, Scaleway org user
// types): "owner" → Owner, "member" → Member. An unknown future value is passed
// through verbatim and no role yields an empty slice.
func ownerMemberRoles(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner":
		return []string{"Owner"}
	case "member":
		return []string{"Member"}
	default:
		if r := strings.TrimSpace(role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}

// isOwnerRole reports whether a provider role is the owner, the only role in the
// owner/member model that grants administrative access.
func isOwnerRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "owner")
}
