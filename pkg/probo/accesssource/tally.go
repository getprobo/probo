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

package accesssource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// TallyDriver fetches users and invites from Tally via REST API requests.
type TallyDriver struct {
	httpClient     *http.Client
	organizationID string
}

func NewTallyDriver(httpClient *http.Client, organizationID string) *TallyDriver {
	return &TallyDriver{
		httpClient:     httpClient,
		organizationID: organizationID,
	}
}

func (d *TallyDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	records, err := d.listUsers(ctx)
	if err != nil {
		return nil, err
	}

	inviteRecords, err := d.listInvites(ctx)
	if err != nil {
		return nil, err
	}

	records = append(records, inviteRecords...)

	return records, nil
}

func (d *TallyDriver) listUsers(ctx context.Context) ([]AccountRecord, error) {
	url := fmt.Sprintf(
		"https://api.tally.so/organizations/%s/users",
		d.organizationID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create tally users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute tally users request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf(
			"tally users request failed with status %d: %s",
			httpResp.StatusCode,
			string(bodyBytes),
		)
	}

	var users []tallyUser
	if err := json.NewDecoder(httpResp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("cannot decode tally users response: %w", err)
	}

	var records []AccountRecord
	for _, u := range users {
		mfaStatus := coredata.MFAStatusDisabled
		if u.HasTwoFactorEnabled {
			mfaStatus = coredata.MFAStatusEnabled
		}

		record := AccountRecord{
			Email:      u.Email,
			FullName:   u.FullName,
			Active:     !u.IsDeleted,
			ExternalID: u.ID,
			MFAStatus:  mfaStatus,
			AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			CreatedAt:  new(u.CreatedAt),
		}

		if record.Email != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

func (d *TallyDriver) listInvites(ctx context.Context) ([]AccountRecord, error) {
	url := fmt.Sprintf(
		"https://api.tally.so/organizations/%s/invites",
		d.organizationID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create tally invites request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute tally invites request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf(
			"tally invites request failed with status %d: %s",
			httpResp.StatusCode,
			string(bodyBytes),
		)
	}

	var invites []tallyInvite
	if err := json.NewDecoder(httpResp.Body).Decode(&invites); err != nil {
		return nil, fmt.Errorf("cannot decode tally invites response: %w", err)
	}

	var records []AccountRecord
	for _, inv := range invites {
		record := AccountRecord{
			Email:      inv.Email,
			Active:     false,
			ExternalID: inv.ID,
			MFAStatus:  coredata.MFAStatusUnknown,
			AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			Role:       "Invited",
		}

		if record.Email != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

type tallyUser struct {
	ID                  string    `json:"id"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	FullName            string    `json:"fullName"`
	Email               string    `json:"email"`
	IsDeleted           bool      `json:"isDeleted"`
	HasTwoFactorEnabled bool      `json:"hasTwoFactorEnabled"`
	CreatedAt           time.Time `json:"createdAt"`
}

type tallyInvite struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
