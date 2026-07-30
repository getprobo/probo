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

//go:build darwin

package update

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.gearno.de/kit/log"
)

// ensureSignatureCompatible refuses an update that would replace a
// Developer ID-signed agent with a binary that lacks the same Apple
// Team ID and code-signing Identifier. A mismatch on either field
// creates a fresh Background Task Management identity.
//
// Machines already running an unsigned (ad-hoc) binary are allowed to
// upgrade so they can recover onto a signed release.
func ensureSignatureCompatible(
	ctx context.Context,
	logger *log.Logger,
	currentPath string,
	candidatePath string,
) error {
	current, err := readCodeSigningIdentity(currentPath)
	if err != nil {
		return fmt.Errorf("cannot read current binary code signature: %w", err)
	}

	candidate, err := readCodeSigningIdentity(candidatePath)
	if err != nil {
		return fmt.Errorf("cannot read candidate binary code signature: %w", err)
	}

	if current.Team == "" {
		logger.InfoCtx(
			ctx,
			"current agent binary has no Apple Team ID; allowing update",
			log.String("candidate_team_id", candidate.Team),
			log.String("candidate_identifier", candidate.Identifier),
		)

		return nil
	}

	return codeSigningIdentitiesCompatible(current, candidate)
}

func readCodeSigningIdentity(path string) (codeSigningIdentity, error) {
	cmd := exec.Command("/usr/bin/codesign", "-d", "--verbose=4", path)
	out, err := cmd.CombinedOutput()
	output := string(out)

	identity := parseCodeSigningIdentity(output)
	if identity.Team != "" {
		return identity, nil
	}

	// Unsigned and ad-hoc binaries exit non-zero and report no Team ID.
	// Treat those as empty rather than failing the update path.
	if err != nil && !isUnsignedCodesignOutput(output) {
		return codeSigningIdentity{}, fmt.Errorf(
			"codesign -d %s: %w: %s",
			path,
			err,
			strings.TrimSpace(output),
		)
	}

	return identity, nil
}
