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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

type AnthropicDriver struct {
	httpClient *http.Client
}

var _ Driver = (*AnthropicDriver)(nil)

const (
	anthropicUsersEndpoint = "https://api.anthropic.com/v1/organizations/users"
	// anthropicAPIVersion is the required anthropic-version header value
	// sent on every Admin API request. Shared with the name resolver.
	anthropicAPIVersion = "2023-06-01"
)

type anthropicUsersResponse struct {
	Data []struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Role    string `json:"role"`
		AddedAt string `json:"added_at"`
	} `json:"data"`
	HasMore bool   `json:"has_more"`
	LastID  string `json:"last_id"`
}

func NewAnthropicDriver(httpClient *http.Client) *AnthropicDriver {
	return &AnthropicDriver{
		httpClient: httpClient,
	}
}

func (d *AnthropicDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		afterID string
	)

	for range maxPaginationPages {
		resp, err := d.fetchUsers(ctx, afterID)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data {
			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.Name,
				Roles:       anthropicRoles(u.Role),
				IsAdmin:     u.Role == "admin",
				ExternalID:  u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
			}

			// added_at is an RFC 3339 datetime string; ignore parse
			// failures rather than dropping the record.
			if u.AddedAt != "" {
				if t, err := time.Parse(time.RFC3339, u.AddedAt); err == nil {
					record.CreatedAt = &t
				}
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if !resp.HasMore || resp.LastID == "" {
			return records, nil
		}

		afterID = resp.LastID
	}

	return nil, fmt.Errorf("cannot list all anthropic accounts: %w", ErrPaginationLimitReached)
}

func (d *AnthropicDriver) fetchUsers(ctx context.Context, afterID string) (*anthropicUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, anthropicUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create anthropic users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", "100")

	if afterID != "" {
		q.Set("after_id", afterID)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute anthropic users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch anthropic users: unexpected status %d", httpResp.StatusCode)
	}

	var resp anthropicUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode anthropic users response: %w", err)
	}

	return &resp, nil
}

func anthropicRoles(role string) []string {
	if role == "" {
		return []string{}
	}

	switch role {
	case "admin":
		return []string{"Admin"}
	case "billing":
		return []string{"Billing"}
	case "developer":
		return []string{"Developer"}
	case "claude_code_user":
		return []string{"Claude Code User"}
	case "user":
		return []string{"User"}
	default:
		return []string{role}
	}
}
