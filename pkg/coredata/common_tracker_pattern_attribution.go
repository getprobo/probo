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
// artifact belongs to the site operator — its own tracker, or a bundled
// library that stores state locally and egresses nothing.
//
// CommonTrackerPatternAttributionNotAttributable: terminal verdict that the
// artifact comes from real software that is neither the operator's nor a
// vendor they engaged, so it belongs in nobody's register. Browser
// extensions and other visitor-installed tooling inject keys into a page
// the operator does not control. Distinct from FIRST_PARTY because calling
// an extension the site's own code is simply false, and a register that
// says so is wrong in a way an operator would notice.
//
// Both are terminal: the mapping pipeline never attributes such a row
// again. Use IsTerminal rather than comparing against FIRST_PARTY.
type CommonTrackerPatternAttribution string

const (
	CommonTrackerPatternAttributionUndetermined CommonTrackerPatternAttribution = "UNDETERMINED"
	CommonTrackerPatternAttributionThirdParty   CommonTrackerPatternAttribution = "THIRD_PARTY"
	CommonTrackerPatternAttributionFirstParty   CommonTrackerPatternAttribution = "FIRST_PARTY"

	// CommonTrackerPatternAttributionNotAttributable covers software that is
	// neither the operator's nor a third party they engaged.
	CommonTrackerPatternAttributionNotAttributable CommonTrackerPatternAttribution = "NOT_ATTRIBUTABLE"
)

// IsTerminal reports whether the verdict settles the question of who set the
// artifact, so the mapping pipeline stops probing it. Callers must use this
// rather than testing for FIRST_PARTY, which would silently keep re-probing
// every other terminal verdict.
// terminalAttributions lists the verdicts that settle attribution, for SQL
// that must treat them as a set. Kept beside IsTerminal so the two cannot
// disagree about which verdicts are terminal.
//
// Returns strings because pgx has no encode plan for a slice of the enum type;
// the column accepts the text form.
func terminalAttributions() []string {
	return []string{
		string(CommonTrackerPatternAttributionFirstParty),
		string(CommonTrackerPatternAttributionNotAttributable),
	}
}

func (v CommonTrackerPatternAttribution) IsTerminal() bool {
	switch v {
	case CommonTrackerPatternAttributionFirstParty,
		CommonTrackerPatternAttributionNotAttributable:
		return true
	}

	return false
}

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
		CommonTrackerPatternAttributionNotAttributable,
	}
}

func (v CommonTrackerPatternAttribution) IsValid() bool {
	switch v {
	case
		CommonTrackerPatternAttributionUndetermined,
		CommonTrackerPatternAttributionThirdParty,
		CommonTrackerPatternAttributionFirstParty,
		CommonTrackerPatternAttributionNotAttributable:
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
