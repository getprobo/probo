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
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// LangfuseDriver lists the members of a single Langfuse organization via the
// organization-scoped public API. The organization API key (HTTP Basic,
// publicKey:secretKey) is bound to one organization on the configured host,
// so GET /api/public/organizations/memberships returns every member with no
// tenant selector. The Basic credential is applied by the connection
// transport; the base URL spans the regional cloud hosts and self-hosting.
type LangfuseDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*LangfuseDriver)(nil)

type langfuseMembershipsResponse struct {
	Memberships []langfuseMembership `json:"memberships"`
}

type langfuseMembership struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func NewLangfuseDriver(httpClient *http.Client, baseURL string) *LangfuseDriver {
	return &LangfuseDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *LangfuseDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	baseURL, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse langfuse base URL: %w", err)
	}

	endpoint := baseURL.JoinPath("api", "public", "organizations", "memberships")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create langfuse memberships request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute langfuse memberships request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch langfuse memberships: unexpected status %d", httpResp.StatusCode)
	}

	var resp langfuseMembershipsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode langfuse memberships response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Memberships))

	for _, m := range resp.Memberships {
		email := strings.TrimSpace(m.Email)
		if email == "" {
			continue
		}

		records = append(records, AccountRecord{
			Email:       email,
			FullName:    langfuseFullName(m, email),
			Roles:       langfuseRoles(m.Role),
			IsAdmin:     langfuseIsAdmin(m.Role),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(m.UserID),
		})
	}

	return records, nil
}

func langfuseFullName(m langfuseMembership, fallback string) string {
	if name := strings.TrimSpace(m.Name); name != "" {
		return name
	}

	return fallback
}

// langfuseRoles maps the Langfuse organization MembershipRole
// (OWNER/ADMIN/MEMBER/VIEWER, and the RBAC NONE) to a stable display label,
// preserving unknown future roles verbatim.
func langfuseRoles(role string) []string {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "OWNER":
		return []string{"Owner"}
	case "ADMIN":
		return []string{"Admin"}
	case "MEMBER":
		return []string{"Member"}
	case "VIEWER":
		return []string{"Viewer"}
	case "NONE":
		return []string{"None"}
	default:
		if r := strings.TrimSpace(role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}

func langfuseIsAdmin(role string) bool {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "OWNER", "ADMIN":
		return true
	default:
		return false
	}
}
