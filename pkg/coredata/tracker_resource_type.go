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

type TrackerResourceType string

const (
	TrackerResourceTypeScript        TrackerResourceType = "SCRIPT"
	TrackerResourceTypeIframe        TrackerResourceType = "IFRAME"
	TrackerResourceTypeImage         TrackerResourceType = "IMAGE"
	TrackerResourceTypeStylesheet    TrackerResourceType = "STYLESHEET"
	TrackerResourceTypeFont          TrackerResourceType = "FONT"
	TrackerResourceTypeBeacon        TrackerResourceType = "BEACON"
	TrackerResourceTypeFetch         TrackerResourceType = "FETCH"
	TrackerResourceTypeMedia         TrackerResourceType = "MEDIA"
	TrackerResourceTypeServiceWorker TrackerResourceType = "SERVICE_WORKER"
)

var (
	_ fmt.Stringer             = TrackerResourceType("")
	_ encoding.TextMarshaler   = TrackerResourceType("")
	_ encoding.TextUnmarshaler = (*TrackerResourceType)(nil)
)

func TrackerResourceTypes() []TrackerResourceType {
	return []TrackerResourceType{
		TrackerResourceTypeScript,
		TrackerResourceTypeIframe,
		TrackerResourceTypeImage,
		TrackerResourceTypeStylesheet,
		TrackerResourceTypeFont,
		TrackerResourceTypeBeacon,
		TrackerResourceTypeFetch,
		TrackerResourceTypeMedia,
		TrackerResourceTypeServiceWorker,
	}
}

func (v TrackerResourceType) IsValid() bool {
	switch v {
	case
		TrackerResourceTypeScript,
		TrackerResourceTypeIframe,
		TrackerResourceTypeImage,
		TrackerResourceTypeStylesheet,
		TrackerResourceTypeFont,
		TrackerResourceTypeBeacon,
		TrackerResourceTypeFetch,
		TrackerResourceTypeMedia,
		TrackerResourceTypeServiceWorker:
		return true
	}

	return false
}

func (v TrackerResourceType) String() string {
	return string(v)
}

func (v TrackerResourceType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TrackerResourceType) UnmarshalText(text []byte) error {
	val := TrackerResourceType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TrackerResourceType value: %q", string(text))
	}

	*v = val

	return nil
}
