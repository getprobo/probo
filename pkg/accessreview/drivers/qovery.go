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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const qoveryAPIBaseURL = "https://api.qovery.com"

type QoveryDriver struct {
	httpClient     *http.Client
	organizationID string
}

var _ Driver = (*QoveryDriver)(nil)

type qoveryMembersResponse struct {
	Results []qoveryMember `json:"results"`
}

type qoveryMember struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Nickname       string `json:"nickname"`
	Email          string `json:"email"`
	LastActivityAt string `json:"last_activity_at"`
	CreatedAt      string `json:"created_at"`
	Role           string `json:"role"`
}

func NewQoveryDriver(httpClient *http.Client, organizationID string) *QoveryDriver {
	return &QoveryDriver{
		httpClient:     httpClient,
		organizationID: organizationID,
	}
}

func (d *QoveryDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(
		qoveryAPIBaseURL,
		"organization",
		url.PathEscape(d.organizationID),
		"member",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build qovery members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create qovery members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute qovery members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch qovery members: unexpected status %d", httpResp.StatusCode)
	}

	var resp qoveryMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode qovery members response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Results))
	for _, member := range resp.Results {
		if member.Email == "" {
			continue
		}

		record := AccountRecord{
			Email:       member.Email,
			FullName:    qoveryFullName(member),
			Roles:       qoveryRoles(member.Role),
			IsAdmin:     qoveryIsAdmin(member.Role),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  member.ID,
		}

		if member.LastActivityAt != "" {
			if t, err := time.Parse(time.RFC3339, member.LastActivityAt); err == nil {
				record.LastLogin = &t
			}
		}

		if member.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, member.CreatedAt); err == nil {
				record.CreatedAt = &t
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func qoveryFullName(member qoveryMember) string {
	if member.Name != "" {
		return member.Name
	}

	if member.Nickname != "" {
		return member.Nickname
	}

	return member.Email
}

func qoveryRoles(role string) []string {
	if role == "" {
		return []string{}
	}

	switch strings.ToUpper(role) {
	case "OWNER":
		return []string{"Owner"}
	case "ADMIN":
		return []string{"Admin"}
	case "DEVELOPER":
		return []string{"Developer"}
	case "VIEWER":
		return []string{"Viewer"}
	default:
		return []string{role}
	}
}

func qoveryIsAdmin(role string) bool {
	switch strings.ToUpper(role) {
	case "OWNER", "ADMIN":
		return true
	default:
		return false
	}
}
