// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type SendGridDriver struct {
	httpClient *http.Client
}

var _ Driver = (*SendGridDriver)(nil)

type sendGridTeammate struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type"`
	IsAdmin   bool   `json:"is_admin"`
}

type sendGridTeammatesResponse struct {
	Result  []sendGridTeammate `json:"result"`
	Results []sendGridTeammate `json:"results"`
}

const (
	sendGridTeammatesEndpoint  = "https://api.sendgrid.com/v3/teammates"
	sendGridTeammatesPageLimit = 500
)

func NewSendGridDriver(httpClient *http.Client) *SendGridDriver {
	return &SendGridDriver{
		httpClient: httpClient,
	}
}

func (d *SendGridDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		offset  int
	)

	for range maxPaginationPages {
		resp, err := d.fetchTeammates(ctx, offset)
		if err != nil {
			return nil, err
		}

		teammates := sendGridResponseItems(resp)
		for _, teammate := range teammates {
			if teammate.Email == "" {
				continue
			}

			records = append(records, AccountRecord{
				Email:       teammate.Email,
				FullName:    sendGridFullName(teammate.FirstName, teammate.LastName),
				Role:        sendGridRole(teammate.UserType, teammate.IsAdmin),
				IsAdmin:     teammate.IsAdmin,
				ExternalID:  strings.TrimSpace(teammate.Username),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			})
		}

		if len(teammates) < sendGridTeammatesPageLimit {
			return records, nil
		}

		offset += len(teammates)
	}

	return nil, fmt.Errorf("cannot list all sendgrid teammates: %w", ErrPaginationLimitReached)
}

func (d *SendGridDriver) fetchTeammates(
	ctx context.Context,
	offset int,
) (*sendGridTeammatesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sendGridTeammatesEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create sendgrid teammates request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(sendGridTeammatesPageLimit))
	q.Set("offset", strconv.Itoa(offset))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute sendgrid teammates request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch sendgrid teammates: unexpected status %d", httpResp.StatusCode)
	}

	var resp sendGridTeammatesResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode sendgrid teammates response: %w", err)
	}

	return &resp, nil
}

func sendGridResponseItems(resp *sendGridTeammatesResponse) []sendGridTeammate {
	if len(resp.Result) > 0 {
		return resp.Result
	}

	return resp.Results
}

func sendGridFullName(firstName, lastName string) string {
	return strings.TrimSpace(strings.Join([]string{firstName, lastName}, " "))
}

func sendGridRole(userType string, isAdmin bool) string {
	switch userType {
	case "owner":
		return "Owner"
	case "admin":
		return "Admin"
	case "teammate":
		return "Teammate"
	case "":
		if isAdmin {
			return "Admin"
		}

		return "Teammate"
	default:
		return userType
	}
}
