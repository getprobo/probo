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
	"net/http"
	"net/url"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	figmaAPIBaseURL = "https://api.figma.com/v1"
)

// FigmaDriver fetches organization members from Figma via OAuth2-authenticated
// REST requests.
type FigmaDriver struct {
	httpClient *http.Client
}

func NewFigmaDriver(httpClient *http.Client) *FigmaDriver {
	return &FigmaDriver{
		httpClient: httpClient,
	}
}

func (d *FigmaDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		before  string
	)

	for {
		resp, err := d.queryMembers(ctx, before)
		if err != nil {
			return nil, err
		}

		for _, m := range resp.Members {
			isAdmin := m.Role == "org_owner" || m.Role == "org_admin"

			record := AccountRecord{
				Email:      m.Email,
				FullName:   m.Name,
				Role:       m.Role,
				Active:     true,
				IsAdmin:    isAdmin,
				ExternalID: m.ID,
				MFAStatus:  coredata.MFAStatusUnknown,
				AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if resp.Pagination.Before == "" {
			break
		}
		before = resp.Pagination.Before
	}

	return records, nil
}

func (d *FigmaDriver) queryMembers(ctx context.Context, before string) (*figmaMembersResponse, error) {
	u, err := url.Parse(figmaAPIBaseURL + "/org/members")
	if err != nil {
		return nil, fmt.Errorf("cannot parse figma members url: %w", err)
	}

	q := u.Query()
	q.Set("page_size", "100")
	if before != "" {
		q.Set("before", before)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create figma members request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute figma members request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch figma members: unexpected status %d", httpResp.StatusCode)
	}

	var resp figmaMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode figma members response: %w", err)
	}

	return &resp, nil
}

type figmaMembersResponse struct {
	Members []struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	} `json:"members"`
	Pagination struct {
		Before string `json:"before"`
	} `json:"pagination"`
}
