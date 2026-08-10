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

package malaysiapdpa

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	TransferRuleVersion = "MY-PDPA-CBPDT-2025-04-29"
	TransferRuleSource  = "https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2025/08/GP_CBPDT_EN-1.pdf"
)

func TransferNextReviewAt(basis coredata.MalaysiaPDPATransferBasis, reviewedAt time.Time) *time.Time {
	if !basis.RequiresTIA() {
		return nil
	}

	targetYear := reviewedAt.Year() + 3
	lastDayOfMonth := time.Date(
		targetYear,
		reviewedAt.Month()+1,
		0,
		reviewedAt.Hour(),
		reviewedAt.Minute(),
		reviewedAt.Second(),
		reviewedAt.Nanosecond(),
		reviewedAt.Location(),
	).Day()
	targetDay := min(reviewedAt.Day(), lastDayOfMonth)
	nextReviewAt := time.Date(
		targetYear,
		reviewedAt.Month(),
		targetDay,
		reviewedAt.Hour(),
		reviewedAt.Minute(),
		reviewedAt.Second(),
		reviewedAt.Nanosecond(),
		reviewedAt.Location(),
	)
	return &nextReviewAt
}
