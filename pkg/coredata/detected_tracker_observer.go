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

// DetectedTrackerObserver records what observed a detected tracker,
// independently of CookieSource, which records how it was written.
//
// DetectedTrackerObserverSDK: reported by the cookie banner SDK running
// in real visitors' browsers.
//
// DetectedTrackerObserverScanner: observed by the server-side scanner
// from a fresh browser profile.
type DetectedTrackerObserver string

const (
	DetectedTrackerObserverSDK     DetectedTrackerObserver = "SDK"
	DetectedTrackerObserverScanner DetectedTrackerObserver = "SCANNER"
)

var (
	_ fmt.Stringer             = DetectedTrackerObserver("")
	_ encoding.TextMarshaler   = DetectedTrackerObserver("")
	_ encoding.TextUnmarshaler = (*DetectedTrackerObserver)(nil)
)

func DetectedTrackerObservers() []DetectedTrackerObserver {
	return []DetectedTrackerObserver{
		DetectedTrackerObserverSDK,
		DetectedTrackerObserverScanner,
	}
}

func (v DetectedTrackerObserver) IsValid() bool {
	switch v {
	case
		DetectedTrackerObserverSDK,
		DetectedTrackerObserverScanner:
		return true
	}

	return false
}

func (v DetectedTrackerObserver) String() string {
	return string(v)
}

func (v DetectedTrackerObserver) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DetectedTrackerObserver) UnmarshalText(text []byte) error {
	val := DetectedTrackerObserver(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DetectedTrackerObserver value: %q", string(text))
	}

	*v = val

	return nil
}
