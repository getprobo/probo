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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// railwayGraphQLEndpoint is Railway's GraphQL API (note the .com TLD — the
// legacy backboard.railway.app host is deprecated).
const railwayGraphQLEndpoint = "https://backboard.railway.com/graphql/v2"

// railwayMembersQuery fetches the authenticated account and the members of all
// its workspaces. members/workspaces are plain lists (not Relay connections),
// so a single request returns everyone; the same user id recurs across
// workspaces and is deduplicated by the driver.
const railwayMembersQuery = `query { me { id name email workspaces { id name members { id email name role twoFactorAuthEnabled } } } }`

// RailwayDriver lists the members of every workspace an account token can see,
// via Railway's GraphQL API. The token flows in the Authorization header as a
// Bearer credential set by the connection transport.
type RailwayDriver struct {
	httpClient *http.Client
}

var _ Driver = (*RailwayDriver)(nil)

func NewRailwayDriver(httpClient *http.Client) *RailwayDriver {
	return &RailwayDriver{httpClient: httpClient}
}

type railwayMember struct {
	ID                   string `json:"id"`
	Email                string `json:"email"`
	Name                 string `json:"name"`
	Role                 string `json:"role"`
	TwoFactorAuthEnabled *bool  `json:"twoFactorAuthEnabled"`
}

type railwayWorkspace struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Members []railwayMember `json:"members"`
}

type railwayMe struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Email      string             `json:"email"`
	Workspaces []railwayWorkspace `json:"workspaces"`
}

type railwayMeResponse struct {
	Data struct {
		Me *railwayMe `json:"me"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// railwayAggregate accumulates a single human's appearances across workspaces:
// roles are unioned, IsAdmin is true if any workspace lists them as ADMIN, and
// MFA is enabled if any workspace reports it (with a separate signal flag so an
// all-null result stays Unknown rather than Disabled).
type railwayAggregate struct {
	record     AccountRecord
	roles      map[string]struct{}
	isAdmin    bool
	mfaEnabled bool
	mfaSignal  bool
}

func (d *RailwayDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	me, err := d.queryMe(ctx)
	if err != nil {
		return nil, err
	}

	return railwayRecords(me), nil
}

// railwayRecords aggregates the members of every workspace into one record per
// human, deduplicated by member id: roles are unioned, IsAdmin is true if any
// workspace lists them as ADMIN, and MFA is enabled if any workspace reports it
// (an all-null twoFactorAuthEnabled stays Unknown).
func railwayRecords(me *railwayMe) []AccountRecord {
	order := make([]string, 0)
	byKey := make(map[string]*railwayAggregate)

	for _, ws := range me.Workspaces {
		for _, m := range ws.Members {
			email := strings.TrimSpace(m.Email)
			if email == "" {
				continue
			}

			id := strings.TrimSpace(m.ID)

			key := id
			if key == "" {
				key = email
			}

			agg, ok := byKey[key]
			if !ok {
				agg = &railwayAggregate{
					record: AccountRecord{
						Email:       email,
						FullName:    railwayFullName(m, email),
						AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
						AccountType: coredata.AccessReviewEntryAccountTypeUser,
						ExternalID:  id,
					},
					roles: make(map[string]struct{}),
				}
				byKey[key] = agg
				order = append(order, key)
			}

			for _, role := range railwayRoles(m.Role) {
				agg.roles[role] = struct{}{}
			}

			if strings.EqualFold(strings.TrimSpace(m.Role), "ADMIN") {
				agg.isAdmin = true
			}

			if m.TwoFactorAuthEnabled != nil {
				agg.mfaSignal = true
				if *m.TwoFactorAuthEnabled {
					agg.mfaEnabled = true
				}
			}
		}
	}

	records := make([]AccountRecord, 0, len(order))

	for _, key := range order {
		agg := byKey[key]

		roles := make([]string, 0, len(agg.roles))
		for role := range agg.roles {
			roles = append(roles, role)
		}

		sort.Strings(roles)

		agg.record.Roles = roles
		agg.record.IsAdmin = agg.isAdmin
		agg.record.MFAStatus = railwayMFAStatus(agg.mfaSignal, agg.mfaEnabled)

		records = append(records, agg.record)
	}

	// Railway does not guarantee a stable member ordering across calls, so sort
	// by email (external id as tiebreak) for deterministic output, mirroring the
	// per-record role sort above.
	sort.Slice(records, func(i, j int) bool {
		if records[i].Email != records[j].Email {
			return records[i].Email < records[j].Email
		}

		return records[i].ExternalID < records[j].ExternalID
	})

	return records
}

func (d *RailwayDriver) queryMe(ctx context.Context) (*railwayMe, error) {
	httpResp, err := railwayPost(ctx, d.httpClient, "members", railwayMembersQuery)
	if err != nil {
		return nil, err
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch railway members: unexpected status %d", httpResp.StatusCode)
	}

	var resp railwayMeResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode railway members response: %w", err)
	}

	// Railway returns HTTP 200 with a populated errors array (and data.me null)
	// for a rejected token, so the status alone cannot be trusted. Provider
	// messages may carry identifiers — never embed them in the returned error.
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("cannot fetch railway members: graphql error")
	}

	if resp.Data.Me == nil {
		return nil, fmt.Errorf("cannot fetch railway members: no authenticated account")
	}

	return resp.Data.Me, nil
}

// railwayPost issues a GraphQL POST carrying query to Railway's endpoint,
// setting the Content-Type and Accept headers; the Bearer credential is attached
// by the connection transport. The caller owns status handling and must close
// the returned response body. label names the request in wrapped errors.
func railwayPost(ctx context.Context, httpClient *http.Client, label, query string) (*http.Response, error) {
	body := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal railway %s query: %w", label, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, railwayGraphQLEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("cannot create railway %s request: %w", label, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute railway %s request: %w", label, err)
	}

	return httpResp, nil
}

func railwayFullName(m railwayMember, fallback string) string {
	if name := strings.TrimSpace(m.Name); name != "" {
		return name
	}

	return fallback
}

// railwayRoles maps Railway's TeamRole enum to a display label. The enum is
// ADMIN/MEMBER/VIEWER; an unknown future value is passed through verbatim and
// no role yields an empty slice.
func railwayRoles(role string) []string {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "ADMIN":
		return []string{"Admin"}
	case "MEMBER":
		return []string{"Member"}
	case "VIEWER":
		return []string{"Viewer"}
	default:
		if r := strings.TrimSpace(role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}

func railwayMFAStatus(hasSignal, enabled bool) coredata.MFAStatus {
	if !hasSignal {
		return coredata.MFAStatusUnknown
	}

	if enabled {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}
