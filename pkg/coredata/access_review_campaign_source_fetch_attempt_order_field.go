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
	AccessReviewCampaignSourceFetchAttemptOrderField string
)

const (
	AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt AccessReviewCampaignSourceFetchAttemptOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = AccessReviewCampaignSourceFetchAttemptOrderField("")
	_ fmt.Stringer             = AccessReviewCampaignSourceFetchAttemptOrderField("")
	_ encoding.TextMarshaler   = AccessReviewCampaignSourceFetchAttemptOrderField("")
	_ encoding.TextUnmarshaler = (*AccessReviewCampaignSourceFetchAttemptOrderField)(nil)
)

func AccessReviewCampaignSourceFetchAttemptOrderFields() []AccessReviewCampaignSourceFetchAttemptOrderField {
	return []AccessReviewCampaignSourceFetchAttemptOrderField{
		AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt,
	}
}

func (v AccessReviewCampaignSourceFetchAttemptOrderField) IsValid() bool {
	switch v {
	case AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v AccessReviewCampaignSourceFetchAttemptOrderField) String() string {
	return string(v)
}

func (v AccessReviewCampaignSourceFetchAttemptOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewCampaignSourceFetchAttemptOrderField) UnmarshalText(text []byte) error {
	val := AccessReviewCampaignSourceFetchAttemptOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewCampaignSourceFetchAttemptOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p AccessReviewCampaignSourceFetchAttemptOrderField) Column() string {
	switch p {
	case AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
