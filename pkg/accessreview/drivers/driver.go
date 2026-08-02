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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// AccountRecord represents a single account from an access source or identity
// source. All fields are best-effort; sources populate what they can. Drivers
// must return ALL accounts the source exposes (including inactive / suspended
// / deleted); classification is the job of the reviewer or of an agent run
// against the campaign, not of the fetch pipeline.
//
// Active is three-valued: nil means the source API has no explicit
// account-status signal for this account (the driver cannot tell), a non-nil
// pointer means the driver observed an explicit signal (true = active at
// source, false = deactivated / suspended / deleted). Drivers whose API does
// not distinguish active from deactivated accounts must leave Active nil
// rather than fabricate a value.
type AccountRecord struct {
	Email       string
	FullName    string
	Roles       []string // system roles/permissions (e.g. "Admin", "Viewer")
	JobTitle    string   // HR job title / department (e.g. "Software Engineer")
	Active      *bool
	IsAdmin     bool
	MFAStatus   coredata.MFAStatus
	AuthMethod  coredata.AccessReviewEntryAuthMethod
	AccountType coredata.AccessReviewEntryAccountType
	LastLogin   *time.Time
	CreatedAt   *time.Time
	ExternalID  string // system-specific user ID
}

// maxPaginationPages is the upper bound on the number of pages a driver will
// fetch from an external API. This prevents infinite loops if an API returns
// a non-empty cursor on every response.
const maxPaginationPages = 500

// ErrPaginationLimitReached is returned when a driver exhausts the maximum
// number of pagination pages without reaching the end of the result set.
var ErrPaginationLimitReached = fmt.Errorf("pagination limit of %d pages reached", maxPaginationPages)

// sameHostNextPageURL pins a server-supplied next-page URL to the host the
// driver was configured with, returning it unchanged when it is safe to
// follow. The connection's bearer token is attached to every request, so an
// absolute `next` pointing at a different host — a compromised or spoofed API
// response, or a provider echoing its real host back at a deployment that
// overrode the API base — would forward the token off-host. Refuse cross-host
// pagination; the error is static so it never echoes an attacker-controlled
// host.
//
// The reference is RESOLVED against the base before its host is checked,
// rather than testing url.URL.IsAbs() first. IsAbs reports whether a scheme is
// present, so a protocol-relative reference such as "//evil.example.com/x"
// reads as relative while resolving to that host — checking before resolving
// would wave it straight through. The scheme is compared too, so a spoofed
// "http://" next page cannot downgrade the crawl and send the bearer token in
// cleartext.
func sameHostNextPageURL(provider, baseURL, next string) (string, error) {
	nextURL, err := url.Parse(next)
	if err != nil {
		return "", fmt.Errorf("cannot parse %s next page URL: %w", provider, err)
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse %s base URL: %w", provider, err)
	}

	resolved := base.ResolveReference(nextURL)

	if !strings.EqualFold(resolved.Host, base.Host) || !strings.EqualFold(resolved.Scheme, base.Scheme) {
		return "", fmt.Errorf("cannot follow %s next page URL: cross-host pagination is not allowed", provider)
	}

	// Pinning the host is not enough: "../../oauth2/token" resolves to the same
	// host and would send the connection's bearer token to an endpoint that is
	// not the collection being paged. Requiring the resolved path to stay under
	// the base also stops a relative reference from silently re-rooting — a
	// bare "users?..." against a versioned base resolves to /users, dropping
	// the version segment and quietly paging the wrong API.
	basePath := strings.TrimSuffix(base.EscapedPath(), "/")
	if basePath != "" {
		got := resolved.EscapedPath()
		if got != basePath && !strings.HasPrefix(got, basePath+"/") {
			return "", fmt.Errorf("cannot follow %s next page URL: it leaves the collection path", provider)
		}
	}

	// Resolving normalizes only literal dot segments, so percent-encoded ones
	// ("%2e%2e", "..%2f") survive with the base prefix intact and pass the
	// check above — then reach the wire still encoded, where a server that
	// decodes before routing lands the bearer token on the path that check just
	// refused. Inspect the DECODED path instead, and refuse a dot segment
	// rather than cleaning it, so nothing here has to agree with the server's
	// normalization order. Backslash is a separator too: RFC 3986 does not make
	// it one, but a server that treats it as one would read "..\x" as a climb.
	for _, segment := range strings.FieldsFunc(resolved.Path, isPathSeparator) {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("cannot follow %s next page URL: it leaves the collection path", provider)
		}
	}

	// Credentials embedded in the reference would reach http.NewRequest and, on
	// some transports, take precedence over the Authorization header the
	// connection sets. Drop them; the host is already pinned to the base.
	resolved.User = nil

	return resolved.String(), nil
}

// isPathSeparator reports whether r separates path segments for the purpose of
// the dot-segment check above.
func isPathSeparator(r rune) bool {
	return r == '/' || r == '\\'
}

// Driver defines the interface for fetching accounts from an access or
// identity source. Each driver implementation corresponds to a specific
// system (e.g. Google Workspace, AWS IAM, Probo memberships, CSV).
//
// All sources in a campaign's scope return "who actually has access" data.
type Driver interface {
	// ListAccounts returns all accounts from the source system.
	ListAccounts(ctx context.Context) ([]AccountRecord, error)
}

// parseRFC3339Ptr parses an RFC 3339 timestamp into a *time.Time, returning
// nil for an empty or unparseable value. Drivers use it for best-effort
// timestamp fields (created_at, last_login_at) that an API may omit.
func parseRFC3339Ptr(s string) *time.Time {
	if s == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}

	return &t
}

// activeFromStatus maps a provider status string to the three-valued Active
// signal for providers whose only "live" state is the literal "active" and
// whose remaining status enum is not otherwise enumerated: "active" → active,
// an empty status → nil (no signal), and any other non-empty status →
// inactive. Used by drivers like Pylon and Brevo; a provider with a fully
// known status enum (e.g. Render's active/inactive) maps its own values
// explicitly instead, so an unrecognised value stays nil rather than false.
func activeFromStatus(status string) *bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active":
		active := true

		return &active
	case "":
		return nil
	default:
		inactive := false

		return &inactive
	}
}

// ownerMemberRoles maps a provider role to a display label for providers whose
// role model is exactly owner/member (e.g. Crisp operators, Scaleway org user
// types): "owner" → Owner, "member" → Member. An unknown future value is passed
// through verbatim and no role yields an empty slice.
func ownerMemberRoles(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner":
		return []string{"Owner"}
	case "member":
		return []string{"Member"}
	default:
		if r := strings.TrimSpace(role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}

// isOwnerRole reports whether a provider role is the owner, the only role in the
// owner/member model that grants administrative access.
func isOwnerRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "owner")
}

// retryRoundTripper retries 429 and 5xx responses with exponential backoff.
//
// The retry budget is bounded by what can still produce a useful answer: a
// fetch runs under a per-source deadline, so a sleep that outlives the
// deadline, or that follows the final attempt, only converts a reportable
// provider status into an opaque timeout. Every wait below is therefore
// guarded.
type retryRoundTripper struct {
	next       http.RoundTripper
	maxRetries int
}

const (
	// retryBaseBackoff is the first backoff step; it doubles per attempt.
	retryBaseBackoff = 250 * time.Millisecond
	// maxRetryWait caps a single wait. A provider asking for longer (via
	// Retry-After) cannot be accommodated inside a per-source budget, so the
	// throttled response is returned instead and the caller reports it.
	maxRetryWait = 5 * time.Second
)

func (rt *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := rt.next
	if transport == nil {
		transport = http.DefaultTransport
	}

	var lastResp *http.Response

	for attempt := range rt.maxRetries {
		resp, err := transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return resp, nil
		}

		// Buffer and re-attach the body so the caller can still read it
		// if this turns out to be the final (retry-exhausted) response.
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))
		lastResp = resp

		// Nothing follows the last attempt, so waiting here would burn the
		// caller's deadline to return the response it already has.
		if attempt == rt.maxRetries-1 {
			break
		}

		wait := retryBaseBackoff << attempt
		// Retry-After is authoritative on 429: retrying sooner just earns
		// another 429 and spends an attempt doing it.
		if after, ok := retryAfter(resp); ok {
			wait = after
		}

		if wait > maxRetryWait {
			break
		}

		// Sleeping past the deadline guarantees a context error that hides
		// the provider's actual status from the caller.
		if deadline, ok := req.Context().Deadline(); ok && time.Until(deadline) <= wait {
			break
		}

		timer := time.NewTimer(wait)

		select {
		case <-req.Context().Done():
			timer.Stop()

			return nil, req.Context().Err()
		case <-timer.C:
		}
	}

	return lastResp, nil
}

// retryAfter reads a Retry-After header in either documented form —
// delta-seconds or an HTTP-date. The bool reports whether the header was
// present and parseable; a date already in the past yields a zero wait.
func retryAfter(resp *http.Response) (time.Duration, bool) {
	value := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, false
		}

		return time.Duration(seconds) * time.Second, true
	}

	if at, err := http.ParseTime(value); err == nil {
		return max(time.Until(at), 0), true
	}

	return 0, false
}
