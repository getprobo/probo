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

	"bytes"
	"encoding/json"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"sort"
	"strings"
)

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
	endpoint   string
}

var _ Driver = (*RailwayDriver)(nil)

// NewRailwayDriver builds a driver against endpoint, Railway's GraphQL API
// (e.g. https://backboard.railway.com/graphql/v2 — note the .com TLD, the
// legacy backboard.railway.app host is deprecated). It is the only endpoint
// the driver calls, so there is no path to join onto it.
func NewRailwayDriver(httpClient *http.Client, endpoint string) *RailwayDriver {
	return &RailwayDriver{httpClient: httpClient, endpoint: endpoint}
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
		agg.record.IsAdmin = new(agg.isAdmin)
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
	httpResp, err := railwayPost(ctx, d.httpClient, d.endpoint, "members", railwayMembersQuery)
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

// railwayPost issues a GraphQL POST carrying query to endpoint, Railway's
// GraphQL API, setting the Content-Type and Accept headers; the Bearer
// credential is attached by the connection transport. The caller owns status
// handling and must close the returned response body. label names the request
// in wrapped errors.
func railwayPost(ctx context.Context, httpClient *http.Client, endpoint, label, query string) (*http.Response, error) {
	body := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal railway %s query: %w", label, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
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

// railwayNameResolver resolves the Railway workspace name via GraphQL, for the
// AccessReviewSource title. With a single workspace it uses that workspace's
// name; with several it falls back to the account holder's name, since the
// source spans all of the account's workspaces.
type railwayNameResolver struct {
	httpClient *http.Client
	endpoint   string
}

// NewRailwayNameResolver resolves the workspace name against endpoint,
// Railway's GraphQL API (e.g. https://backboard.railway.com/graphql/v2).
func NewRailwayNameResolver(httpClient *http.Client, endpoint string) NameResolver {
	return &railwayNameResolver{httpClient: httpClient, endpoint: endpoint}
}

func (r *railwayNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	httpResp, err := railwayPost(ctx, r.httpClient, r.endpoint, "account", `query { me { name workspaces { id name } } }`)
	if err != nil {
		return "", err
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx must not make the source-name worker retry forever
	// — keep the generic name. (Railway also signals auth failure with a 200 +
	// errors body, handled below.)
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Data struct {
			Me *struct {
				Name       string `json:"name"`
				Workspaces []struct {
					Name string `json:"name"`
				} `json:"workspaces"`
			} `json:"me"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode railway account response: %w", err)
	}

	if len(resp.Errors) > 0 || resp.Data.Me == nil {
		return "", nil
	}

	// A single workspace names the source directly; with several (or none) fall
	// back to the account holder's display name. Never the email — a terminal
	// empty result keeps the generic source name, which the worker tolerates.
	me := resp.Data.Me
	if len(me.Workspaces) == 1 {
		return me.Workspaces[0].Name, nil
	}

	return me.Name, nil
}

func railwaySource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		return capable(
			NewRailwayDriver(credential.Client, opened.Endpoints.APIBase),
			NewRailwayNameResolver(credential.Client, opened.Endpoints.APIBase),
			nil,
		), nil
	})
}
