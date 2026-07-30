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

package update

import (
	"fmt"
	"strings"
)

// codeSigningIdentity is the Apple code-signing identity fields that
// Background Task Management uses to key a background item.
type codeSigningIdentity struct {
	Team       string
	Identifier string
}

// codeSigningIdentitiesCompatible refuses an update that would change
// the BTM identity of a Developer ID-signed agent. Both Team ID and
// signing Identifier must match. Machines already running an unsigned
// (ad-hoc) binary are allowed to upgrade onto a signed release.
func codeSigningIdentitiesCompatible(
	current codeSigningIdentity,
	candidate codeSigningIdentity,
) error {
	if current.Team == "" {
		return nil
	}

	if candidate.Team == "" {
		return fmt.Errorf(
			"candidate binary has no Apple Team ID (current Team ID %s); refusing signature downgrade",
			current.Team,
		)
	}

	if candidate.Team != current.Team {
		return fmt.Errorf(
			"candidate Team ID %s does not match current Team ID %s",
			candidate.Team,
			current.Team,
		)
	}

	if candidate.Identifier != current.Identifier {
		return fmt.Errorf(
			"candidate Identifier %q does not match current Identifier %q",
			candidate.Identifier,
			current.Identifier,
		)
	}

	return nil
}

func parseCodeSigningIdentity(codesignOutput string) codeSigningIdentity {
	return codeSigningIdentity{
		Team:       parseCodesignField(codesignOutput, "TeamIdentifier="),
		Identifier: parseCodesignField(codesignOutput, "Identifier="),
	}
}

func parseCodesignField(codesignOutput string, prefix string) string {
	for line := range strings.SplitSeq(codesignOutput, "\n") {
		line = strings.TrimSpace(line)

		after, ok := strings.CutPrefix(line, prefix)
		if !ok {
			continue
		}

		after = strings.TrimSpace(after)
		if after == "" || after == "not set" {
			return ""
		}

		return after
	}

	return ""
}

func isUnsignedCodesignOutput(output string) bool {
	lower := strings.ToLower(output)

	return strings.Contains(lower, "code object is not signed") ||
		strings.Contains(lower, "not signed at all") ||
		strings.Contains(lower, "teamidentifier=not set")
}
