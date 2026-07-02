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

type OpenAIDriver struct {
	httpClient *http.Client
}

var _ Driver = (*OpenAIDriver)(nil)

type openaiUsersResponse struct {
	Data []struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		AddedAt  int64  `json:"added_at"`
		Disabled bool   `json:"disabled"`
	} `json:"data"`
	HasMore bool   `json:"has_more"`
	LastID  string `json:"last_id"`
}

const openaiUsersEndpoint = "https://api.openai.com/v1/organization/users"

func NewOpenAIDriver(httpClient *http.Client) *OpenAIDriver {
	return &OpenAIDriver{
		httpClient: httpClient,
	}
}

func (d *OpenAIDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		after   string
	)

	for range maxPaginationPages {
		resp, err := d.fetchUsers(ctx, after)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data {
			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.Name,
				Roles:       openaiRoles(u.Role),
				Active:      new(!u.Disabled),
				IsAdmin:     u.Role == "owner",
				ExternalID:  u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
			}

			if u.AddedAt != 0 {
				t := time.Unix(u.AddedAt, 0)
				record.CreatedAt = &t
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if !resp.HasMore || resp.LastID == "" {
			return records, nil
		}

		after = resp.LastID
	}

	return nil, fmt.Errorf("cannot list all openai accounts: %w", ErrPaginationLimitReached)
}

func (d *OpenAIDriver) fetchUsers(ctx context.Context, after string) (*openaiUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openaiUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create openai users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", "100")

	if after != "" {
		q.Set("after", after)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute openai users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch openai users: unexpected status %d", httpResp.StatusCode)
	}

	var resp openaiUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode openai users response: %w", err)
	}

	return &resp, nil
}

func openaiRoles(role string) []string {
	if role == "" {
		return []string{}
	}

	switch role {
	case "owner":
		return []string{"Owner"}
	case "reader":
		return []string{"Reader"}
	default:
		return []string{"Member"}
	}
}
