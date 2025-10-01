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

import (
	"fmt"
)

type TrustCenterDocumentAccessOrderField string

const (
	TrustCenterDocumentAccessOrderFieldCreatedAt TrustCenterDocumentAccessOrderField = "CREATED_AT"
)

func (tcdaof TrustCenterDocumentAccessOrderField) Column() string {
	return string(tcdaof)
}

func (tcdaof TrustCenterDocumentAccessOrderField) String() string {
	return string(tcdaof)
}

func (tcdaof TrustCenterDocumentAccessOrderField) MarshalText() ([]byte, error) {
	return []byte(tcdaof.String()), nil
}

func (tcdaof *TrustCenterDocumentAccessOrderField) UnmarshalText(text []byte) error {
	val := string(text)
	switch val {
	case string(TrustCenterDocumentAccessOrderFieldCreatedAt):
		*tcdaof = TrustCenterDocumentAccessOrderField(val)
		return nil
	}
	return fmt.Errorf("invalid TrustCenterDocumentAccessOrderField value: %q", val)
}
