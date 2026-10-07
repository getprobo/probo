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
	supabaseOrganizationsPath = "organizations"
	supabaseMembersPath       = "members"
)

type SupabaseDriver struct {
	httpClient *http.Client
	orgSlug    string
	baseURL    string
}

var _ Driver = (*SupabaseDriver)(nil)

type supabaseMember struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	UserName   string `json:"user_name"`
	RoleName   string `json:"role_name"`
	MFAEnabled bool   `json:"mfa_enabled"`
}

func NewSupabaseDriver(httpClient *http.Client, orgSlug, baseURL string) *SupabaseDriver {
	return &SupabaseDriver{
		httpClient: httpClient,
		orgSlug:    orgSlug,
		baseURL:    baseURL,
	}
}

func (d *SupabaseDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	members, err := d.queryMembers(ctx)
	if err != nil {
		return nil, err
	}

	var records []AccountRecord

	for _, m := range members {
		mfaStatus := coredata.MFAStatusDisabled
		if m.MFAEnabled {
			mfaStatus = coredata.MFAStatusEnabled
		}

		isAdmin := m.RoleName == "Owner" || m.RoleName == "Administrator"

		role := strings.TrimSpace(m.RoleName)

		roles := []string{}
		if role != "" {
			roles = []string{role}
		}

		record := AccountRecord{
			Email:       m.Email,
			FullName:    m.UserName,
			Roles:       roles,
			IsAdmin:     new(isAdmin),
			ExternalID:  m.UserID,
			MFAStatus:   mfaStatus,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}

		records = append(records, record)
	}

	return records, nil
}

func (d *SupabaseDriver) queryMembers(ctx context.Context) ([]supabaseMember, error) {
	endpoint, err := SupabaseMembersURL(d.baseURL, d.orgSlug)
	if err != nil {
		return nil, fmt.Errorf("cannot build supabase members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create supabase members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute supabase members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if rejected := SupabaseMembersRejection(httpResp); rejected != nil {
		return nil, rejected
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"cannot fetch supabase members: unexpected status %d",
			httpResp.StatusCode,
		)
	}

	var members []supabaseMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("cannot decode supabase members response: %w", err)
	}

	return members, nil
}

// SupabaseMembersURL returns the members URL of the organization orgSlug.
func SupabaseMembersURL(baseURL, orgSlug string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse supabase base URL: %w", err)
	}

	return base.JoinPath(supabaseOrganizationsPath, url.PathEscape(orgSlug), supabaseMembersPath).String(), nil
}

// Codes for the organization members refusals: a slug no organization has, and
// an organization the token cannot reach.
const (
	SupabaseOrganizationNotFound      = "supabase_organization_not_found"
	SupabaseOrganizationNotAccessible = "supabase_organization_not_accessible"
)

// SupabaseMembersRejection returns the refusal a members response carries, or
// nil.
func SupabaseMembersRejection(resp *http.Response) *SettingRejectedError {
	if !respondsWithJSON(resp) {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return &SettingRejectedError{
			Code:       SupabaseOrganizationNotFound,
			StatusCode: resp.StatusCode,
			Message:    "Supabase has no organization with this slug. Copy it from the organization URL in the Supabase dashboard (supabase.com/dashboard/org/<slug>).",
		}
	case http.StatusForbidden:
		return &SettingRejectedError{
			Code:       SupabaseOrganizationNotAccessible,
			StatusCode: resp.StatusCode,
			Message:    "This access token cannot read the members of this organization. Create a token with Organization resource access that includes this organization, and Organization Members set to Read.",
		}
	default:
		return nil
	}
}

// supabaseNameResolver returns the Supabase organization slug as the name.
type supabaseNameResolver struct {
	orgSlug string
}

func NewSupabaseNameResolver(orgSlug string) NameResolver {
	return &supabaseNameResolver{orgSlug: orgSlug}
}

func (r *supabaseNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return r.orgSlug, nil
}
