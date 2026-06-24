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

	"go.probo.inc/probo/pkg/coredata"
)

const renderAPIBaseURL = "https://api.render.com/v1"

type RenderDriver struct {
	httpClient *http.Client
	ownerID    string
}

var _ Driver = (*RenderDriver)(nil)

// renderMember mirrors one element of the flat array returned by
// GET /v1/owners/{ownerId}/members. The endpoint takes no query parameters
// and is not paginated: it returns every workspace member (active and
// inactive) in a single response.
type renderMember struct {
	UserID     string `json:"userId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Status     string `json:"status"` // "active" | "inactive"
	Role       string `json:"role"`   // always uppercase
	MFAEnabled bool   `json:"mfaEnabled"`
}

func NewRenderDriver(httpClient *http.Client, ownerID string) *RenderDriver {
	return &RenderDriver{
		httpClient: httpClient,
		ownerID:    ownerID,
	}
}

func (d *RenderDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(
		renderAPIBaseURL,
		"owners",
		url.PathEscape(d.ownerID),
		"members",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build render members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create render members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute render members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch render members: unexpected status %d", httpResp.StatusCode)
	}

	var members []renderMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("cannot decode render members response: %w", err)
	}

	records := make([]AccountRecord, 0, len(members))
	for _, member := range members {
		if member.Email == "" {
			continue
		}

		records = append(records, AccountRecord{
			Email:       member.Email,
			FullName:    renderFullName(member),
			Roles:       renderRoles(member.Role),
			Active:      renderActive(member.Status),
			IsAdmin:     renderIsAdmin(member.Role),
			MFAStatus:   renderMFAStatus(member.MFAEnabled),
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  member.UserID,
		})
	}

	return records, nil
}

func renderFullName(member renderMember) string {
	if member.Name != "" {
		return member.Name
	}

	return member.Email
}

// renderRoles maps Render's uppercase role enum to a human-readable label.
// Render documents ADMIN, DEVELOPER, WORKSPACE_CONTRIBUTOR,
// WORKSPACE_BILLING, and WORKSPACE_VIEWER; unknown future roles fall through
// to the raw value.
func renderRoles(role string) []string {
	if role == "" {
		return []string{}
	}

	switch strings.ToUpper(role) {
	case "ADMIN":
		return []string{"Admin"}
	case "DEVELOPER":
		return []string{"Developer"}
	case "WORKSPACE_CONTRIBUTOR":
		return []string{"Contributor"}
	case "WORKSPACE_BILLING":
		return []string{"Billing"}
	case "WORKSPACE_VIEWER":
		return []string{"Viewer"}
	default:
		return []string{role}
	}
}

// renderIsAdmin reports whether a Render role grants workspace
// administration. Only ADMIN does (it also covers the workspace owner, who
// is reported with the ADMIN role).
func renderIsAdmin(role string) bool {
	return strings.EqualFold(role, "ADMIN")
}

func renderMFAStatus(enabled bool) coredata.MFAStatus {
	if enabled {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

// renderActive maps Render's documented status enum to the three-valued
// Active signal. Only the documented "active"/"inactive" values are an
// explicit signal; an empty or unrecognized status leaves Active nil
// (unknown) rather than fabricating a deactivated state, per the
// AccountRecord contract.
func renderActive(status string) *bool {
	switch strings.ToLower(status) {
	case "active":
		active := true
		return &active
	case "inactive":
		inactive := false
		return &inactive
	default:
		return nil
	}
}
