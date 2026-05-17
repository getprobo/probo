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

// DeelDriver fetches people from the Deel REST API using a
// pre-authenticated HTTP client (Bearer token). Pagination is
// offset-based: increment `offset` by `limit` until the response
// `data` array is empty.
//
// Notes on data quality:
//   - Active is derived from `hiring_status == "active"`. When `end_date`
//     is set the worker is considered inactive regardless of hiring_status.
//   - MFA and last-login are not exposed by the people endpoint.
type DeelDriver struct {
	httpClient *http.Client
}

var _ Driver = (*DeelDriver)(nil)

func NewDeelDriver(httpClient *http.Client) *DeelDriver {
	return &DeelDriver{httpClient: httpClient}
}

type deelPerson struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	JobTitle     string `json:"job_title"`
	HiringStatus string `json:"hiring_status"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

type deelPeoplePage struct {
	Data []deelPerson `json:"data"`
}

func (d *DeelDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	const limit = 100
	offset := 0

	for range maxPaginationPages {
		people, err := d.queryPeople(ctx, offset, limit)
		if err != nil {
			return nil, err
		}

		if len(people) == 0 {
			return records, nil
		}

		for _, p := range people {
			fullName := strings.TrimSpace(p.FirstName + " " + p.LastName)

			active := p.HiringStatus == "active"
			if p.EndDate != "" {
				active = false
			}

			record := AccountRecord{
				Email:       p.Email,
				FullName:    fullName,
				JobTitle:    p.JobTitle,
				Active:      &active,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
				ExternalID:  p.ID,
			}

			records = append(records, record)
		}

		offset += limit
	}

	return nil, fmt.Errorf("cannot list all deel accounts: %w", ErrPaginationLimitReached)
}

func (d *DeelDriver) queryPeople(ctx context.Context, offset, limit int) ([]deelPerson, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	endpoint := "https://api.letsdeel.com/rest/v2/people?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create deel people request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute deel people request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch deel people: unexpected status %d", httpResp.StatusCode)
	}

	var page deelPeoplePage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode deel people response: %w", err)
	}

	return page.Data, nil
}
