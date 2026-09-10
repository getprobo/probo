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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	// twingateUsersPageSize is the page size requested from the users
	// connection, and it is the largest Twingate's cursor pagination is
	// documented to serve.
	//
	// Twingate rate-limits reads to 60 a minute per account.
	twingateUsersPageSize = 100

	// twingatePageInterval keeps a long roster inside that budget: at one page
	// a second the driver cannot exceed 60 reads in a minute however many pages
	// it needs. It is waited only BETWEEN pages, so the common single-page
	// network pays nothing.
	//
	// The ceiling this leaves is the access-review service's own
	// sourceFetchTimeout, two minutes for one ListAccounts call. Spacing reads
	// to stay legal means a network of roughly ten thousand users is the most
	// that fits, after which the sync fails on the deadline rather than on a
	// 429. That is a higher ceiling than pacing-free requests reach, since
	// those exhaust the minute's budget at around six thousand users, but it is
	// still a ceiling: lifting it means giving a rate-limited driver a longer
	// budget than the shared default, which is a change to the review engine
	// rather than to this connector.
	twingatePageInterval = time.Second

	twingateStateActive   = "ACTIVE"
	twingateStatePending  = "PENDING"
	twingateStateDisabled = "DISABLED"

	// The seven values of Twingate's UserRole enum, confirmed by introspecting
	// a live network. An unrecognised eighth is preserved verbatim rather than
	// dropped, but naming these keeps a reviewer from reading SCREAMING_CASE.
	twingateRoleAdmin          = "ADMIN"
	twingateRoleDevops         = "DEVOPS"
	twingateRoleSupport        = "SUPPORT"
	twingateRoleHelpdesk       = "HELPDESK"
	twingateRoleAccessReviewer = "ACCESS_REVIEWER"
	twingateRoleBilling        = "BILLING"
	twingateRoleMember         = "MEMBER"

	twingateTypeSynced = "SYNCED"

	twingateUsersQuery = `query($after: String, $first: Int!) {
  users(after: $after, first: $first) {
    pageInfo { hasNextPage endCursor }
    edges { node { id email firstName lastName role isAdmin state type createdAt } }
  }
}`
)

// twingateNetworkPattern is the shape of a Twingate network name: it is a DNS
// label in <network>.twingate.com, so it is what a hostname label may hold and
// no longer. Validating it here is what keeps a settings value out of the
// host position of a URL.
var twingateNetworkPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

// TwingateNetwork canonicalises a customer-supplied network name to the label
// that appears in the host. Trimming and lowercasing happen here rather than at
// each use, so one network cannot be stored as "acme", "ACME" and "  acme  "
// and then read as three different sources by a reviewer.
func TwingateNetwork(network string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(network))
	if !twingateNetworkPattern.MatchString(name) {
		return "", fmt.Errorf("invalid twingate network name")
	}

	return name, nil
}

// TwingateEndpoint builds the GraphQL endpoint of one Twingate network. It is
// the single place the host is composed, so the driver, the name resolver and
// the connection probe cannot drift onto different networks for one connector.
func TwingateEndpoint(network string) (string, error) {
	name, err := TwingateNetwork(network)
	if err != nil {
		return "", err
	}

	// Built through net/url rather than concatenated: the label is the one part
	// of this URL a customer supplies, so it goes into the Host field instead
	// of into a string that happens to look like a URL.
	endpoint := url.URL{
		Scheme: "https",
		Host:   name + ".twingate.com",
		Path:   "/api/graphql/",
	}

	return endpoint.String(), nil
}

// TwingateDriver lists the users of one Twingate network. Twingate gives every
// tenant its own host, so the network the API token belongs to is already
// fixed by the endpoint and there is nothing to pick.
//
// Service accounts are a separate type in Twingate's schema and are not
// returned by the users connection, so every record here is a person.
type TwingateDriver struct {
	httpClient *http.Client
	endpoint   string
}

var _ Driver = (*TwingateDriver)(nil)

// NewTwingateDriver builds a driver against endpoint, the network's GraphQL
// endpoint (https://<network>.twingate.com/api/graphql/). It is the only
// endpoint the driver calls, so there is no path to join onto it.
func NewTwingateDriver(httpClient *http.Client, endpoint string) *TwingateDriver {
	client := *httpClient
	client.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &TwingateDriver{httpClient: &client, endpoint: endpoint}
}

type twingateUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
	IsAdmin   bool   `json:"isAdmin"`
	State     string `json:"state"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
}

type twingateUsersResponse struct {
	Data struct {
		Users struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Edges []struct {
				Node twingateUser `json:"node"`
			} `json:"edges"`
		} `json:"users"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (d *TwingateDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records       []AccountRecord
		after         string
		lastRequestAt time.Time
	)

	for range maxPaginationPages {
		// Spaced by when the previous request STARTED, not by when it
		// finished, so the time Twingate spent answering counts toward the
		// interval instead of being added to it. A sync that waits a flat
		// second on top of every round trip runs out of the source deadline
		// far sooner than the rate limit requires.
		if !lastRequestAt.IsZero() {
			if wait := twingatePageInterval - time.Since(lastRequestAt); wait > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(wait):
				}
			}
		}

		lastRequestAt = time.Now()

		page, err := d.queryUsers(ctx, after)
		if err != nil {
			return nil, err
		}

		for _, edge := range page.Data.Users.Edges {
			u := edge.Node

			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    twingateFullName(u, email),
				Roles:       twingateRoles(u),
				IsAdmin:     new(u.IsAdmin),
				Active:      twingateActive(u),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  twingateAuthMethod(u),
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(u.CreatedAt),
				ExternalID:  strings.TrimSpace(u.ID),
			})
		}

		if !page.Data.Users.PageInfo.HasNextPage {
			return records, nil
		}

		// Twingate says there is more but will not say where to resume. Ending
		// here would hand back a short roster with no error, and a member
		// missing from a campaign is reviewed by nobody.
		if page.Data.Users.PageInfo.EndCursor == "" {
			return nil, fmt.Errorf("cannot list all twingate users: next page has no cursor")
		}

		after = page.Data.Users.PageInfo.EndCursor
	}

	return nil, fmt.Errorf("cannot list all twingate users: %w", ErrPaginationLimitReached)
}

func (d *TwingateDriver) queryUsers(ctx context.Context, after string) (*twingateUsersResponse, error) {
	variables := map[string]any{"first": twingateUsersPageSize}
	if after != "" {
		variables["after"] = after
	}

	payload, err := json.Marshal(map[string]any{
		"query":     twingateUsersQuery,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal twingate users request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("cannot create twingate users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute twingate users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch twingate users: unexpected status %d", httpResp.StatusCode)
	}

	var resp twingateUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode twingate users response: %w", err)
	}

	// Twingate answers a rejected query with 200 and an errors array. Provider
	// messages may carry tenant identifiers — never embed them.
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("cannot fetch twingate users: graphql error")
	}

	return &resp, nil
}

func twingateFullName(u twingateUser, fallback string) string {
	name := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	if name != "" {
		return name
	}

	return fallback
}

// twingateActive reads Twingate's own lifecycle state. PENDING is an invited
// user who has not joined yet and DISABLED is a revoked one; both are reported
// inactive, and an unrecognised future state reports no signal rather than
// guessing that it means active.
func twingateActive(u twingateUser) *bool {
	switch u.State {
	case twingateStateActive:
		return new(true)
	case twingateStatePending, twingateStateDisabled:
		return new(false)
	default:
		return nil
	}
}

// twingateAuthMethod reads how the account came to exist. SYNCED means an
// identity provider provisioned it, which is the SSO case. MANUAL means an
// admin typed it in, which says nothing about how that person then signs in —
// they may still use the network's SSO — so it reports no signal rather than
// claiming a password.
func twingateAuthMethod(u twingateUser) coredata.AccessReviewEntryAuthMethod {
	if u.Type == twingateTypeSynced {
		return coredata.AccessReviewEntryAuthMethodSSO
	}

	return coredata.AccessReviewEntryAuthMethodUnknown
}

// twingateRoles maps the user's network role to a display label. isAdmin is
// reported separately on the record rather than folded in here: Twingate sets
// it for ADMIN and DEVOPS alike, so it answers "can this person change the
// network" while role answers "as what".
func twingateRoles(u twingateUser) []string {
	switch u.Role {
	case twingateRoleAdmin:
		return []string{"Admin"}
	case twingateRoleDevops:
		return []string{"DevOps"}
	case twingateRoleSupport:
		return []string{"Support"}
	case twingateRoleHelpdesk:
		return []string{"Helpdesk"}
	case twingateRoleAccessReviewer:
		return []string{"Access Reviewer"}
	case twingateRoleBilling:
		return []string{"Billing"}
	case twingateRoleMember:
		return []string{"Member"}
	default:
		if role := strings.TrimSpace(u.Role); role != "" {
			return []string{role}
		}

		return []string{}
	}
}

// twingateNameResolver names the source after the customer's Twingate network.
// The network is the tenant's own subdomain and is already in connector
// settings, so no request is needed — and none would be better: Twingate's
// schema exposes no account-name field.
type twingateNameResolver struct {
	network string
}

func NewTwingateNameResolver(network string) NameResolver {
	return &twingateNameResolver{network: network}
}

func (r *twingateNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return r.network, nil
}
