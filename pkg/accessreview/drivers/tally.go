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
	"errors"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"net/url"
	"time"
)

type TallyDriver struct {
	httpClient     *http.Client
	organizationID string
	baseURL        string
}

var _ Driver = (*TallyDriver)(nil)

const (
	tallyOrganizationsPath = "/organizations"
	tallyUsersSegment      = "users"
	tallyInvitesSegment    = "invites"
	tallyCurrentUserPath   = "/users/me"
)

// ErrTallyUnauthorized reports that Tally rejected the presented API key.
// Tally's auth middleware runs before routing, so a 401 carries no route
// information: it always means the credential itself was refused.
var ErrTallyUnauthorized = errors.New("tally rejected the api key")

// TallyCurrentUser is the subset of GET /users/me the driver needs: the
// organization the API key belongs to, and the identities used to label
// the access source.
type TallyCurrentUser struct {
	FullName          string                  `json:"fullName"`
	Email             string                  `json:"email"`
	OrganizationID    string                  `json:"organizationId"`
	OrganizationOwner *TallyOrganizationOwner `json:"organizationOwner"`
}

type TallyOrganizationOwner struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

// GetTallyCurrentUser fetches the profile bound to the API key
// (GET /users/me). It is the only Tally endpoint that both accepts API-key
// auth and identifies the key: /me and /organizations/{id} are session-only
// and return 401 for every API key, valid or not. The create-connector
// resolver uses it to validate the key and derive the organization ID; the
// name resolver labels the source from it.
func GetTallyCurrentUser(ctx context.Context, httpClient *http.Client, baseURL string) (*TallyCurrentUser, error) {
	endpoint, err := url.JoinPath(baseURL, tallyCurrentUserPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build tally current user URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create tally current user request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute tally current user request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden {
		return nil, ErrTallyUnauthorized
	}

	// nameStatusError keeps the name worker's terminal-versus-retryable
	// classification: permanent client errors (400, 404) must not re-enqueue
	// the source forever, while 5xx stays retryable.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, nameStatusError("tally current user", httpResp.StatusCode)
	}

	var user TallyCurrentUser
	if err := json.NewDecoder(httpResp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("cannot decode tally current user response: %w", err)
	}

	return &user, nil
}

type tallyUser struct {
	ID                  string    `json:"id"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	FullName            string    `json:"fullName"`
	Email               string    `json:"email"`
	IsDeleted           bool      `json:"isDeleted"`
	HasTwoFactorEnabled bool      `json:"hasTwoFactorEnabled"`
	CreatedAt           time.Time `json:"createdAt"`
}

type tallyInvite struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewTallyDriver(httpClient *http.Client, organizationID, baseURL string) *TallyDriver {
	return &TallyDriver{
		httpClient:     httpClient,
		organizationID: organizationID,
		baseURL:        baseURL,
	}
}

func (d *TallyDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	records, err := d.listUsers(ctx)
	if err != nil {
		return nil, err
	}

	inviteRecords, err := d.listInvites(ctx)
	if err != nil {
		return nil, err
	}

	records = append(records, inviteRecords...)

	return records, nil
}

func (d *TallyDriver) listUsers(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, tallyOrganizationsPath, url.PathEscape(d.organizationID), tallyUsersSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build tally users URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create tally users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute tally users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"cannot fetch tally users: unexpected status %d",
			httpResp.StatusCode,
		)
	}

	var users []tallyUser
	if err := json.NewDecoder(httpResp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("cannot decode tally users response: %w", err)
	}

	var records []AccountRecord

	for _, u := range users {
		mfaStatus := coredata.MFAStatusDisabled
		if u.HasTwoFactorEnabled {
			mfaStatus = coredata.MFAStatusEnabled
		}

		record := AccountRecord{
			Email:       u.Email,
			FullName:    u.FullName,
			Active:      new(!u.IsDeleted),
			ExternalID:  u.ID,
			MFAStatus:   mfaStatus,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			CreatedAt:   new(u.CreatedAt),
		}

		if record.Email != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

func (d *TallyDriver) listInvites(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, tallyOrganizationsPath, url.PathEscape(d.organizationID), tallyInvitesSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build tally invites URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create tally invites request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute tally invites request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"cannot fetch tally invites: unexpected status %d",
			httpResp.StatusCode,
		)
	}

	var invites []tallyInvite
	if err := json.NewDecoder(httpResp.Body).Decode(&invites); err != nil {
		return nil, fmt.Errorf("cannot decode tally invites response: %w", err)
	}

	var records []AccountRecord

	for _, inv := range invites {
		record := AccountRecord{
			Email:       inv.Email,
			Active:      new(false),
			ExternalID:  inv.ID,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			Roles:       tallyRoles(),
		}

		if record.Email != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

func tallyRoles() []string {
	return []string{"Invited"}
}

// tallyNameResolver labels the source from the current-user profile. Tally
// has no organization-name API (organizations are unnamed in its model), so
// the closest stable label is the organization owner's identity, falling
// back to the key's user.
type tallyNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

func NewTallyNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &tallyNameResolver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *tallyNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	user, err := GetTallyCurrentUser(ctx, r.httpClient, r.baseURL)
	if err != nil {
		if errors.Is(err, ErrTallyUnauthorized) {
			return "", fmt.Errorf("cannot fetch tally current user: unauthorized: %w", ErrTerminalNameResolution)
		}

		return "", err
	}

	// Prefer the owner's identity entirely (name, then email) before
	// falling back to the key's user, so a non-owner key still labels
	// the source with the owner.
	if user.OrganizationOwner != nil {
		if user.OrganizationOwner.FullName != "" {
			return user.OrganizationOwner.FullName, nil
		}

		if user.OrganizationOwner.Email != "" {
			return user.OrganizationOwner.Email, nil
		}
	}

	if user.FullName != "" {
		return user.FullName, nil
	}

	return user.Email, nil
}

func tallySource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := tallySourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			NewTallyNameResolver(credential.Client, opened.Endpoints.APIBase),
			nil,
		), nil
	})
}

func tallySourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	_ *log.Logger,
	ep provider.Endpoints,
) (Driver, error) {
	s, err := coredata.ConnectorSettings[coredata.TallyConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read tally connector settings: %w", err)
	}

	if s.OrganizationID == "" {
		return nil, fmt.Errorf("cannot create tally driver: organization_id is required")
	}

	return NewTallyDriver(c, s.OrganizationID, ep.APIBase), nil
}
