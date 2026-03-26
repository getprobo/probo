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
	"errors"
	"fmt"
	"net/http"
	"time"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
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
		httpClient: httpClient,
	}
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
			Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := d.listUsersWithRetry(ctx, call)
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
			} else if u.IsEnforcedIn2Sv {
				// Enforced but not yet enrolled
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

			// Note: OrgUnitPath is an organizational unit (e.g. "/Engineering"),
			// not a job title. We map it to Role as an approximation.
			if u.OrgUnitPath != "" {
				rec.Role = u.OrgUnitPath
			}

			records = append(records, rec)
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return records, nil
}

func (d *GoogleWorkspaceDriver) listUsersWithRetry(
	ctx context.Context,
	call *admin.UsersListCall,
) (*admin.Users, error) {
	var lastErr error
	for attempt := range 3 {
		resp, err := call.Do()
		if err == nil {
			return resp, nil
		}

		lastErr = err
		apiErr := &googleapi.Error{}
		if ok := errors.As(err, &apiErr); !ok {
			return nil, err
		}
		if apiErr.Code != http.StatusTooManyRequests && apiErr.Code < 500 {
			return nil, err
		}

		backoff := time.Duration(250*(1<<attempt)) * time.Millisecond
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}

	return nil, lastErr
}
