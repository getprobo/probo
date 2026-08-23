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

	"encoding/json"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"net/url"
	"strconv"
	"time"
)

// OnePasswordDriver fetches user accounts from a 1Password SCIM bridge.
type OnePasswordDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*OnePasswordDriver)(nil)

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

func NewOnePasswordDriver(httpClient *http.Client, baseURL string) *OnePasswordDriver {
	return &OnePasswordDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *OnePasswordDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	startIndex := 1

	for range maxPaginationPages {
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
				Email:       email,
				FullName:    u.DisplayName,
				Active:      new(u.Active),
				ExternalID:  u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
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

			// Note: SCIM Meta.LastModified is the profile update time, not
			// the last login time, so we intentionally do not map it.

			if email != "" {
				records = append(records, record)
			}
		}

		if len(resp.Resources) == 0 || resp.ItemsPerPage <= 0 || startIndex+resp.ItemsPerPage > resp.TotalResults {
			return records, nil
		}

		startIndex += resp.ItemsPerPage
	}

	return nil, fmt.Errorf("cannot list all 1password accounts: %w", ErrPaginationLimitReached)
}

func (d *OnePasswordDriver) queryUsers(ctx context.Context, startIndex int) (*onePasswordSCIMListResponse, error) {
	u, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 1password base url: %w", err)
	}

	u = u.JoinPath("scim", "v2", "Users")
	q := u.Query()
	q.Set("startIndex", strconv.Itoa(startIndex))
	q.Set("count", "100")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
		return nil, fmt.Errorf("cannot fetch 1password users: unexpected status %d", httpResp.StatusCode)
	}

	var resp onePasswordSCIMListResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode 1password users response: %w", err)
	}

	return &resp, nil
}

func onePasswordSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := onePasswordSourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			nil,
			nil,
		), nil
	})
}

func onePasswordSourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	_ *log.Logger,
	_ provider.Endpoints,
) (Driver, error) {
	// The client-credentials grant uses the Users API driver.
	// Everything else is the API-key connection, whose
	// *APIKeyConnection makes GrantType() return "": it uses the
	// SCIM-bridge driver. 1Password declares no AuthURL/TokenURL, so
	// the authorization-code path is unreachable.
	if conn.GrantType() == string(connector.OAuth2GrantTypeClientCredentials) {
		s, err := coredata.ConnectorSettings[coredata.OnePasswordUsersAPISettings](conn)
		if err != nil {
			return nil, fmt.Errorf("cannot read 1password users api settings: %w", err)
		}

		return NewOnePasswordUsersAPIDriver(c, s.AccountID, s.Region), nil
	}

	s, err := coredata.ConnectorSettings[coredata.OnePasswordConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read 1password connector settings: %w", err)
	}

	if s.SCIMBridgeURL == "" {
		return nil, fmt.Errorf("cannot create 1password driver: scim_bridge_url is required")
	}

	return NewOnePasswordDriver(c, s.SCIMBridgeURL), nil
}
