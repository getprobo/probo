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

const (
	// herokuPersonalAccountSlug is the reserved org-picker slug for a
	// personal Heroku account (one with no Team). Heroku Teams are an
	// opt-in paid construct, so a solo account has nothing in GET /teams;
	// selecting this entry runs the driver in personal mode (app owner +
	// collaborators) instead of team-member mode.
	herokuPersonalAccountSlug = "@personal"

	// herokuPersonalAccountDisplayName is the picker label and source name
	// shown for a personal Heroku account.
	herokuPersonalAccountDisplayName = "Personal account"
)

// HerokuDriver fetches members from the Heroku Platform API using a
// pre-authenticated HTTP client (Bearer token). Pagination is via Heroku's
// Range / Next-Range header pair (RFC 7233 style).
//
// The driver runs in one of two modes:
//   - team mode (teamID set): list the members of GET /teams/{id}/members.
//   - personal mode (teamID empty or the personal-account slug): a solo
//     Heroku account has no Team, so access is granted per-app; enumerate
//     the personal apps' owners and collaborators instead.
//
// Notes on data quality:
//   - The team-members endpoint does not expose suspension state, so
//     Active is left nil for v1.
//   - For federated teams the IdP is the source of truth for MFA, but
//     the API still reports `two_factor_authentication`. The driver
//     populates MFAStatus from that field and lets the access-review
//     UI surface federation context separately.
//   - The collaborators endpoint does not expose MFA, so personal-mode
//     records leave MFAStatus unknown.
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

type herokuApp struct {
	ID    string `json:"id"`
	Owner struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"owner"`
	// Team is nil for personal apps and set for team-owned apps; we use it
	// to keep personal mode scoped to the user's own apps.
	Team *struct {
		ID string `json:"id"`
	} `json:"team"`
}

type herokuCollaborator struct {
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	User      struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

func (d *HerokuDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	if d.teamID == "" || d.teamID == herokuPersonalAccountSlug {
		return d.listPersonalAccounts(ctx)
	}

	return d.listTeamMembers(ctx)
}

func (d *HerokuDriver) listTeamMembers(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath("https://api.heroku.com", "teams", url.PathEscape(d.teamID), "members")
	if err != nil {
		return nil, fmt.Errorf("cannot build heroku members URL: %w", err)
	}

	members, err := herokuListAll[herokuTeamMember](ctx, d.httpClient, endpoint, "members")
	if err != nil {
		return nil, fmt.Errorf("cannot list heroku team members: %w", err)
	}

	var records []AccountRecord

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

	return records, nil
}

// listPersonalAccounts enumerates the people with access to a personal
// (non-team) Heroku account: the owner of each personal app plus every
// collaborator on it, deduplicated by Heroku user ID. Personal accounts
// have no Team, so there are no team members to list; access is granted
// per-app via collaborators.
func (d *HerokuDriver) listPersonalAccounts(ctx context.Context) ([]AccountRecord, error) {
	// GET /apps returns every app the token can see, both owned and
	// collaborated-on. We skip team-owned apps below (those belong to a
	// team-scoped review); the remaining personal apps are the account
	// under review. Scoping strictly to apps the connector owns would need
	// the account ID from GET /account, which requires the identity OAuth
	// scope we deliberately do not request.
	apps, err := herokuListAll[herokuApp](ctx, d.httpClient, "https://api.heroku.com/apps", "apps")
	if err != nil {
		return nil, fmt.Errorf("cannot list heroku personal apps: %w", err)
	}

	var records []AccountRecord

	// Dedupe by Heroku user ID across all apps: a user collaborating on
	// several apps is one account in the review. An admin grant on any
	// app wins.
	seen := make(map[string]int)

	upsert := func(rec AccountRecord) {
		key := rec.ExternalID
		if key == "" {
			key = rec.Email
		}

		if key == "" {
			return
		}

		if i, ok := seen[key]; ok {
			if rec.IsAdmin {
				records[i].IsAdmin = true
			}

			return
		}

		seen[key] = len(records)
		records = append(records, rec)
	}

	for _, app := range apps {
		// Skip team-owned apps; those belong to a team-scoped review.
		if app.Team != nil {
			continue
		}

		// The owner always has access, whether or not they also appear in
		// the collaborators list.
		upsert(herokuPersonalRecord(app.Owner.ID, app.Owner.Email, "owner", true))

		endpoint, err := url.JoinPath("https://api.heroku.com", "apps", url.PathEscape(app.ID), "collaborators")
		if err != nil {
			return nil, fmt.Errorf("cannot build heroku collaborators URL: %w", err)
		}

		collaborators, err := herokuListAll[herokuCollaborator](ctx, d.httpClient, endpoint, "collaborators")
		if err != nil {
			return nil, fmt.Errorf("cannot list heroku collaborators for app %q: %w", app.ID, err)
		}

		for _, c := range collaborators {
			record := herokuPersonalRecord(c.User.ID, c.User.Email, c.Role, c.Role == "admin" || c.Role == "owner")

			if c.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, c.CreatedAt); err == nil {
					record.CreatedAt = &t
				}
			}

			upsert(record)
		}
	}

	return records, nil
}

// herokuPersonalRecord builds an AccountRecord for a person with access to a
// personal Heroku app. These endpoints expose no display name or MFA signal,
// so the email doubles as the full name and MFA is left unknown.
func herokuPersonalRecord(externalID, email, role string, isAdmin bool) AccountRecord {
	return AccountRecord{
		Email:       email,
		FullName:    email,
		Role:        role,
		IsAdmin:     isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
		AccountType: coredata.AccessEntryAccountTypeUser,
		ExternalID:  externalID,
	}
}

// herokuListAll fetches every page of a Heroku collection endpoint,
// following the Range / Next-Range pagination header pair, and decodes each
// page into T. label names the resource in error messages (e.g. "members").
func herokuListAll[T any](ctx context.Context, client *http.Client, endpoint, label string) ([]T, error) {
	var all []T

	rangeHeader := ""

	for range maxPaginationPages {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create heroku %s request: %w", label, err)
		}

		req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

		if rangeHeader != "" {
			req.Header.Set("Range", rangeHeader)
		}

		httpResp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("cannot execute heroku %s request: %w", label, err)
		}

		// Heroku returns 206 Partial Content for ranged responses with more
		// pages, and 200 OK for the final/only page.
		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
			_ = httpResp.Body.Close()
			return nil, fmt.Errorf("cannot fetch heroku %s: unexpected status %d", label, httpResp.StatusCode)
		}

		var page []T
		if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
			_ = httpResp.Body.Close()
			return nil, fmt.Errorf("cannot decode heroku %s response: %w", label, err)
		}

		nextRange := httpResp.Header.Get("Next-Range")
		_ = httpResp.Body.Close()

		all = append(all, page...)

		if nextRange == "" {
			return all, nil
		}

		rangeHeader = nextRange
	}

	return nil, fmt.Errorf("cannot list all heroku %s: %w", label, ErrPaginationLimitReached)
}
