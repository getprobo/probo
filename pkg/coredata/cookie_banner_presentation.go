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

// CookieBannerPresentation is the UI presentation policy of the banner. It is
// a separate concern from the consent mode: the consent mode (OPT_IN /
// OPT_OUT) is the legal tracker-firing model, while the presentation decides
// how the banner looks and behaves (which buttons, initial visibility, wording
// variant).
//
// A regulation maps to exactly one presentation; several regulations can share
// the same presentation.
type CookieBannerPresentation string

const (
	// CookieBannerPresentationOptIn blocks non-necessary trackers until the
	// visitor gives explicit consent and offers accept / reject / customize
	// choices.
	CookieBannerPresentationOptIn CookieBannerPresentation = "OPT_IN"

	// CookieBannerPresentationOptOut lets trackers fire immediately but must
	// offer the visitor a way to opt out (e.g. CCPA "Do Not Sell or Share").
	// Nothing opens on first load: the settings link the integrator places in
	// the footer is the only surface the visitor ever sees unaided.
	CookieBannerPresentationOptOut CookieBannerPresentation = "OPT_OUT"

	// CookieBannerPresentationNotice is a purely informational banner used
	// where the law mandates a *visible* cookie notice without requiring
	// consent: trackers fire immediately and the visitor only acknowledges the
	// notice. It is the only non-opt-in presentation that opens on first load,
	// so a regulation belongs here only when a statute demands the disclosure
	// be displayed rather than merely published in a privacy policy.
	CookieBannerPresentationNotice CookieBannerPresentation = "NOTICE"
)

var (
	_ fmt.Stringer             = CookieBannerPresentation("")
	_ encoding.TextMarshaler   = CookieBannerPresentation("")
	_ encoding.TextUnmarshaler = (*CookieBannerPresentation)(nil)
)

func CookieBannerPresentations() []CookieBannerPresentation {
	return []CookieBannerPresentation{
		CookieBannerPresentationOptIn,
		CookieBannerPresentationOptOut,
		CookieBannerPresentationNotice,
	}
}

func (v CookieBannerPresentation) IsValid() bool {
	switch v {
	case
		CookieBannerPresentationOptIn,
		CookieBannerPresentationOptOut,
		CookieBannerPresentationNotice:
		return true
	}

	return false
}

func (v CookieBannerPresentation) String() string {
	return string(v)
}

func (v CookieBannerPresentation) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieBannerPresentation) UnmarshalText(text []byte) error {
	val := CookieBannerPresentation(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieBannerPresentation value: %q", string(text))
	}

	*v = val

	return nil
}
