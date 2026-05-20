// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"net/url"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// HerokuDriver fetches team members from the Heroku Platform API using
// a pre-authenticated HTTP client (Bearer token). Pagination is via
// Heroku's Range / Next-Range header pair (RFC 7233 style).
//
// Notes on data quality:
//   - The team-members endpoint does not expose suspension state, so
//     Active is left nil for v1.
//   - For federated teams the IdP is the source of truth for MFA, but
//     the API still reports `two_factor_authentication`. The driver
//     populates MFAStatus from that field and lets the access-review
//     UI surface federation context separately.
type HerokuDriver struct {
	httpClient *http.Client
	teamID     string
}

var _ Driver = (*HerokuDriver)(nil)

func NewHerokuDriver(httpClient *http.Client, teamID string) *HerokuDriver {
	return &HerokuDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		teamID: teamID,
	}
}

type herokuTeamMember struct {
	ID                      string `json:"id"`
	Email                   string `json:"email"`
	Role                    string `json:"role"`
	TwoFactorAuthentication bool   `json:"two_factor_authentication"`
	Federated               bool   `json:"federated"`
	CreatedAt               string `json:"created_at"`
	User                    struct {
		Email string `json:"email"`
		ID    string `json:"id"`
		Name  string `json:"name"`
	} `json:"user"`
}

func (d *HerokuDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	endpoint, err := url.JoinPath("https://api.heroku.com", "teams", url.PathEscape(d.teamID), "members")
	if err != nil {
		return nil, fmt.Errorf("cannot build heroku members URL: %w", err)
	}
	rangeHeader := ""

	for range maxPaginationPages {
		members, nextRange, err := d.queryMembers(ctx, endpoint, rangeHeader)
		if err != nil {
			return nil, err
		}

		for _, m := range members {
			email := m.Email
			if email == "" {
				email = m.User.Email
			}

			fullName := m.User.Name

			mfaStatus := coredata.MFAStatusDisabled
			if m.TwoFactorAuthentication {
				mfaStatus = coredata.MFAStatusEnabled
			}

			isAdmin := m.Role == "admin" || m.Role == "owner"

			externalID := m.User.ID
			if externalID == "" {
				externalID = m.ID
			}

			record := AccountRecord{
				Email:       email,
				FullName:    fullName,
				Role:        m.Role,
				IsAdmin:     isAdmin,
				MFAStatus:   mfaStatus,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
				ExternalID:  externalID,
			}

			if m.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
					record.CreatedAt = &t
				}
			}

			records = append(records, record)
		}

		if nextRange == "" {
			return records, nil
		}

		rangeHeader = nextRange
	}

	return nil, fmt.Errorf("cannot list all heroku accounts: %w", ErrPaginationLimitReached)
}

func (d *HerokuDriver) queryMembers(ctx context.Context, endpoint, rangeHeader string) ([]herokuTeamMember, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create heroku members request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute heroku members request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Heroku returns 206 Partial Content for ranged responses with more
	// pages, and 200 OK for the final/only page.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("cannot fetch heroku members: unexpected status %d", httpResp.StatusCode)
	}

	var members []herokuTeamMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, "", fmt.Errorf("cannot decode heroku members response: %w", err)
	}

	return members, httpResp.Header.Get("Next-Range"), nil
}
