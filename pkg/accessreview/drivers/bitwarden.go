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

	"github.com/google/uuid"
	"go.probo.inc/probo/pkg/coredata"
)

// Each Bitwarden cloud runs its own API and identity hosts, and an
// organization API key is only known to the cloud it was created on.
const (
	bitwardenRegionUS = "US"
	bitwardenRegionEU = "EU"

	bitwardenUSAPIHost      = "api.bitwarden.com"
	bitwardenUSIdentityHost = "identity.bitwarden.com"
	bitwardenEUAPIHost      = "api.bitwarden.eu"
	bitwardenEUIdentityHost = "identity.bitwarden.eu"

	bitwardenTokenPath = "/connect/token"
)

// Bitwarden serializes its enums as integers.
const (
	bitwardenStatusRevoked   = -1
	bitwardenStatusInvited   = 0
	bitwardenStatusAccepted  = 1
	bitwardenStatusConfirmed = 2
	bitwardenStatusStaged    = 3

	bitwardenTypeOwner  = 0
	bitwardenTypeAdmin  = 1
	bitwardenTypeUser   = 2
	bitwardenTypeCustom = 4
)

// BitwardenDriver lists the members of the organization an organization API
// key belongs to, through the Bitwarden Public API.
type BitwardenDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*BitwardenDriver)(nil)

type bitwardenMembersResponse struct {
	Data              []bitwardenMember `json:"data"`
	ContinuationToken string            `json:"continuationToken"`
}

type bitwardenMember struct {
	ID               string          `json:"id"`
	UserID           string          `json:"userId"`
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	TwoFactorEnabled bool            `json:"twoFactorEnabled"`
	Status           int             `json:"status"`
	Type             int             `json:"type"`
	Permissions      map[string]bool `json:"permissions"`
}

// NewBitwardenDriver builds a driver against baseURL, the API root of the
// connector's region (BitwardenAPIBaseURL).
func NewBitwardenDriver(httpClient *http.Client, baseURL string) *BitwardenDriver {
	client := *httpClient
	client.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &BitwardenDriver{
		httpClient: &client,
		baseURL:    baseURL,
	}
}

// IsBitwardenRegion reports whether region names a Bitwarden cloud, in the
// form the connector stores it.
func IsBitwardenRegion(region string) bool {
	_, _, ok := bitwardenHosts(region)

	return ok
}

// BitwardenOrganizationID returns the organization id in clientID when it is
// the client_id of an organization API key: "organization." and the id. A
// personal API key ("user.") mints tokens the Public API refuses.
func BitwardenOrganizationID(clientID string) (string, bool) {
	orgID, ok := strings.CutPrefix(clientID, "organization.")

	// uuid.Validate also takes the braced, URN and undashed forms.
	if !ok || len(orgID) != 36 || uuid.Validate(orgID) != nil {
		return "", false
	}

	// The account id is compared as a string, so keep one case.
	return strings.ToLower(orgID), true
}

// BitwardenAPIBaseURL returns the API root of region.
func BitwardenAPIBaseURL(region string) (string, error) {
	apiHost, _, ok := bitwardenHosts(region)
	if !ok {
		return "", fmt.Errorf("cannot resolve bitwarden api host: unknown region %q", region)
	}

	u := url.URL{Scheme: "https", Host: apiHost}

	return u.String(), nil
}

// BitwardenTokenURL returns the token endpoint of region.
func BitwardenTokenURL(region string) (string, error) {
	_, identityHost, ok := bitwardenHosts(region)
	if !ok {
		return "", fmt.Errorf("cannot resolve bitwarden identity host: unknown region %q", region)
	}

	u := url.URL{Scheme: "https", Host: identityHost, Path: bitwardenTokenPath}

	return u.String(), nil
}

func bitwardenHosts(region string) (apiHost string, identityHost string, ok bool) {
	switch region {
	case bitwardenRegionUS:
		return bitwardenUSAPIHost, bitwardenUSIdentityHost, true
	case bitwardenRegionEU:
		return bitwardenEUAPIHost, bitwardenEUIdentityHost, true
	default:
		return "", "", false
	}
}

func (d *BitwardenDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, "public", "members")
	if err != nil {
		return nil, fmt.Errorf("cannot build bitwarden members url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create bitwarden members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute bitwarden members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch bitwarden members: unexpected status %d", httpResp.StatusCode)
	}

	var resp bitwardenMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode bitwarden members response: %w", err)
	}

	// The endpoint answers in one page and ignores a continuationToken query,
	// so a token here would mean the roster is truncated.
	if resp.ContinuationToken != "" {
		return nil, fmt.Errorf("cannot list all bitwarden members: response is paginated")
	}

	records := make([]AccountRecord, 0, len(resp.Data))
	for _, m := range resp.Data {
		records = append(records, bitwardenMemberRecord(m))
	}

	return records, nil
}

func bitwardenMemberRecord(m bitwardenMember) AccountRecord {
	return AccountRecord{
		Email:    m.Email,
		FullName: m.Name,
		Roles:    bitwardenRoles(m.Type),
		Active:   bitwardenMemberActive(m.Status),
		IsAdmin:  bitwardenMemberIsAdmin(m),
		// An invitee has no Bitwarden account yet, so twoFactorEnabled is false
		// without meaning anything.
		MFAStatus:   bitwardenMFAStatus(m),
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		// The member id, not userId: userId is null until the invitation is
		// accepted, and the entry must survive that.
		ExternalID: m.ID,
	}
}

func bitwardenRoles(memberType int) []string {
	switch memberType {
	case bitwardenTypeOwner:
		return []string{"Owner"}
	case bitwardenTypeAdmin:
		return []string{"Admin"}
	case bitwardenTypeUser:
		return []string{"User"}
	case bitwardenTypeCustom:
		return []string{"Custom"}
	default:
		return []string{}
	}
}

// bitwardenMemberActive treats only a confirmed member as having access: an
// invited, accepted or staged one cannot open the organization's vault yet.
func bitwardenMemberActive(status int) *bool {
	switch status {
	case bitwardenStatusConfirmed:
		return new(true)
	case bitwardenStatusRevoked, bitwardenStatusInvited, bitwardenStatusAccepted, bitwardenStatusStaged:
		return new(false)
	default:
		return nil
	}
}

// bitwardenMemberIsAdmin counts a Custom member as an admin when it holds a
// permission that opens an Admin Console tab, which is every one but
// createNewCollections.
func bitwardenMemberIsAdmin(m bitwardenMember) *bool {
	switch m.Type {
	case bitwardenTypeOwner, bitwardenTypeAdmin:
		return new(true)
	case bitwardenTypeUser:
		return new(false)
	case bitwardenTypeCustom:
		if m.Permissions == nil {
			return nil
		}

		for permission, granted := range m.Permissions {
			if granted && permission != "createNewCollections" {
				return new(true)
			}
		}

		return new(false)
	default:
		return nil
	}
}

func bitwardenMFAStatus(m bitwardenMember) coredata.MFAStatus {
	switch {
	case m.UserID == "":
		return coredata.MFAStatusUnknown
	case m.TwoFactorEnabled:
		return coredata.MFAStatusEnabled
	default:
		return coredata.MFAStatusDisabled
	}
}
