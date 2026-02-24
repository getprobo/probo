// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	linearGraphQLEndpoint = "https://api.linear.app/graphql"
)

// LinearDriver fetches workspace users from Linear via OAuth2-authenticated
// GraphQL requests.
type LinearDriver struct {
	httpClient *http.Client
}

func NewLinearDriver(httpClient *http.Client) *LinearDriver {
	return &LinearDriver{
		httpClient: httpClient,
	}
}

func (d *LinearDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		after   *string
	)

	for {
		resp, err := d.queryUsers(ctx, after)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data.Users.Nodes {
			record := AccountRecord{
				Email:      u.Email,
				FullName:   u.Name,
				Active:     u.Active,
				IsAdmin:    u.Admin,
				ExternalID: u.ID,
				MFAStatus:  coredata.MFAStatusUnknown,
				AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if !resp.Data.Users.PageInfo.HasNextPage || resp.Data.Users.PageInfo.EndCursor == "" {
			break
		}
		nextCursor := resp.Data.Users.PageInfo.EndCursor
		after = &nextCursor
	}

	return records, nil
}

func (d *LinearDriver) queryUsers(ctx context.Context, after *string) (*linearUsersResponse, error) {
	const query = `
query AccessReviewLinearUsers($after: String) {
  users(first: 100, after: $after) {
    nodes {
      id
      email
      name
      active
      admin
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

	body := linearUsersRequest{
		Query: query,
		Variables: linearUsersVariables{
			After: after,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal linear users query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, linearGraphQLEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("cannot create linear users request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute linear users request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("linear users request failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var resp linearUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode linear users response: %w", err)
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("linear graphql error: %s", resp.Errors[0].Message)
	}

	return &resp, nil
}

type linearUsersRequest struct {
	Query     string               `json:"query"`
	Variables linearUsersVariables `json:"variables"`
}

type linearUsersVariables struct {
	After *string `json:"after"`
}

type linearUsersResponse struct {
	Data struct {
		Users struct {
			Nodes []struct {
				ID     string `json:"id"`
				Email  string `json:"email"`
				Name   string `json:"name"`
				Active bool   `json:"active"`
				Admin  bool   `json:"admin"`
			} `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"users"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}
