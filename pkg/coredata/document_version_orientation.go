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

import (
	"database/sql/driver"
	"fmt"
)

type (
	DocumentVersionOrientation string
)

const (
	DocumentVersionOrientationPortrait  DocumentVersionOrientation = "PORTRAIT"
	DocumentVersionOrientationLandscape DocumentVersionOrientation = "LANDSCAPE"
)

func DocumentVersionOrientations() []DocumentVersionOrientation {
	return []DocumentVersionOrientation{
		DocumentVersionOrientationPortrait,
		DocumentVersionOrientationLandscape,
	}
}

func (o DocumentVersionOrientation) MarshalText() ([]byte, error) {
	return []byte(o.String()), nil
}

func (o *DocumentVersionOrientation) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case DocumentVersionOrientationPortrait.String():
		*o = DocumentVersionOrientationPortrait
	case DocumentVersionOrientationLandscape.String():
		*o = DocumentVersionOrientationLandscape
	default:
		return fmt.Errorf("invalid DocumentVersionOrientation value: %q", val)
	}

	return nil
}

func (o DocumentVersionOrientation) String() string {
	return string(o)
}

func (o *DocumentVersionOrientation) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for DocumentVersionOrientation, expected string got %T", value)
	}

	return o.UnmarshalText([]byte(val))
}

func (o DocumentVersionOrientation) Value() (driver.Value, error) {
	return o.String(), nil
}
