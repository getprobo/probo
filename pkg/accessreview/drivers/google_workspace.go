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

	"errors"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"time"
)

// GoogleWorkspaceDriver fetches user accounts from Google Workspace
// using the Admin Directory API via an OAuth2-authenticated HTTP client.
//
// It carries no baseURL, unlike its sibling drivers: the Admin Directory
// host comes from the generated SDK's basePath, not from a literal here, so
// there is nothing for the registration's Endpoints.APIBase to inject. See
// the comment on googleWorkspaceRegistration.
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
				IsAdmin:     new(u.IsAdmin),
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

// googleWorkspaceNameResolver resolves the Google Workspace primary domain.
type googleWorkspaceNameResolver struct {
	httpClient *http.Client
}

func NewGoogleWorkspaceNameResolver(httpClient *http.Client) NameResolver {
	return &googleWorkspaceNameResolver{httpClient: httpClient}
}

func (r *googleWorkspaceNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	adminService, err := admin.NewService(ctx, option.WithHTTPClient(r.httpClient))
	if err != nil {
		return "", fmt.Errorf("cannot create google admin service: %w", err)
	}

	customer, err := adminService.Customers.Get("my_customer").Context(ctx).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusForbidden {
			return "", nil
		}

		return "", fmt.Errorf("cannot fetch google workspace customer: %w", err)
	}

	return customer.CustomerDomain, nil
}

func googleWorkspaceSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		return capable(
			NewGoogleWorkspaceDriver(credential.Client),
			NewGoogleWorkspaceNameResolver(credential.Client),
			nil,
		), nil
	})
}
