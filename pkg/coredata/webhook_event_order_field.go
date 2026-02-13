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

type (
	WebhookEventOrderField string
)

const (
	WebhookEventOrderFieldCreatedAt WebhookEventOrderField = "CREATED_AT"
)

func (p WebhookEventOrderField) Column() string {
	return string(p)
}

func (p WebhookEventOrderField) String() string {
	return string(p)
}

func (p WebhookEventOrderField) IsValid() bool {
	switch p {
	case WebhookEventOrderFieldCreatedAt:
		return true
	}
	return false
}

func (p WebhookEventOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *WebhookEventOrderField) UnmarshalText(text []byte) error {
	*p = WebhookEventOrderField(text)
	if !p.IsValid() {
		return fmt.Errorf("%s is not a valid WebhookEventOrderField", string(text))
	}
	return nil
}
