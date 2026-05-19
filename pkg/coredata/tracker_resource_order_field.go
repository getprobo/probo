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

package coredata

import "fmt"

type TrackerResourceOrderField string

const (
	TrackerResourceOrderFieldCreatedAt      TrackerResourceOrderField = "CREATED_AT"
	TrackerResourceOrderFieldLastDetectedAt TrackerResourceOrderField = "LAST_DETECTED_AT"
	TrackerResourceOrderFieldOrigin         TrackerResourceOrderField = "ORIGIN"
	TrackerResourceOrderFieldUpdatedAt      TrackerResourceOrderField = "UPDATED_AT"
)

func (p TrackerResourceOrderField) Column() string {
	switch p {
	case TrackerResourceOrderFieldCreatedAt:
		return "created_at"
	case TrackerResourceOrderFieldLastDetectedAt:
		return "COALESCE(last_detected_at, '0001-01-01T00:00:00Z'::timestamptz)"
	case TrackerResourceOrderFieldOrigin:
		return "origin"
	case TrackerResourceOrderFieldUpdatedAt:
		return "updated_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}

func (p TrackerResourceOrderField) IsValid() bool {
	switch p {
	case TrackerResourceOrderFieldCreatedAt,
		TrackerResourceOrderFieldLastDetectedAt,
		TrackerResourceOrderFieldOrigin,
		TrackerResourceOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (p TrackerResourceOrderField) String() string {
	return string(p)
}

func (p *TrackerResourceOrderField) UnmarshalText(text []byte) error {
	*p = TrackerResourceOrderField(text)
	if !p.IsValid() {
		return fmt.Errorf("%s is not a valid TrackerResourceOrderField", string(text))
	}

	return nil
}

func (p TrackerResourceOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}
