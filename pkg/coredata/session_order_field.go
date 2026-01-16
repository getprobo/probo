// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

type (
	SessionOrderField string
)

const (
	SessionOrderFieldCreatedAt SessionOrderField = "CREATED_AT"
	SessionOrderFieldExpiredAt SessionOrderField = "EXPIRED_AT"
	SessionOrderFieldUpdatedAt SessionOrderField = "UPDATED_AT"
)

func (p SessionOrderField) Column() string {
	switch p {
	case SessionOrderFieldCreatedAt:
		return "created_at"
	case SessionOrderFieldExpiredAt:
		return "expired_at"
	case SessionOrderFieldUpdatedAt:
		return "updated_at"
	}

	return string(p)
}

func (p SessionOrderField) String() string {
	return string(p)
}

func (p SessionOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *SessionOrderField) UnmarshalText(text []byte) error {
	*p = SessionOrderField(text)
	return nil
}
