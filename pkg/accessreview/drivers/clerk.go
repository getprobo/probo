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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	clerkUsersEndpoint = "https://api.clerk.com/v1/users"
	clerkUsersPageSize = 100
)

type ClerkDriver struct {
	httpClient *http.Client
}

var _ Driver = (*ClerkDriver)(nil)

type clerkUser struct {
	ID                    string  `json:"id"`
	PrimaryEmailAddressID *string `json:"primary_email_address_id"`
	Username              *string `json:"username"`
	FirstName             *string `json:"first_name"`
	LastName              *string `json:"last_name"`
	PasswordEnabled       bool    `json:"password_enabled"`
	TwoFactorEnabled      bool    `json:"two_factor_enabled"`
	TOTPEnabled           bool    `json:"totp_enabled"`
	BackupCodeEnabled     bool    `json:"backup_code_enabled"`
	Banned                bool    `json:"banned"`
	Locked                bool    `json:"locked"`
	Deprovisioned         bool    `json:"deprovisioned"`
	LastSignInAt          *int64  `json:"last_sign_in_at"`
	CreatedAt             int64   `json:"created_at"`
	EmailAddresses        []struct {
		ID           string `json:"id"`
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
}

func NewClerkDriver(httpClient *http.Client) *ClerkDriver {
	return &ClerkDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

func (d *ClerkDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		offset  = 0
	)

	for range maxPaginationPages {
		users, err := d.fetchUsersPage(ctx, offset)
		if err != nil {
			return nil, err
		}

		for _, u := range users {
			email := clerkPrimaryEmail(u)
			if email == "" {
				continue
			}

			record := AccountRecord{
				Email:       email,
				FullName:    clerkFullName(u, email),
				Active:      new(!u.Banned && !u.Locked && !u.Deprovisioned),
				IsAdmin:     false,
				MFAStatus:   clerkMFAStatus(u),
				AuthMethod:  clerkAuthMethod(u),
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  u.ID,
			}

			if createdAt := clerkUnixMillisToTime(u.CreatedAt); createdAt != nil {
				record.CreatedAt = createdAt
			}

			if u.LastSignInAt != nil {
				record.LastLogin = clerkUnixMillisToTime(*u.LastSignInAt)
			}

			records = append(records, record)
		}

		offset += len(users)

		if len(users) < clerkUsersPageSize {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all clerk users: %w", ErrPaginationLimitReached)
}

func (d *ClerkDriver) fetchUsersPage(ctx context.Context, offset int) ([]clerkUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, clerkUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create clerk users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(clerkUsersPageSize))
	q.Set("offset", strconv.Itoa(offset))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute clerk users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch clerk users: unexpected status %d", httpResp.StatusCode)
	}

	// GET /v1/users returns a bare JSON array of user objects; the total
	// count is exposed separately via /v1/users/count. Decode directly
	// into a slice.
	var users []clerkUser
	if err := json.NewDecoder(httpResp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("cannot decode clerk users response: %w", err)
	}

	return users, nil
}

func clerkPrimaryEmail(u clerkUser) string {
	if u.PrimaryEmailAddressID != nil && *u.PrimaryEmailAddressID != "" {
		for _, email := range u.EmailAddresses {
			if email.ID == *u.PrimaryEmailAddressID && email.EmailAddress != "" {
				return email.EmailAddress
			}
		}
	}

	for _, email := range u.EmailAddresses {
		if email.EmailAddress != "" {
			return email.EmailAddress
		}
	}

	return ""
}

func clerkFullName(u clerkUser, fallback string) string {
	firstName := ""
	lastName := ""
	username := ""

	if u.FirstName != nil {
		firstName = *u.FirstName
	}

	if u.LastName != nil {
		lastName = *u.LastName
	}

	if u.Username != nil {
		username = *u.Username
	}

	fullName := strings.TrimSpace(firstName + " " + lastName)
	if fullName != "" {
		return fullName
	}

	if username != "" {
		return username
	}

	return fallback
}

func clerkMFAStatus(u clerkUser) coredata.MFAStatus {
	if u.TwoFactorEnabled || u.TOTPEnabled || u.BackupCodeEnabled {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

func clerkAuthMethod(u clerkUser) coredata.AccessReviewEntryAuthMethod {
	if u.PasswordEnabled {
		return coredata.AccessReviewEntryAuthMethodPassword
	}

	return coredata.AccessReviewEntryAuthMethodUnknown
}

func clerkUnixMillisToTime(unixMillis int64) *time.Time {
	if unixMillis <= 0 {
		return nil
	}

	t := time.UnixMilli(unixMillis).UTC()

	return &t
}
