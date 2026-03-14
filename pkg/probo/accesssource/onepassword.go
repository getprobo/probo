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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// OnePasswordDriver fetches user accounts from a 1Password SCIM bridge.
type OnePasswordDriver struct {
	httpClient *http.Client
	baseURL    string
}

func NewOnePasswordDriver(httpClient *http.Client, baseURL string) *OnePasswordDriver {
	return &OnePasswordDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *OnePasswordDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord
	startIndex := 1

	for {
		resp, err := d.queryUsers(ctx, startIndex)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Resources {
			email := u.UserName
			if email == "" {
				for _, e := range u.Emails {
					if e.Primary {
						email = e.Value
						break
					}
				}
			}

			record := AccountRecord{
				Email:      email,
				FullName:   u.DisplayName,
				Active:     u.Active,
				ExternalID: u.ID,
				MFAStatus:  coredata.MFAStatusUnknown,
				AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			}

			if record.FullName == "" && u.Name.Formatted != "" {
				record.FullName = u.Name.Formatted
			}
			if record.FullName == "" && (u.Name.GivenName != "" || u.Name.FamilyName != "") {
				record.FullName = u.Name.GivenName + " " + u.Name.FamilyName
			}

			if u.Title != "" {
				record.JobTitle = u.Title
			}

			if u.Meta.Created != "" {
				if t, err := time.Parse(time.RFC3339, u.Meta.Created); err == nil {
					record.CreatedAt = &t
				}
			}

			if u.Meta.LastModified != "" {
				if t, err := time.Parse(time.RFC3339, u.Meta.LastModified); err == nil {
					record.LastLogin = &t
				}
			}

			if email != "" {
				records = append(records, record)
			}
		}

		if startIndex+resp.ItemsPerPage > resp.TotalResults {
			break
		}
		startIndex += resp.ItemsPerPage
	}

	return records, nil
}

func (d *OnePasswordDriver) queryUsers(ctx context.Context, startIndex int) (*onePasswordSCIMListResponse, error) {
	url := fmt.Sprintf("%s/scim/v2/Users?startIndex=%d&count=100", d.baseURL, startIndex)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create 1password users request: %w", err)
	}
	req.Header.Set("Accept", "application/scim+json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute 1password users request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("1password users request failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var resp onePasswordSCIMListResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode 1password users response: %w", err)
	}

	return &resp, nil
}

type onePasswordSCIMListResponse struct {
	TotalResults int                   `json:"totalResults"`
	StartIndex   int                   `json:"startIndex"`
	ItemsPerPage int                   `json:"itemsPerPage"`
	Resources    []onePasswordSCIMUser `json:"Resources"`
}

type onePasswordSCIMUser struct {
	ID          string `json:"id"`
	UserName    string `json:"userName"`
	DisplayName string `json:"displayName"`
	Title       string `json:"title"`
	Active      bool   `json:"active"`
	Name        struct {
		Formatted  string `json:"formatted"`
		GivenName  string `json:"givenName"`
		FamilyName string `json:"familyName"`
	} `json:"name"`
	Emails []struct {
		Value   string `json:"value"`
		Primary bool   `json:"primary"`
	} `json:"emails"`
	Meta struct {
		Created      string `json:"created"`
		LastModified string `json:"lastModified"`
	} `json:"meta"`
}
