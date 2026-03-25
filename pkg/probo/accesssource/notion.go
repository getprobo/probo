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

const (
	notionUsersEndpoint = "https://api.notion.com/v1/users"
	notionAPIVersion    = "2022-06-28"
)

// NotionDriver fetches workspace users from Notion via OAuth2-authenticated
// REST API requests.
type NotionDriver struct {
	httpClient *http.Client
}

func NewNotionDriver(httpClient *http.Client) *NotionDriver {
	return &NotionDriver{
		httpClient: httpClient,
	}
}

func (d *NotionDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records     []AccountRecord
		startCursor *string
	)

	for range maxPaginationPages {
		resp, err := d.queryUsers(ctx, startCursor)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Results {
			if u.Type != "person" {
				continue
			}

			record := AccountRecord{
				Email:      u.Person.Email,
				FullName:   u.Name,
				Role:       "Member",
				Active:     true,
				IsAdmin:    false,
				ExternalID: u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if !resp.HasMore || resp.NextCursor == "" {
			break
		}
		nextCursor := resp.NextCursor
		startCursor = &nextCursor
	}

	return records, nil
}

func (d *NotionDriver) queryUsers(ctx context.Context, startCursor *string) (*notionUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, notionUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create notion users request: %w", err)
	}

	req.Header.Set("Notion-Version", notionAPIVersion)
	req.Header.Set("Accept", "application/json")

	q := req.URL.Query()
	q.Set("page_size", "100")
	if startCursor != nil {
		q.Set("start_cursor", *startCursor)
	}
	req.URL.RawQuery = q.Encode()

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute notion users request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch notion users: unexpected status %d", httpResp.StatusCode)
	}

	var resp notionUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode notion users response: %w", err)
	}

	return &resp, nil
}

type notionUsersResponse struct {
	Results []struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Name   string `json:"name"`
		Person struct {
			Email string `json:"email"`
		} `json:"person"`
	} `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}
