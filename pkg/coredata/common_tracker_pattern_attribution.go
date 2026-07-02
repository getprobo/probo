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
