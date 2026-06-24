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

type SupabaseDriver struct {
	httpClient *http.Client
	orgSlug    string
}

var _ Driver = (*SupabaseDriver)(nil)

type supabaseMember struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	UserName   string `json:"user_name"`
	RoleName   string `json:"role_name"`
	MFAEnabled bool   `json:"mfa_enabled"`
}

func NewSupabaseDriver(httpClient *http.Client, orgSlug string) *SupabaseDriver {
	return &SupabaseDriver{
		httpClient: httpClient,
		orgSlug:    orgSlug,
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
			IsAdmin:     isAdmin,
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
	u := &url.URL{
		Scheme: "https",
		Host:   "api.supabase.com",
	}
	u = u.JoinPath("v1", "organizations", d.orgSlug, "members")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
