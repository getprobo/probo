// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
