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
	"net/url"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

type SendGridDriver struct {
	httpClient *http.Client
}

var _ Driver = (*SendGridDriver)(nil)

type sendGridTeammate struct {
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	UserType  string   `json:"user_type"`
	IsAdmin   bool     `json:"is_admin"`
	Scopes    []string `json:"scopes"`
}

type sendGridTeammatesResponse struct {
	Result  []sendGridTeammate `json:"result"`
	Results []sendGridTeammate `json:"results"`
}

type sendGridTeammateResponse struct {
	Result sendGridTeammate `json:"result"`
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

			mfaStatus := sendGridMFAStatus(teammate.Scopes)
			if mfaStatus == coredata.MFAStatusUnknown && teammate.Username != "" {
				detailedTeammate, err := d.fetchTeammate(ctx, teammate.Username)
				if err == nil {
					mfaStatus = sendGridMFAStatus(detailedTeammate.Scopes)
				}
			}

			records = append(records, AccountRecord{
				Email:       teammate.Email,
				FullName:    sendGridFullName(teammate.FirstName, teammate.LastName),
				Role:        sendGridRole(teammate.UserType, teammate.IsAdmin),
				IsAdmin:     teammate.IsAdmin,
				ExternalID:  strings.TrimSpace(teammate.Username),
				MFAStatus:   mfaStatus,
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

func (d *SendGridDriver) fetchTeammate(ctx context.Context, username string) (*sendGridTeammate, error) {
	endpoint, err := url.JoinPath(sendGridTeammatesEndpoint, url.PathEscape(username))
	if err != nil {
		return nil, fmt.Errorf("cannot build sendgrid teammate details url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create sendgrid teammate details request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute sendgrid teammate details request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch sendgrid teammate details: unexpected status %d", httpResp.StatusCode)
	}

	var resp sendGridTeammateResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode sendgrid teammate details response: %w", err)
	}

	return &resp.Result, nil
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

// sendGridMFAStatus derives a teammate's MFA status from the auto-set 2fa
// scopes SendGrid attaches to the teammate detail. A restricted teammate
// carries exactly one of them to reflect their real status. Full-access
// users (the account owner and full-access teammates) are the exception:
// their scope list is the entire catalog and therefore contains BOTH
// 2fa_exempt and 2fa_required, which says nothing about their actual MFA.
// Only report a definitive status when exactly one scope is present;
// both-or-neither is ambiguous, so report Unknown rather than guessing.
func sendGridMFAStatus(scopes []string) coredata.MFAStatus {
	var exempt, required bool
	for _, scope := range scopes {
		switch scope {
		case "2fa_exempt":
			exempt = true
		case "2fa_required":
			required = true
		}
	}

	switch {
	case required && !exempt:
		return coredata.MFAStatusEnabled
	case exempt && !required:
		return coredata.MFAStatusDisabled
	default:
		return coredata.MFAStatusUnknown
	}
}
