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

package coredata

import (
	"encoding"
	"fmt"

	"go.probo.inc/probo/pkg/page"
)

type (
	AccessReviewCampaignOrderField string
)

const (
	AccessReviewCampaignOrderFieldCreatedAt AccessReviewCampaignOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = AccessReviewCampaignOrderField("")
	_ fmt.Stringer             = AccessReviewCampaignOrderField("")
	_ encoding.TextMarshaler   = AccessReviewCampaignOrderField("")
	_ encoding.TextUnmarshaler = (*AccessReviewCampaignOrderField)(nil)
)

func AccessReviewCampaignOrderFields() []AccessReviewCampaignOrderField {
	return []AccessReviewCampaignOrderField{
		AccessReviewCampaignOrderFieldCreatedAt,
	}
}

func (v AccessReviewCampaignOrderField) IsValid() bool {
	switch v {
	case
		AccessReviewCampaignOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v AccessReviewCampaignOrderField) String() string {
	return string(v)
}

func (v AccessReviewCampaignOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewCampaignOrderField) UnmarshalText(text []byte) error {
	val := AccessReviewCampaignOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewCampaignOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p AccessReviewCampaignOrderField) Column() string {
	switch p {
	case AccessReviewCampaignOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
