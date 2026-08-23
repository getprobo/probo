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
)

// IntercomDriver fetches workspace admins from Intercom via Bearer
// token-authenticated REST API requests.
type IntercomDriver struct {
	httpClient *http.Client
	baseURL    string
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
	intercomAdminsPath = "/admins"
	intercomMePath     = "/me"
	intercomAPIVersion = "2.11"
)

func NewIntercomDriver(httpClient *http.Client, baseURL string) *IntercomDriver {
	return &IntercomDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
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
			Roles:       intercomRoles(a.HasInboxSeat),
			JobTitle:    a.JobTitle,
			ExternalID:  a.ID,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}

		if record.Email != "" || record.FullName != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

func (d *IntercomDriver) fetchAdmins(ctx context.Context) (*intercomAdminsResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, intercomAdminsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build intercom admins URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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
func intercomRoles(hasInboxSeat bool) []string {
	if hasInboxSeat {
		return []string{"Agent"}
	}

	return []string{"Viewer"}
}

// intercomNameResolver resolves the Intercom app name.
type intercomNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

func NewIntercomNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &intercomNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *intercomNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, intercomMePath)
	if err != nil {
		return "", fmt.Errorf("cannot build intercom me URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create intercom me request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Intercom-Version", intercomAPIVersion)

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute intercom me request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		App struct {
			Name string `json:"name"`
		} `json:"app"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode intercom me response: %w", err)
	}

	return resp.App.Name, nil
}

func intercomSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		return capable(
			NewIntercomDriver(credential.Client, opened.Endpoints.APIBase),
			NewIntercomNameResolver(credential.Client, opened.Endpoints.APIBase),
			nil,
		), nil
	})
}
