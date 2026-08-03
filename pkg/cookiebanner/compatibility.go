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

package cookiebanner

import (
	"strconv"
	"strings"
)

// This file isolates the SDK-version-based backward-compatibility logic for the
// served banner config, so the rest of the service stays free of shims.
//
// There are three eras of client, keyed off the SDK version:
//
//   - <=0.2 ("pre-remap"): render banner text on their own; they must receive
//     the raw text keys untouched.
//   - 0.3-0.10 ("remap shim"): infer their UI from the generic text keys, so
//     the server remaps the active variant's wording onto those keys.
//   - >=0.11 ("layout-aware"): understand the structured layout and select
//     their own wording per layout.text_variant; they receive the raw keys.

// applyBannerTextCompat adapts the config's text keys for clients that cannot
// consume the structured layout. Layout-aware and pre-remap clients receive the
// config untouched; only remap-shim clients get the variant wording folded onto
// the generic keys.
func applyBannerTextCompat(config *BannerConfig, sdkVersion string) {
	if supportsLayout(sdkVersion) || !supportsTextRemap(sdkVersion) {
		return
	}

	variant := config.Layout.TextVariant

	// Remap-shim clients cannot render the notice presentation, so degrade its
	// wording to opt-out: it matches the notice firing model (cookies fire
	// immediately) and renders coherently with their fixed button set.
	if variant == TextVariantNotice {
		variant = TextVariantOptOut
	}

	remapTextsForVariant(config.Texts, variant)
}

// supportsLayout reports whether the SDK version understands the structured
// layout / presentation fields, which shipped in 0.11. Empty or unparseable
// versions are treated as current (and therefore layout-aware).
func supportsLayout(version string) bool {
	if version == "" {
		return true
	}

	major, minor, ok := parseMajorMinor(version)
	if !ok {
		return true
	}

	if major > 0 {
		return true
	}

	return minor >= 11
}

// supportsTextRemap reports whether the SDK version relies on the server folding
// variant wording onto the generic text keys, which began in 0.3. Clients older
// than that (<=0.2) render text on their own and must receive the raw keys
// untouched. Empty or unparseable versions are treated as current.
func supportsTextRemap(version string) bool {
	if version == "" {
		return true
	}

	major, minor, ok := parseMajorMinor(version)
	if !ok {
		return true
	}

	if major > 0 {
		return true
	}

	return minor >= 3
}

// remapTextsForVariant overrides the generic banner text keys with the
// variant-specific wording so remap-shim clients, which infer their UI from the
// generic keys, render the appropriate copy without presentation awareness.
//
// It only handles the opt-out variant: the notice presentation is unrenderable
// by those clients and is degraded to opt-out by the caller before this runs.
func remapTextsForVariant(texts map[string]string, variant TextVariant) {
	if texts == nil {
		return
	}

	if variant == TextVariantOptOut {
		remapTextKey(texts, "banner_title_opt_out", "banner_title")
		remapTextKey(texts, "banner_description_opt_out", "banner_description")
		remapTextKey(texts, "button_acknowledge", "button_accept_all")
		remapTextKey(texts, "button_opt_out", "button_reject_all")
		texts["button_customize"] = ""
	}
}

func remapTextKey(texts map[string]string, src, dst string) {
	if v, ok := texts[src]; ok && v != "" {
		texts[dst] = v
	}
}

func parseMajorMinor(version string) (major, minor int, ok bool) {
	v := strings.TrimPrefix(version, "v")

	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return 0, 0, false
	}

	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}

	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}

	return maj, min, true
}
