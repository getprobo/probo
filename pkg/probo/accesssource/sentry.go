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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// SentryDriver fetches organization members from the Sentry API.
type SentryDriver struct {
	httpClient       *http.Client
	organizationSlug string
}

func NewSentryDriver(httpClient *http.Client, organizationSlug string) *SentryDriver {
	return &SentryDriver{
		httpClient:       httpClient,
		organizationSlug: organizationSlug,
	}
}

func (d *SentryDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		url     = fmt.Sprintf(
			"https://sentry.io/api/0/organizations/%s/members/?per_page=100",
			d.organizationSlug,
		)
	)

	for url != "" {
		resp, nextURL, err := d.queryMembers(ctx, url)
		if err != nil {
			return nil, err
		}

		for _, m := range resp {
			record := AccountRecord{
				Email:      m.Email,
				FullName:   m.Name,
				Role:       m.OrgRole,
				Active:     !m.Pending && !m.Expired,
				IsAdmin:    m.OrgRole == "owner" || m.OrgRole == "admin" || m.OrgRole == "manager",
				ExternalID: m.ID,
				MFAStatus:  sentryMFAStatus(m.User),
				AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			}

			if m.DateCreated != nil {
				record.CreatedAt = m.DateCreated
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		url = nextURL
	}

	return records, nil
}

func (d *SentryDriver) queryMembers(
	ctx context.Context,
	url string,
) ([]sentryMember, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create sentry members request: %w", err)
	}

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute sentry members request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf(
			"cannot fetch sentry members: unexpected status %d",
			httpResp.StatusCode,
		)
	}

	var members []sentryMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, "", fmt.Errorf("cannot decode sentry members response: %w", err)
	}

	nextURL := parseSentryLinkNext(httpResp.Header.Get("Link"))

	return members, nextURL, nil
}

// parseSentryLinkNext extracts the "next" URL from Sentry's Link header.
// Sentry uses cursor-based pagination via Link headers with rel="next" and
// results="true"/"false".
func parseSentryLinkNext(header string) string {
	if header == "" {
		return ""
	}

	// Sentry Link header format:
	// <url>; rel="previous"; results="false"; cursor="...",
	// <url>; rel="next"; results="true"; cursor="..."
	for _, part := range splitLink(header) {
		if containsParam(part, `rel="next"`) && containsParam(part, `results="true"`) {
			return extractURL(part)
		}
	}

	return ""
}

func splitLink(header string) []string {
	var parts []string
	start := 0
	inAngle := false
	for i, c := range header {
		switch c {
		case '<':
			inAngle = true
		case '>':
			inAngle = false
		case ',':
			if !inAngle {
				parts = append(parts, header[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, header[start:])
	return parts
}

func containsParam(part, param string) bool {
	for i := 0; i <= len(part)-len(param); i++ {
		if part[i:i+len(param)] == param {
			return true
		}
	}
	return false
}

func extractURL(part string) string {
	start := -1
	for i, c := range part {
		switch c {
		case '<':
			start = i + 1
		case '>':
			if start >= 0 {
				return part[start:i]
			}
		}
	}
	return ""
}

func sentryMFAStatus(user *sentryUser) coredata.MFAStatus {
	if user == nil {
		return coredata.MFAStatusUnknown
	}
	if user.Has2FA {
		return coredata.MFAStatusEnabled
	}
	return coredata.MFAStatusDisabled
}

type sentryMember struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	OrgRole     string     `json:"orgRole"`
	Pending     bool       `json:"pending"`
	Expired     bool       `json:"expired"`
	DateCreated *time.Time `json:"dateCreated"`
	User        *sentryUser `json:"user"`
}

type sentryUser struct {
	ID     string `json:"id"`
	Has2FA bool   `json:"has2fa"`
}
