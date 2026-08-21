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

type (
	RipplingDriver struct {
		httpClient *http.Client
		baseURL    string
	}

	ripplingUsersResponse struct {
		Results  []ripplingUser `json:"results"`
		NextLink string         `json:"next_link"`
	}

	ripplingUser struct {
		ID          string `json:"id"`
		CreatedAt   string `json:"created_at"`
		Active      *bool  `json:"active"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Name        struct {
			Formatted string `json:"formatted"`
		} `json:"name"`
		Emails []struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"emails"`
	}
)

const ripplingUsersPath = "users"

var _ Driver = (*RipplingDriver)(nil)

func NewRipplingDriver(httpClient *http.Client, baseURL string) *RipplingDriver {
	return &RipplingDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *RipplingDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, ripplingUsersPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build rippling users URL: %w", err)
	}

	next := endpoint
	var records []AccountRecord

	for range maxPaginationPages {
		resp, err := d.fetchUsersPage(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, user := range resp.Results {
			email := ripplingEmail(user)
			if email == "" {
				continue
			}

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    ripplingFullName(user),
				Active:      user.Active,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(user.CreatedAt),
				ExternalID:  user.ID,
			})
		}

		if resp.NextLink == "" {
			return records, nil
		}

		next, err = sameHostNextPageURL("rippling", endpoint, resp.NextLink)
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("cannot list all rippling accounts: %w", ErrPaginationLimitReached)
}

func (d *RipplingDriver) fetchUsersPage(ctx context.Context, endpoint string) (*ripplingUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create rippling users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute rippling users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch rippling users: unexpected status %d", httpResp.StatusCode)
	}

	var resp ripplingUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode rippling users response: %w", err)
	}

	return &resp, nil
}

func ripplingEmail(user ripplingUser) string {
	for _, email := range user.Emails {
		if strings.EqualFold(email.Type, "WORK") && email.Value != "" {
			return email.Value
		}
	}

	for _, email := range user.Emails {
		if email.Value != "" {
			return email.Value
		}
	}

	return ""
}

func ripplingFullName(user ripplingUser) string {
	for _, name := range []string{user.DisplayName, user.Name.Formatted, user.Username} {
		if name := strings.TrimSpace(name); name != "" {
			return name
		}
	}

	return ripplingEmail(user)
}
