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

		record := AccountRecord{
			Email:       m.Email,
			FullName:    m.UserName,
			Role:        m.RoleName,
			IsAdmin:     isAdmin,
			ExternalID:  m.UserID,
			MFAStatus:   mfaStatus,
			AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
			AccountType: coredata.AccessEntryAccountTypeUser,
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
