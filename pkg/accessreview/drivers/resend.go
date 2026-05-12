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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

type ResendDriver struct {
	httpClient *http.Client
}

var _ Driver = (*ResendDriver)(nil)

type resendAPIKeysResponse struct {
	Data []struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		CreatedAt  string  `json:"created_at"`
		LastUsedAt *string `json:"last_used_at"`
	} `json:"data"`
}

const resendAPIKeysEndpoint = "https://api.resend.com/api-keys"

func NewResendDriver(httpClient *http.Client) *ResendDriver {
	return &ResendDriver{
		httpClient: httpClient,
	}
}

func (d *ResendDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	resp, err := d.fetchAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	var records []AccountRecord
	for _, k := range resp.Data {
		record := AccountRecord{
			FullName:    k.Name,
			IsAdmin:     false,
			ExternalID:  k.ID,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
			AccountType: coredata.AccessEntryAccountTypeServiceAccount,
		}

		if k.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, k.CreatedAt); err == nil {
				record.CreatedAt = &t
			}
		}

		if k.LastUsedAt != nil {
			if t, err := time.Parse(time.RFC3339, *k.LastUsedAt); err == nil {
				record.LastLogin = &t
			}
		}

		if record.FullName != "" || record.Email != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

func (d *ResendDriver) fetchAPIKeys(ctx context.Context) (*resendAPIKeysResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resendAPIKeysEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create resend api-keys request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute resend api-keys request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch resend api-keys: unexpected status %d", httpResp.StatusCode)
	}

	var resp resendAPIKeysResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode resend api-keys response: %w", err)
	}

	return &resp, nil
}
