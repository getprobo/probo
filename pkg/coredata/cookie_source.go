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

type CookieSource string

// CookieSourceScript: JS write observed via a detector hook on the
// page realm's prototypes. In practice this is a page-script write
// with high confidence -- isolated-world content scripts use their
// own realm's prototypes and never trip the hook (they don't land
// in this bucket at all), and page-world extensions (MV3 main world,
// userscripts with @grant none) reliably leave a browser-extension
// frame on the stack and classify as CookieSourceExtension. The
// only residual contamination is rare cases where a page-world
// extension's frame gets stripped from the stack (deep async,
// page-side trampolines).
//
// CookieSourceExtension: synchronous JS write whose stack at the
// hook contained at least one chrome-/moz-/safari-web-extension
// frame. A page-world extension write is confirmed.
//
// CookieSourcePreExisting: enumerated from the storage at SDK init
// rather than observed at write time. This is the catch-all bucket:
// it bundles real pre-existing cookies/storage from prior sessions,
// HTTP-set cookies that landed before our SDK ran, and -- crucially
// -- writes from any extension realm (including isolated-world
// content scripts) that happened before SDK init. Many extensions
// inject at document_start specifically to set state before page
// scripts run, so a meaningful share of PRE_EXISTING rows can be
// extension-origin even though we cannot prove it. Treat this value
// as low-signal for "is this a real page tracker" decisions.
//
// CookieSourceHTTP: cookie change observed via the CookieStore API
// change event. Set by the server (Set-Cookie response header).
//
// Rows persisted before CookieSourceExtension was introduced cannot
// be backfilled -- the stack at write time is gone -- so historic
// SCRIPT rows retain the (mild) ambiguity above for that period.
const (
	CookieSourceScript      CookieSource = "SCRIPT"
	CookieSourcePreExisting CookieSource = "PRE_EXISTING"
	CookieSourceHTTP        CookieSource = "HTTP"
	CookieSourceExtension   CookieSource = "EXTENSION"
)

var (
	_ fmt.Stringer             = CookieSource("")
	_ encoding.TextMarshaler   = CookieSource("")
	_ encoding.TextUnmarshaler = (*CookieSource)(nil)
)

func CookieSources() []CookieSource {
	return []CookieSource{
		CookieSourceScript,
		CookieSourcePreExisting,
		CookieSourceHTTP,
		CookieSourceExtension,
	}
}

func (v CookieSource) IsValid() bool {
	switch v {
	case
		CookieSourceScript,
		CookieSourcePreExisting,
		CookieSourceHTTP,
		CookieSourceExtension:
		return true
	}

	return false
}

func (v CookieSource) String() string {
	return string(v)
}

func (v CookieSource) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieSource) UnmarshalText(text []byte) error {
	val := CookieSource(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieSource value: %q", string(text))
	}

	*v = val

	return nil
}
