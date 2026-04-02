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

package drivers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"go.probo.inc/probo/pkg/coredata"
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

		_ = resp.Body.Close()
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
				Active:      !u.Suspended && !u.Archived,
				IsAdmin:     u.IsAdmin,
				ExternalID:  u.Id,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodSSO,
				AccountType: coredata.AccessEntryAccountTypeUser,
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
				rec.Role = "Super Admin"
			case u.IsDelegatedAdmin:
				rec.Role = "Delegated Admin"
			default:
				rec.Role = "User"
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
