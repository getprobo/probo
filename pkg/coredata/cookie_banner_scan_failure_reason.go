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

// CookieBannerScanFailureReason explains why a scan ended FAILED.
//
// CookieBannerScanFailureReasonUnreachable: DNS, TLS or connection
// failure on the banner origin.
//
// CookieBannerScanFailureReasonBlocked: the site answered with a bot
// wall or challenge page instead of content.
//
// CookieBannerScanFailureReasonNoPagesLoaded: the origin was reachable
// but no page finished loading within its navigation timeout.
//
// CookieBannerScanFailureReasonInterrupted: the scan stayed RUNNING past
// the stale threshold, typically because the worker crashed.
//
// CookieBannerScanFailureReasonInternal: any other error on our side.
type CookieBannerScanFailureReason string

const (
	CookieBannerScanFailureReasonUnreachable   CookieBannerScanFailureReason = "UNREACHABLE"
	CookieBannerScanFailureReasonBlocked       CookieBannerScanFailureReason = "BLOCKED"
	CookieBannerScanFailureReasonNoPagesLoaded CookieBannerScanFailureReason = "NO_PAGES_LOADED"
	CookieBannerScanFailureReasonInterrupted   CookieBannerScanFailureReason = "INTERRUPTED"
	CookieBannerScanFailureReasonInternal      CookieBannerScanFailureReason = "INTERNAL"
)

var (
	_ fmt.Stringer             = CookieBannerScanFailureReason("")
	_ encoding.TextMarshaler   = CookieBannerScanFailureReason("")
	_ encoding.TextUnmarshaler = (*CookieBannerScanFailureReason)(nil)
)

func CookieBannerScanFailureReasons() []CookieBannerScanFailureReason {
	return []CookieBannerScanFailureReason{
		CookieBannerScanFailureReasonUnreachable,
		CookieBannerScanFailureReasonBlocked,
		CookieBannerScanFailureReasonNoPagesLoaded,
		CookieBannerScanFailureReasonInterrupted,
		CookieBannerScanFailureReasonInternal,
	}
}

func (v CookieBannerScanFailureReason) IsValid() bool {
	switch v {
	case
		CookieBannerScanFailureReasonUnreachable,
		CookieBannerScanFailureReasonBlocked,
		CookieBannerScanFailureReasonNoPagesLoaded,
		CookieBannerScanFailureReasonInterrupted,
		CookieBannerScanFailureReasonInternal:
		return true
	}

	return false
}

func (v CookieBannerScanFailureReason) String() string {
	return string(v)
}

func (v CookieBannerScanFailureReason) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieBannerScanFailureReason) UnmarshalText(text []byte) error {
	val := CookieBannerScanFailureReason(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieBannerScanFailureReason value: %q", string(text))
	}

	*v = val

	return nil
}
