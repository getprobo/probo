// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

type ConsentAction string

const (
	ConsentActionAcceptAll      ConsentAction = "ACCEPT_ALL"
	ConsentActionRejectAll      ConsentAction = "REJECT_ALL"
	ConsentActionCustomize      ConsentAction = "CUSTOMIZE"
	ConsentActionAcceptCategory ConsentAction = "ACCEPT_CATEGORY"
	ConsentActionGPC            ConsentAction = "GPC"
)

func ConsentActions() []ConsentAction {
	return []ConsentAction{
		ConsentActionAcceptAll,
		ConsentActionRejectAll,
		ConsentActionCustomize,
		ConsentActionAcceptCategory,
		ConsentActionGPC,
	}
}

func (a ConsentAction) String() string {
	return string(a)
}

func (a *ConsentAction) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for ConsentAction: %T", value)
	}

	switch ConsentAction(v) {
	case ConsentActionAcceptAll:
		*a = ConsentActionAcceptAll
	case ConsentActionRejectAll:
		*a = ConsentActionRejectAll
	case ConsentActionCustomize:
		*a = ConsentActionCustomize
	case ConsentActionAcceptCategory:
		*a = ConsentActionAcceptCategory
	case ConsentActionGPC:
		*a = ConsentActionGPC
	default:
		return fmt.Errorf("invalid ConsentAction value: %q", v)
	}

	return nil
}

func (a ConsentAction) Value() (driver.Value, error) {
	switch a {
	case ConsentActionAcceptAll,
		ConsentActionRejectAll,
		ConsentActionCustomize,
		ConsentActionAcceptCategory,
		ConsentActionGPC:
		return string(a), nil
	default:
		return nil, fmt.Errorf("invalid ConsentAction: %s", a)
	}
}
