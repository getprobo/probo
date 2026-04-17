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

	"go.probo.inc/probo/pkg/coredata"
)

// IntercomDriver fetches workspace admins from Intercom via Bearer
// token-authenticated REST API requests.
type IntercomDriver struct {
	httpClient *http.Client
}

var _ Driver = (*IntercomDriver)(nil)

type intercomAdminsResponse struct {
	Type   string `json:"type"`
	Admins []struct {
		Type         string `json:"type"`
		ID           string `json:"id"`
		Name         string `json:"name"`
		Email        string `json:"email"`
		JobTitle     string `json:"job_title"`
		HasInboxSeat bool   `json:"has_inbox_seat"`
	} `json:"admins"`
}

const (
	intercomAdminsEndpoint = "https://api.intercom.io/admins"
	intercomAPIVersion     = "2.11"
)

func NewIntercomDriver(httpClient *http.Client) *IntercomDriver {
	return &IntercomDriver{
		httpClient: httpClient,
	}
}

func (d *IntercomDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	resp, err := d.fetchAdmins(ctx)
	if err != nil {
		return nil, err
	}

	var records []AccountRecord
	for _, a := range resp.Admins {
		record := AccountRecord{
			Email:       a.Email,
			FullName:    a.Name,
			Role:        intercomRole(a.HasInboxSeat),
			JobTitle:    a.JobTitle,
			IsAdmin:     false, // Intercom API does not expose admin role information
			ExternalID:  a.ID,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
			AccountType: coredata.AccessEntryAccountTypeUser,
		}

		if record.Email != "" || record.FullName != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

func (d *IntercomDriver) fetchAdmins(ctx context.Context) (*intercomAdminsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, intercomAdminsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create intercom admins request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Intercom-Version", intercomAPIVersion)

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute intercom admins request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch intercom admins: unexpected status %d", httpResp.StatusCode)
	}

	var resp intercomAdminsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode intercom admins response: %w", err)
	}

	return &resp, nil
}

// intercomRole returns a role label based on whether the admin has an inbox
// seat. The Intercom API does not expose a proper role field, so this is the
// best approximation available: users with inbox seats are active agents,
// those without are limited/viewer users.
func intercomRole(hasInboxSeat bool) string {
	if hasInboxSeat {
		return "Agent"
	}
	return "Viewer"
}
