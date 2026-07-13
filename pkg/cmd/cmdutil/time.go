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

package cmdutil

import (
	"fmt"
	"time"
)

// FormatTime parses an RFC3339 timestamp and returns a human-friendly
// local representation. If parsing fails it returns the raw string.
func FormatTime(raw string) string {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}

	return t.Local().Format("Jan 02, 2006 15:04 MST")
}

// NormalizeDatetime converts a user-supplied date or datetime into the RFC3339
// format expected by the GraphQL Datetime scalar. A full RFC3339 timestamp is
// returned unchanged; a date-only value (YYYY-MM-DD) is anchored to midnight
// UTC. Any other format returns an error.
func NormalizeDatetime(raw string) (string, error) {
	if _, err := time.Parse(time.RFC3339, raw); err == nil {
		return raw, nil
	}

	if t, err := time.Parse(time.DateOnly, raw); err == nil {
		return t.Format(time.RFC3339), nil
	}

	return "", fmt.Errorf(
		"invalid date %q: expected YYYY-MM-DD or RFC3339 (e.g. 2026-01-02 or 2026-01-02T15:04:05Z)",
		raw,
	)
}
