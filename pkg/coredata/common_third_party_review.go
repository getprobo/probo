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
)

// CommonThirdPartyReview records whether a human has confirmed what a
// catalog row is. It exists because nothing else in the row can express
// it: an agent-created entry that somebody checked and kept is otherwise
// indistinguishable from one nobody has ever read, so every sweep
// re-adjudicates the same rows from keyword guesses.
//
// CommonThirdPartyReviewUnreviewed: the default. No one has judged this
// row. Seeded rows are validated by the seed, so a large unreviewed set
// is the backlog rather than a problem.
//
// CommonThirdPartyReviewValidated: a human confirmed the row names an
// entity an organization could actually engage — a company, with the
// legal identity that implies. This is orthogonal to whether the vendor
// egresses anything from a browser: a payroll processor or an auditor is
// a valid register entry with no tracker at all, which is why the test is
// entity nature and not data flow.
//
// CommonThirdPartyReviewRejected: a human determined the row does not
// name an engageable entity — a bundled library, a page artifact, or
// software the visitor installed. The row is kept rather than deleted:
// deleting it only means the next scan recreates it and someone
// adjudicates it again, whereas a rejected row tells the mapping
// pipeline that patterns resolving to this name earn RejectedVerdict
// instead of a vendor link. That turns one review into a standing answer.
type CommonThirdPartyReview string

const (
	CommonThirdPartyReviewUnreviewed CommonThirdPartyReview = "UNREVIEWED"
	CommonThirdPartyReviewValidated  CommonThirdPartyReview = "VALIDATED"
	CommonThirdPartyReviewRejected   CommonThirdPartyReview = "REJECTED"
)

func CommonThirdPartyReviews() []CommonThirdPartyReview {
	return []CommonThirdPartyReview{
		CommonThirdPartyReviewUnreviewed,
		CommonThirdPartyReviewValidated,
		CommonThirdPartyReviewRejected,
	}
}

func (v CommonThirdPartyReview) IsValid() bool {
	switch v {
	case
		CommonThirdPartyReviewUnreviewed,
		CommonThirdPartyReviewValidated,
		CommonThirdPartyReviewRejected:
		return true
	}

	return false
}

func (v CommonThirdPartyReview) String() string {
	return string(v)
}

func (v CommonThirdPartyReview) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CommonThirdPartyReview) UnmarshalText(text []byte) error {
	val := CommonThirdPartyReview(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CommonThirdPartyReview value: %q", string(text))
	}

	*v = val

	return nil
}

var (
	_ fmt.Stringer             = CommonThirdPartyReview("")
	_ encoding.TextMarshaler   = CommonThirdPartyReview("")
	_ encoding.TextUnmarshaler = (*CommonThirdPartyReview)(nil)
)
