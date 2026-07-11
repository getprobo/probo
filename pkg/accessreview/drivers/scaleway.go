// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	scalewayAPIHost   = "api.scaleway.com"
	scalewayUsersPath = "/iam/v1alpha1/users"
	scalewayPageSize  = 100
)

// ScalewayDriver lists the IAM users of a single Scaleway Organization. The
// secret key (sent in the X-Auth-Token header by the connection transport) is
// scoped to one Organization, but GET /iam/v1alpha1/users requires the
// organization_id explicitly, so it is captured up front as a connector
// setting rather than discovered.
type ScalewayDriver struct {
	httpClient     *http.Client
	organizationID string
}

var _ Driver = (*ScalewayDriver)(nil)

type scalewayUsersResponse struct {
	Users      []scalewayUser `json:"users"`
	TotalCount uint32         `json:"total_count"`
}

type scalewayUser struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Username    string  `json:"username"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	CreatedAt   string  `json:"created_at"`
	LastLoginAt *string `json:"last_login_at"`
	// Type is the org-level user type ("owner" | "member"). Fine-grained IAM
	// roles live on separate policy/group endpoints and are out of scope.
	Type   string `json:"type"`
	Status string `json:"status"`
	// MFA is always present; TwoFactorEnabled is the newer pointer mirror of
	// the same state and is preferred when set (see scalewayMFAStatus).
	MFA              bool  `json:"mfa"`
	TwoFactorEnabled *bool `json:"two_factor_enabled"`
	Locked           bool  `json:"locked"`
}

func NewScalewayDriver(httpClient *http.Client, organizationID string) *ScalewayDriver {
	return &ScalewayDriver{
		httpClient:     httpClient,
		organizationID: organizationID,
	}
}

func (d *ScalewayDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	fetched := 0
	page := 1

	for range maxPaginationPages {
		resp, err := d.fetchPage(ctx, page)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Users {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			record := AccountRecord{
				Email:       email,
				FullName:    scalewayFullName(u, email),
				Roles:       scalewayRoles(u.Type),
				Active:      scalewayActive(u.Status, u.Locked),
				IsAdmin:     scalewayIsAdmin(u.Type),
				MFAStatus:   scalewayMFAStatus(u),
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(u.CreatedAt),
				ExternalID:  strings.TrimSpace(u.ID),
			}

			if u.LastLoginAt != nil {
				record.LastLogin = parseRFC3339Ptr(*u.LastLoginAt)
			}

			records = append(records, record)
		}

		fetched += len(resp.Users)
		if len(resp.Users) < scalewayPageSize || uint32(fetched) >= resp.TotalCount {
			return records, nil
		}

		page++
	}

	return nil, fmt.Errorf("cannot list all scaleway accounts: %w", ErrPaginationLimitReached)
}

func (d *ScalewayDriver) fetchPage(ctx context.Context, page int) (*scalewayUsersResponse, error) {
	q := url.Values{}
	q.Set("organization_id", d.organizationID)
	q.Set("order_by", "created_at_asc")
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(scalewayPageSize))

	endpoint := url.URL{
		Scheme:   "https",
		Host:     scalewayAPIHost,
		Path:     scalewayUsersPath,
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create scaleway users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute scaleway users request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch scaleway users: unexpected status %d", httpResp.StatusCode)
	}

	var resp scalewayUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode scaleway users response: %w", err)
	}

	return &resp, nil
}

func scalewayFullName(u scalewayUser, fallback string) string {
	if name := strings.TrimSpace(u.FirstName + " " + u.LastName); name != "" {
		return name
	}

	if username := strings.TrimSpace(u.Username); username != "" {
		return username
	}

	return fallback
}

// scalewayRoles maps the Scaleway org-level user type to a display label. The
// users endpoint exposes only owner/member; an unknown future value is passed
// through verbatim and no type yields an empty slice.
func scalewayRoles(userType string) []string {
	switch strings.ToLower(strings.TrimSpace(userType)) {
	case "owner":
		return []string{"Owner"}
	case "member":
		return []string{"Member"}
	default:
		if t := strings.TrimSpace(userType); t != "" {
			return []string{t}
		}

		return []string{}
	}
}

// scalewayIsAdmin reports whether a Scaleway user type grants administrative
// access. Only the organization owner is an administrator; members are not.
func scalewayIsAdmin(userType string) bool {
	return strings.EqualFold(strings.TrimSpace(userType), "owner")
}

// scalewayActive maps the Scaleway user status to the three-valued Active
// signal. A locked account is always inactive; otherwise only the documented
// "activated"/"invitation_pending" values are an explicit signal and any other
// or missing status leaves Active nil (no signal). The literal live value is
// "activated", not "active", so the shared activeFromStatus helper is not used.
func scalewayActive(status string, locked bool) *bool {
	if locked {
		inactive := false

		return &inactive
	}

	switch strings.ToLower(strings.TrimSpace(status)) {
	case "activated":
		active := true

		return &active
	case "invitation_pending":
		inactive := false

		return &inactive
	default:
		return nil
	}
}

// scalewayMFAStatus reads the two-factor state, preferring the newer
// two_factor_enabled pointer when present and otherwise the always-present mfa
// boolean.
func scalewayMFAStatus(u scalewayUser) coredata.MFAStatus {
	enabled := u.MFA
	if u.TwoFactorEnabled != nil {
		enabled = *u.TwoFactorEnabled
	}

	if enabled {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}
