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
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const brevoInvitedUsersEndpoint = "https://api.brevo.com/v3/organization/invited/users"

// BrevoDriver lists the invited users (organization seats) of a single Brevo
// account. The API key (sent in the api-key header by the connection
// transport) is bound to one account, so GET /v3/organization/invited/users
// returns every invited user of that account with no tenant selector and no
// pagination.
type BrevoDriver struct {
	httpClient *http.Client
}

var _ Driver = (*BrevoDriver)(nil)

type brevoInvitedUser struct {
	// ID is Brevo's stable user identifier (a Mongo-style ObjectID). The
	// documented schema omits it, but the live API returns it, so it is
	// preferred over the email as the ExternalID.
	ID    string `json:"id"`
	Email string `json:"email"`
	// IsOwner flags the account owner. The live API returns a JSON boolean
	// while older docs/SDK show a string ("true"/"false"); decoded as
	// RawMessage and read via brevoIsOwner to tolerate both shapes.
	IsOwner json.RawMessage `json:"is_owner"`
	// Status is the invitation state: "active" or "pending".
	Status string `json:"status"`
	// FeatureAccess maps a feature area (marketing / crm / conversations /
	// transactional / phone / …) to the user's access level on it. Values are
	// strings (e.g. "owner", "full", "none"); decoded as RawMessage so a
	// non-string shape degrades gracefully instead of failing the whole
	// decode.
	FeatureAccess map[string]json.RawMessage `json:"feature_access"`
}

type brevoInvitedUsersResponse struct {
	Users []brevoInvitedUser `json:"users"`
}

func NewBrevoDriver(httpClient *http.Client) *BrevoDriver {
	return &BrevoDriver{httpClient: httpClient}
}

func (d *BrevoDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, brevoInvitedUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create brevo invited users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute brevo invited users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch brevo invited users: unexpected status %d", httpResp.StatusCode)
	}

	var resp brevoInvitedUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode brevo invited users response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Users))

	for _, u := range resp.Users {
		email := strings.TrimSpace(u.Email)
		if email == "" {
			continue
		}

		records = append(records, AccountRecord{
			Email: email,
			// Brevo's invited-users API exposes no display name.
			FullName:    email,
			Roles:       brevoRoles(u.FeatureAccess),
			Active:      activeFromStatus(u.Status),
			IsAdmin:     brevoIsOwner(u.IsOwner),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  brevoExternalID(u, email),
		})
	}

	return records, nil
}

// brevoExternalID prefers Brevo's stable user id, falling back to the email
// (the only other durable identifier) when an account has none.
func brevoExternalID(u brevoInvitedUser, email string) string {
	if id := strings.TrimSpace(u.ID); id != "" {
		return id
	}

	return email
}

// brevoIsOwner reads the account-owner flag, tolerating both the JSON boolean
// the live API returns and the string ("true"/"false") shown in older
// docs/SDK.
func brevoIsOwner(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}

	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.EqualFold(strings.TrimSpace(s), "true")
	}

	return false
}

// brevoRoles summarises an invited user's per-feature access levels
// (marketing / crm / conversations) into a de-duplicated, sorted set of role
// labels. Each feature_access value is normally a string such as "owner";
// the "none" level and any non-string shape are skipped so the result holds
// only the access levels actually granted.
func brevoRoles(featureAccess map[string]json.RawMessage) []string {
	seen := make(map[string]struct{})

	for _, raw := range featureAccess {
		var level string
		if err := json.Unmarshal(raw, &level); err != nil {
			continue
		}

		level = strings.TrimSpace(level)
		if level == "" || strings.EqualFold(level, "none") {
			continue
		}

		seen[level] = struct{}{}
	}

	roles := make([]string, 0, len(seen))
	for level := range seen {
		roles = append(roles, level)
	}

	sort.Strings(roles)

	return roles
}
