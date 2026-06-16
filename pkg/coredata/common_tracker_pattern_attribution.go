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

// CommonTrackerPatternAttribution is the terminal verdict a catalog row
// carries about who, if anyone, sets the tracker.
//
// CommonTrackerPatternAttributionUndetermined: the pipeline has not
// resolved a vendor yet. The deterministic signals and the mapping agent
// keep probing it (this is the state of the unmatched fallback row).
//
// CommonTrackerPatternAttributionThirdParty: a third party has been
// resolved; the row carries a common_third_party_id.
//
// CommonTrackerPatternAttributionFirstParty: terminal verdict that the
// artifact has no third party — it is the scanned site's own, a generic
// library/log key, an extension key embedding the site origin, or
// otherwise not attributable to any vendor. The mapping pipeline never
// attributes such a row again.
type CommonTrackerPatternAttribution string

const (
	CommonTrackerPatternAttributionUndetermined CommonTrackerPatternAttribution = "UNDETERMINED"
	CommonTrackerPatternAttributionThirdParty   CommonTrackerPatternAttribution = "THIRD_PARTY"
	CommonTrackerPatternAttributionFirstParty   CommonTrackerPatternAttribution = "FIRST_PARTY"
)

var (
	_ fmt.Stringer             = CommonTrackerPatternAttribution("")
	_ encoding.TextMarshaler   = CommonTrackerPatternAttribution("")
	_ encoding.TextUnmarshaler = (*CommonTrackerPatternAttribution)(nil)
)

func CommonTrackerPatternAttributions() []CommonTrackerPatternAttribution {
	return []CommonTrackerPatternAttribution{
		CommonTrackerPatternAttributionUndetermined,
		CommonTrackerPatternAttributionThirdParty,
		CommonTrackerPatternAttributionFirstParty,
	}
}

func (v CommonTrackerPatternAttribution) IsValid() bool {
	switch v {
	case
		CommonTrackerPatternAttributionUndetermined,
		CommonTrackerPatternAttributionThirdParty,
		CommonTrackerPatternAttributionFirstParty:
		return true
	}

	return false
}

func (v CommonTrackerPatternAttribution) String() string {
	return string(v)
}

func (v CommonTrackerPatternAttribution) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CommonTrackerPatternAttribution) UnmarshalText(text []byte) error {
	val := CommonTrackerPatternAttribution(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CommonTrackerPatternAttribution value: %q", string(text))
	}

	*v = val

	return nil
}
