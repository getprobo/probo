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
	TrustCenterReferenceOrderField string
)

const (
	TrustCenterReferenceOrderFieldRank      TrustCenterReferenceOrderField = "RANK"
	TrustCenterReferenceOrderFieldName      TrustCenterReferenceOrderField = "NAME"
	TrustCenterReferenceOrderFieldCreatedAt TrustCenterReferenceOrderField = "CREATED_AT"
	TrustCenterReferenceOrderFieldUpdatedAt TrustCenterReferenceOrderField = "UPDATED_AT"
)

func (p TrustCenterReferenceOrderField) Column() string {
	switch p {
	case TrustCenterReferenceOrderFieldRank:
		return "rank"
	case TrustCenterReferenceOrderFieldName:
		return "name"
	case TrustCenterReferenceOrderFieldCreatedAt:
		return "created_at"
	case TrustCenterReferenceOrderFieldUpdatedAt:
		return "updated_at"
	default:
		return string(p)
	}
}

func (p TrustCenterReferenceOrderField) String() string {
	return string(p)
}

func (p TrustCenterReferenceOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *TrustCenterReferenceOrderField) UnmarshalText(text []byte) error {
	*p = TrustCenterReferenceOrderField(text)
	return nil
}
