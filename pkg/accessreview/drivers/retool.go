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
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	retoolUsersSegment = "users"

	// retoolUsersPageSize is the page size requested from GET /users. The
	// endpoint returns every user when `limit` is omitted, but an explicit
	// page size keeps one response bounded on large organizations.
	retoolUsersPageSize = 100

	retoolUserTypeDefault = "default"
	retoolUserTypeMobile  = "mobile"
	retoolUserTypeEmbed   = "embed"

	retoolSeatBuilder      = "builder"
	retoolSeatInternalUser = "internalUser"
	retoolSeatExternalUser = "externalUser"
)

// RetoolDriver lists the members of the one Retool organization its API token
// belongs to. The token carries the organization — on Retool Cloud it routes
// through the shared api.retool.com gateway, on a self-hosted instance the
// base URL names the instance — so there is no tenant selector.
type RetoolDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*RetoolDriver)(nil)

type retoolUser struct {
	ID                   string  `json:"id"`
	Email                string  `json:"email"`
	Active               bool    `json:"active"`
	CreatedAt            string  `json:"created_at"`
	LastActive           *string `json:"last_active"`
	FirstName            *string `json:"first_name"`
	LastName             *string `json:"last_name"`
	IsAdmin              bool    `json:"is_admin"`
	UserType             string  `json:"user_type"`
	SeatType             *string `json:"seat_type"`
	TwoFactorAuthEnabled bool    `json:"two_factor_auth_enabled"`
}

type retoolUsersResponse struct {
	Data      []retoolUser `json:"data"`
	NextToken *string      `json:"next_token"`
	HasMore   bool         `json:"has_more"`
}

// RetoolUsersURL builds the users URL from an API base. It is exported because
// the connection probe checks this same endpoint from another package.
func RetoolUsersURL(baseURL string) (string, error) {
	endpoint, err := url.JoinPath(baseURL, retoolUsersSegment)
	if err != nil {
		return "", fmt.Errorf("cannot build retool users URL: %w", err)
	}

	return endpoint, nil
}

func NewRetoolDriver(httpClient *http.Client, baseURL string) *RetoolDriver {
	return &RetoolDriver{httpClient: httpClient, baseURL: baseURL}
}

func (d *RetoolDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records   []AccountRecord
		nextToken string
	)

	for range maxPaginationPages {
		resp, err := d.fetchUsersPage(ctx, nextToken)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			record := AccountRecord{
				Email:       email,
				FullName:    retoolFullName(u, email),
				Roles:       retoolRoles(u),
				IsAdmin:     new(u.IsAdmin),
				Active:      new(u.Active),
				MFAStatus:   retoolMFAStatus(u),
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(u.CreatedAt),
				ExternalID:  strings.TrimSpace(u.ID),
			}

			if u.LastActive != nil {
				record.LastLogin = parseRFC3339Ptr(*u.LastActive)
			}

			records = append(records, record)
		}

		if !resp.HasMore {
			return records, nil
		}

		// Retool says there is more but will not say where to resume. Returning
		// what we have would be a short roster with no error, and a member
		// missing from a campaign is reviewed by nobody. Erroring also stops
		// the loop replaying page one forever.
		if resp.NextToken == nil || *resp.NextToken == "" {
			return nil, fmt.Errorf("cannot list all retool users: next page has no token")
		}

		nextToken = *resp.NextToken
	}

	return nil, fmt.Errorf("cannot list all retool users: %w", ErrPaginationLimitReached)
}

func (d *RetoolDriver) fetchUsersPage(ctx context.Context, nextToken string) (*retoolUsersResponse, error) {
	endpoint, err := RetoolUsersURL(d.baseURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create retool users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(retoolUsersPageSize))

	if nextToken != "" {
		q.Set("next_token", nextToken)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute retool users request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch retool users: unexpected status %d", resp.StatusCode)
	}

	var body retoolUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot decode retool users response: %w", err)
	}

	return &body, nil
}

func retoolFullName(u retoolUser, fallback string) string {
	var parts []string

	if u.FirstName != nil {
		if first := strings.TrimSpace(*u.FirstName); first != "" {
			parts = append(parts, first)
		}
	}

	if u.LastName != nil {
		if last := strings.TrimSpace(*u.LastName); last != "" {
			parts = append(parts, last)
		}
	}

	if len(parts) > 0 {
		return strings.Join(parts, " ")
	}

	return fallback
}

// retoolMFAStatus reads Retool's own per-user two-factor flag. Retool reports
// it for every user, so the status is never unknown.
func retoolMFAStatus(u retoolUser) coredata.MFAStatus {
	if u.TwoFactorAuthEnabled {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

// retoolRoles combines the three independent things Retool records about a
// user's standing: whether they administer the organization, which seat they
// hold (what they may build), and which kind of account they are. Each is
// appended separately rather than folded into one label, since a reviewer
// acts on them differently. Group membership is a fourth axis and lives
// behind a separate endpoint the roster does not carry.
func retoolRoles(u retoolUser) []string {
	roles := []string{}

	if u.IsAdmin {
		roles = append(roles, "Admin")
	}

	if u.SeatType != nil {
		switch *u.SeatType {
		case retoolSeatBuilder:
			roles = append(roles, "Builder")
		case retoolSeatInternalUser:
			roles = append(roles, "Internal User")
		case retoolSeatExternalUser:
			roles = append(roles, "External User")
		default:
			if seat := strings.TrimSpace(*u.SeatType); seat != "" {
				roles = append(roles, seat)
			}
		}
	}

	switch u.UserType {
	case retoolUserTypeMobile:
		roles = append(roles, "Mobile")
	case retoolUserTypeEmbed:
		roles = append(roles, "Embed")
	case retoolUserTypeDefault:
		// The ordinary account kind carries no privilege of its own.
	default:
		if kind := strings.TrimSpace(u.UserType); kind != "" {
			roles = append(roles, kind)
		}
	}

	return roles
}
