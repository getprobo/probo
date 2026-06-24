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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// GoogleWorkspaceDriver fetches user accounts from Google Workspace
// using the Admin Directory API via an OAuth2-authenticated HTTP client.
type GoogleWorkspaceDriver struct {
	httpClient *http.Client
}

func NewGoogleWorkspaceDriver(httpClient *http.Client) *GoogleWorkspaceDriver {
	return &GoogleWorkspaceDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

// retryRoundTripper retries requests that receive 5xx or 429 responses
// with exponential backoff.
type retryRoundTripper struct {
	next       http.RoundTripper
	maxRetries int
}

func (rt *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := rt.next
	if transport == nil {
		transport = http.DefaultTransport
	}

	var lastResp *http.Response

	for attempt := range rt.maxRetries {
		resp, err := transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return resp, nil
		}

		// Buffer and re-attach the body so the caller can still read it
		// if this turns out to be the final (retry-exhausted) response.
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))
		lastResp = resp

		backoff := time.Duration(250*(1<<attempt)) * time.Millisecond
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(backoff):
		}
	}

	return lastResp, nil
}

func (d *GoogleWorkspaceDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	adminService, err := admin.NewService(ctx, option.WithHTTPClient(d.httpClient))
	if err != nil {
		return nil, fmt.Errorf("cannot create google admin service: %w", err)
	}

	var records []AccountRecord

	pageToken := ""

	for range maxPaginationPages {
		call := adminService.Users.List().
			Customer("my_customer").
			MaxResults(500).
			Projection("full").
			Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("cannot list google workspace users: %w", err)
		}

		for _, u := range resp.Users {
			rec := AccountRecord{
				Email:       u.PrimaryEmail,
				FullName:    u.Name.FullName,
				Active:      new(!u.Suspended && !u.Archived),
				IsAdmin:     u.IsAdmin,
				ExternalID:  u.Id,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
			}

			if u.IsEnrolledIn2Sv {
				rec.MFAStatus = coredata.MFAStatusEnabled
			} else {
				rec.MFAStatus = coredata.MFAStatusDisabled
			}

			if u.CreationTime != "" {
				if t, err := time.Parse(time.RFC3339, u.CreationTime); err == nil {
					rec.CreatedAt = &t
				}
			}

			if u.LastLoginTime != "" {
				if t, err := time.Parse(time.RFC3339, u.LastLoginTime); err == nil {
					rec.LastLogin = &t
				}
			}

			switch {
			case u.IsAdmin:
				rec.Roles = []string{"Super Admin"}
			case u.IsDelegatedAdmin:
				rec.Roles = []string{"Delegated Admin"}
			default:
				rec.Roles = []string{"User"}
			}

			records = append(records, rec)
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all google workspace accounts: %w", ErrPaginationLimitReached)
}
