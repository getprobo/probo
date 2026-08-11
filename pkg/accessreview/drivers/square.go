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
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	// Square API paths joined onto the driver's base URL.
	squareTeamMembersSearchPath = "team-members/search"
	squareMerchantsMePath       = "merchants/me"
	// squareAPIVersion pins the request version so behavior is deterministic
	// rather than following the application's console default.
	squareAPIVersion  = "2026-05-20"
	squareSearchLimit = 200
)

// SquareDriver lists the team members of a single Square merchant. A Square
// token — OAuth Bearer or Personal Access Token — is always scoped to one
// merchant, so POST /v2/team-members/search returns every team member of that
// merchant with no tenant selector. The search returns is_owner directly, so
// there is no role resolution.
type SquareDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*SquareDriver)(nil)

type squareTeamMember struct {
	ID           string `json:"id"`
	IsOwner      bool   `json:"is_owner"`
	Status       string `json:"status"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	EmailAddress string `json:"email_address"`
	CreatedAt    string `json:"created_at"`
}

type squareSearchResponse struct {
	TeamMembers []squareTeamMember `json:"team_members"`
	Cursor      string             `json:"cursor"`
}

// NewSquareDriver builds a driver against baseURL, the versioned Square
// Connect API origin (e.g. https://connect.squareup.com/v2).
func NewSquareDriver(httpClient *http.Client, baseURL string) *SquareDriver {
	return &SquareDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		baseURL: baseURL,
	}
}

func (d *SquareDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	cursor := ""

	for range maxPaginationPages {
		page, err := d.searchTeamMembers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, m := range page.TeamMembers {
			email := strings.TrimSpace(m.EmailAddress)
			if email == "" {
				continue
			}

			role := "member"
			if m.IsOwner {
				role = "owner"
			}

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    squareFullName(m, email),
				Roles:       ownerMemberRoles(role),
				Active:      activeFromStatus(m.Status),
				IsAdmin:     new(m.IsOwner),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(m.CreatedAt),
				ExternalID:  m.ID,
			})
		}

		if page.Cursor == "" {
			return records, nil
		}

		cursor = page.Cursor
	}

	return nil, fmt.Errorf("cannot list all square accounts: %w", ErrPaginationLimitReached)
}

func (d *SquareDriver) searchTeamMembers(ctx context.Context, cursor string) (*squareSearchResponse, error) {
	// No query filter: return team members of every status (ACTIVE and
	// INACTIVE) so deactivated members are reviewed too.
	reqBody := struct {
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor,omitempty"`
	}{
		Limit:  squareSearchLimit,
		Cursor: cursor,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal square search request: %w", err)
	}

	endpoint, err := url.JoinPath(d.baseURL, squareTeamMembersSearchPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build square team members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create square team members request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Square-Version", squareAPIVersion)

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute square team members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch square team members: unexpected status %d", httpResp.StatusCode)
	}

	var resp squareSearchResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode square team members response: %w", err)
	}

	return &resp, nil
}

// squareFullName joins the team member's given and family names, falling back
// to the email when Square exposes neither.
func squareFullName(m squareTeamMember, email string) string {
	name := strings.TrimSpace(strings.TrimSpace(m.GivenName) + " " + strings.TrimSpace(m.FamilyName))
	if name == "" {
		return email
	}

	return name
}

// squareNameResolver resolves the Square merchant's business name via
// GET /v2/merchants/me. A Square token — OAuth or PAT — is scoped to a single
// merchant, so "me" resolves it for both connection kinds.
type squareNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

// NewSquareNameResolver resolves the merchant name against baseURL, the
// versioned Square Connect API origin (e.g. https://connect.squareup.com/v2).
func NewSquareNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &squareNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *squareNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, squareMerchantsMePath)
	if err != nil {
		return "", fmt.Errorf("cannot build square merchant URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create square merchant request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Square-Version", squareAPIVersion)

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute square merchant request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	// A non-2xx (revoked token, missing scope) is terminal: keep the generic
	// source name rather than make the source-name worker retry forever.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Merchant struct {
			BusinessName string `json:"business_name"`
		} `json:"merchant"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode square merchant response: %w", err)
	}

	return resp.Merchant.BusinessName, nil
}
