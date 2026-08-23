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
	"fmt"
	"net/http"

	"bytes"
	"encoding/json"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"strings"
	"time"
)

// LinearDriver fetches workspace users from Linear via OAuth2-authenticated
// GraphQL requests.
type LinearDriver struct {
	httpClient *http.Client
	endpoint   string
}

var _ Driver = (*LinearDriver)(nil)

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
				ID        string `json:"id"`
				Email     string `json:"email"`
				Name      string `json:"name"`
				Active    bool   `json:"active"`
				Admin     bool   `json:"admin"`
				Guest     bool   `json:"guest"`
				LastSeen  string `json:"lastSeen"`
				CreatedAt string `json:"createdAt"`
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

// NewLinearDriver builds a driver against endpoint, Linear's GraphQL API
// (e.g. https://api.linear.app/graphql). Linear exposes a single endpoint,
// so there is no path to join onto it.
func NewLinearDriver(httpClient *http.Client, endpoint string) *LinearDriver {
	return &LinearDriver{
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

func (d *LinearDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		after   *string
	)

	for range maxPaginationPages {
		resp, err := d.queryUsers(ctx, after)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data.Users.Nodes {
			accountType := coredata.AccessReviewEntryAccountTypeUser
			if strings.HasSuffix(u.Email, ".linear.app") {
				accountType = coredata.AccessReviewEntryAccountTypeServiceAccount
			}

			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.Name,
				Roles:       linearRoles(u.Admin, u.Guest),
				Active:      new(u.Active),
				IsAdmin:     new(u.Admin),
				ExternalID:  u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: accountType,
			}

			if u.LastSeen != "" {
				if t, err := time.Parse(time.RFC3339, u.LastSeen); err == nil {
					record.LastLogin = &t
				}
			}

			if u.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, u.CreatedAt); err == nil {
					record.CreatedAt = &t
				}
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if !resp.Data.Users.PageInfo.HasNextPage || resp.Data.Users.PageInfo.EndCursor == "" {
			return records, nil
		}

		nextCursor := resp.Data.Users.PageInfo.EndCursor
		after = &nextCursor
	}

	return nil, fmt.Errorf("cannot list all linear accounts: %w", ErrPaginationLimitReached)
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
      guest
      lastSeen
      createdAt
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, bytes.NewReader(payload))
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
		return nil, fmt.Errorf("cannot fetch linear users: unexpected status %d", httpResp.StatusCode)
	}

	var resp linearUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode linear users response: %w", err)
	}

	if len(resp.Errors) > 0 {
		// Provider-supplied messages may carry tenant identifiers or
		// query fragments — never embed them. Resolver scrubs the same
		// field; keep both call sites aligned.
		return nil, fmt.Errorf("cannot fetch linear users: graphql error")
	}

	return &resp, nil
}

func linearRoles(admin, guest bool) []string {
	switch {
	case admin:
		return []string{"Admin"}
	case guest:
		return []string{"Guest"}
	default:
		return []string{"Member"}
	}
}

// linearNameResolver resolves the Linear organization name via GraphQL.
type linearNameResolver struct {
	httpClient *http.Client
	endpoint   string
}

// NewLinearNameResolver resolves the org name against endpoint, Linear's
// GraphQL API (e.g. https://api.linear.app/graphql).
func NewLinearNameResolver(httpClient *http.Client, endpoint string) NameResolver {
	return &linearNameResolver{httpClient: httpClient, endpoint: endpoint}
}

func (r *linearNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	body := struct {
		Query string `json:"query"`
	}{
		Query: `{ organization { name } }`,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("cannot marshal linear organization query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("cannot create linear organization request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute linear organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("linear organization", httpResp.StatusCode)
	}

	var resp struct {
		Data struct {
			Organization struct {
				Name string `json:"name"`
			} `json:"organization"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode linear organization response: %w", err)
	}

	if len(resp.Errors) > 0 {
		// Provider-supplied messages may carry tenant identifiers or
		// query fragments — never embed them. Driver scrubs the same
		// field; keep both call sites aligned.
		return "", fmt.Errorf("cannot fetch linear organization: graphql error")
	}

	return resp.Data.Organization.Name, nil
}

func linearSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		return capable(
			NewLinearDriver(credential.Client, opened.Endpoints.APIBase),
			NewLinearNameResolver(credential.Client, opened.Endpoints.APIBase),
			nil,
		), nil
	})
}
