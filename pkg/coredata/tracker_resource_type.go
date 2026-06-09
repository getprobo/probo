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
