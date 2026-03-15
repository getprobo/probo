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

package accesssource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.probo.inc/probo/pkg/coredata"
)

const brexUsersEndpoint = "https://platform.brexapis.com/v2/users"

// BrexDriver fetches users from Brex via OAuth2-authenticated REST API
// requests.
type BrexDriver struct {
	httpClient *http.Client
}

func NewBrexDriver(httpClient *http.Client) *BrexDriver {
	return &BrexDriver{
		httpClient: httpClient,
	}
}

func (d *BrexDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		cursor  *string
	)

	for {
		resp, err := d.queryUsers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Items {
			record := AccountRecord{
				Email:      u.Email,
				FullName:   u.FirstName + " " + u.LastName,
				Role:       u.Role,
				Active:     u.Status == "ACTIVE",
				IsAdmin:    false,
				ExternalID: u.ID,
				MFAStatus:  coredata.MFAStatusUnknown,
				AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if resp.NextCursor == "" {
			break
		}
		nextCursor := resp.NextCursor
		cursor = &nextCursor
	}

	return records, nil
}

func (d *BrexDriver) queryUsers(ctx context.Context, cursor *string) (*brexUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, brexUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create brex users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if cursor != nil {
		q := req.URL.Query()
		q.Set("cursor", *cursor)
		req.URL.RawQuery = q.Encode()
	}

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute brex users request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch brex users: unexpected status %d", httpResp.StatusCode)
	}

	var resp brexUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode brex users response: %w", err)
	}

	return &resp, nil
}

type brexUsersResponse struct {
	Items []struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Status    string `json:"status"`
		Role      string `json:"role"`
	} `json:"items"`
	NextCursor string `json:"next_cursor"`
}
