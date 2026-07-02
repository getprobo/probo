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
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// tailscaleDefaultTailnet is the "-" shorthand Tailscale accepts in the
// tailnet path segment; it resolves to the access token's own tailnet, so
// the connector never needs to know the organization name up front.
const tailscaleDefaultTailnet = "-"

// TailscaleDriver fetches tailnet users from the Tailscale API via Bearer
// token-authenticated REST requests. It always targets the access token's
// default tailnet, so no tailnet identifier is required.
type TailscaleDriver struct {
	httpClient *http.Client
}

var _ Driver = (*TailscaleDriver)(nil)

type tailscaleUser struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"displayName"`
	LoginName          string `json:"loginName"`
	Created            string `json:"created"`
	Role               string `json:"role"`
	Status             string `json:"status"`
	LastSeen           string `json:"lastSeen"`
	CurrentlyConnected bool   `json:"currentlyConnected"`
}

type tailscaleUsersResponse struct {
	Users []tailscaleUser `json:"users"`
}

func NewTailscaleDriver(httpClient *http.Client) *TailscaleDriver {
	return &TailscaleDriver{
		httpClient: httpClient,
	}
}

func (d *TailscaleDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	users, err := d.fetchUsers(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(users))

	for _, u := range users {
		email := u.LoginName
		if email == "" {
			continue
		}

		role := strings.TrimSpace(u.Role)

		roles := []string{}
		if role != "" {
			roles = []string{role}
		}

		record := AccountRecord{
			Email:      email,
			FullName:   u.DisplayName,
			Roles:      roles,
			Active:     tailscaleUserActive(u.Status),
			IsAdmin:    tailscaleUserIsAdmin(u.Role),
			ExternalID: u.ID,
			MFAStatus:  coredata.MFAStatusUnknown,
			// Tailscale has no local credentials; it always delegates
			// authentication to an upstream identity provider, so every
			// account is SSO regardless of which IdP backs the tailnet.
			AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}

		if u.Created != "" {
			if t, err := time.Parse(time.RFC3339, u.Created); err == nil {
				record.CreatedAt = &t
			}
		}

		if u.LastSeen != "" {
			if t, err := time.Parse(time.RFC3339, u.LastSeen); err == nil {
				record.LastLogin = &t
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func (d *TailscaleDriver) fetchUsers(ctx context.Context) ([]tailscaleUser, error) {
	endpoint, err := url.JoinPath(
		"https://api.tailscale.com",
		"api",
		"v2",
		"tailnet",
		tailscaleDefaultTailnet,
		"users",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build tailscale users URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create tailscale users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute tailscale users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch tailscale users: unexpected status %d", httpResp.StatusCode)
	}

	var resp tailscaleUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode tailscale users response: %w", err)
	}

	return resp.Users, nil
}

func tailscaleUserActive(status string) *bool {
	switch strings.ToLower(status) {
	case "active", "idle":
		return new(true)
	case "suspended":
		return new(false)
	default:
		return nil
	}
}

func tailscaleUserIsAdmin(role string) bool {
	switch role {
	case "owner", "admin", "it-admin", "network-admin", "billing-admin":
		return true
	default:
		return false
	}
}

// tailscaleNameResolver derives the tailnet name from the email domain shared
// by the tailnet's users. Tailscale exposes no API endpoint that returns the
// tailnet/organization name directly, and the connector targets the "-"
// default tailnet so the identifier is never captured up front. For tailnets
// backed by a custom domain the user login domain matches the tailnet ID
// exactly (e.g. "example.com"); for shared-domain tailnets it degrades to the
// provider domain, which is still a useful label.
type tailscaleNameResolver struct {
	httpClient *http.Client
}

var _ NameResolver = (*tailscaleNameResolver)(nil)

func NewTailscaleNameResolver(httpClient *http.Client) NameResolver {
	return &tailscaleNameResolver{httpClient: httpClient}
}

func (r *tailscaleNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	driver := &TailscaleDriver{httpClient: r.httpClient}

	users, err := driver.fetchUsers(ctx)
	if err != nil {
		return "", err
	}

	return tailscaleTailnetName(users), nil
}

// tailscaleTailnetName returns the most common email domain among the tailnet
// users, preserving first-seen order to break ties deterministically.
func tailscaleTailnetName(users []tailscaleUser) string {
	counts := make(map[string]int, len(users))
	order := make([]string, 0, len(users))

	for _, u := range users {
		at := strings.LastIndex(u.LoginName, "@")
		if at < 0 || at == len(u.LoginName)-1 {
			continue
		}

		domain := strings.ToLower(u.LoginName[at+1:])
		if _, seen := counts[domain]; !seen {
			order = append(order, domain)
		}

		counts[domain]++
	}

	best := ""
	bestCount := 0

	for _, domain := range order {
		if counts[domain] > bestCount {
			best = domain
			bestCount = counts[domain]
		}
	}

	return best
}
