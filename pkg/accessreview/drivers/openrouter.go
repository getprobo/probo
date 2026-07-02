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
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	openRouterMembersEndpoint = "https://openrouter.ai/api/v1/organization/members"
	// openRouterPageSize is the maximum page size GET /organization/members
	// accepts (limit must be between 1 and 100).
	openRouterPageSize = 100
)

// OpenRouterDriver lists the members of a single OpenRouter organization. The
// management (provisioning) API key is bound to one organization, so GET
// /api/v1/organization/members returns every member of that organization
// with no tenant selector.
type OpenRouterDriver struct {
	httpClient *http.Client
}

var _ Driver = (*OpenRouterDriver)(nil)

type openRouterMember struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	// Role is OpenRouter's organization role enum: "org:admin" or
	// "org:member".
	Role string `json:"role"`
}

type openRouterMembersResponse struct {
	Data       []openRouterMember `json:"data"`
	TotalCount int                `json:"total_count"`
}

func NewOpenRouterDriver(httpClient *http.Client) *OpenRouterDriver {
	return &OpenRouterDriver{httpClient: httpClient}
}

func (d *OpenRouterDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	offset := 0

	for range maxPaginationPages {
		resp, err := d.fetchMembersPage(ctx, offset)
		if err != nil {
			return nil, err
		}

		for _, m := range resp.Data {
			email := strings.TrimSpace(m.Email)
			if email == "" {
				continue
			}

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    openRouterFullName(m, email),
				Roles:       openRouterRoles(m.Role),
				IsAdmin:     m.Role == "org:admin",
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strings.TrimSpace(m.ID),
			})
		}

		offset += len(resp.Data)
		if len(resp.Data) < openRouterPageSize || offset >= resp.TotalCount {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all openrouter members: %w", ErrPaginationLimitReached)
}

func (d *OpenRouterDriver) fetchMembersPage(ctx context.Context, offset int) (*openRouterMembersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openRouterMembersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create openrouter members request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(openRouterPageSize))
	q.Set("offset", strconv.Itoa(offset))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute openrouter members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch openrouter members: unexpected status %d", httpResp.StatusCode)
	}

	var resp openRouterMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode openrouter members response: %w", err)
	}

	return &resp, nil
}

func openRouterFullName(m openRouterMember, fallback string) string {
	first := ""
	if m.FirstName != nil {
		first = strings.TrimSpace(*m.FirstName)
	}

	last := ""
	if m.LastName != nil {
		last = strings.TrimSpace(*m.LastName)
	}

	full := strings.TrimSpace(first + " " + last)
	if full != "" {
		return full
	}

	return fallback
}

// openRouterRoles maps OpenRouter's organization role enum
// (org:admin / org:member) to a display label, preserving any unknown
// future role verbatim.
func openRouterRoles(role string) []string {
	switch role {
	case "org:admin":
		return []string{"Admin"}
	case "org:member":
		return []string{"Member"}
	default:
		if strings.TrimSpace(role) != "" {
			return []string{role}
		}

		return []string{}
	}
}
