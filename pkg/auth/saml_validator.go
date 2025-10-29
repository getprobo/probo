// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/crewjam/saml"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/kit/pg"
)

func PreventReplayAttack(
	ctx context.Context,
	conn pg.Conn,
	scope coredata.Scoper,
	assertionID string,
	organizationID gid.GID,
	expiresAt time.Time,
) error {
	var assertion coredata.SAMLAssertion
	exists, err := assertion.CheckExists(ctx, conn, assertionID)
	if err != nil {
		return fmt.Errorf("cannot check assertion ID: %w", err)
	}

	if exists {
		return coredata.ErrAssertionAlreadyUsed{AssertionID: assertionID}
	}

	now := time.Now()
	assertion = coredata.SAMLAssertion{
		ID:             assertionID,
		OrganizationID: organizationID,
		UsedAt:         now,
		ExpiresAt:      expiresAt,
	}

	if err := assertion.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot store assertion ID: %w", err)
	}

	return nil
}

func ValidateAssertion(
	assertion *saml.Assertion,
	expectedAudience string,
	now time.Time,
) error {
	const clockSkewTolerance = 5 * time.Minute

	if assertion.Conditions != nil && !assertion.Conditions.NotBefore.IsZero() {
		if now.Add(clockSkewTolerance).Before(assertion.Conditions.NotBefore) {
			return fmt.Errorf("assertion not yet valid (NotBefore: %v, now: %v, tolerance: %v)",
				assertion.Conditions.NotBefore, now, clockSkewTolerance)
		}
	}

	if assertion.Conditions != nil && !assertion.Conditions.NotOnOrAfter.IsZero() {
		if now.Add(-clockSkewTolerance).After(assertion.Conditions.NotOnOrAfter) ||
		   now.Add(-clockSkewTolerance).Equal(assertion.Conditions.NotOnOrAfter) {
			return fmt.Errorf("assertion expired (NotOnOrAfter: %v, now: %v, tolerance: %v)",
				assertion.Conditions.NotOnOrAfter, now, clockSkewTolerance)
		}
	}

	if assertion.Conditions != nil && len(assertion.Conditions.AudienceRestrictions) > 0 {
		audienceValid := false
		for _, restriction := range assertion.Conditions.AudienceRestrictions {
			if restriction.Audience.Value == expectedAudience {
				audienceValid = true
				break
			}
		}

		if !audienceValid {
			return fmt.Errorf("assertion audience restriction does not match expected audience %q", expectedAudience)
		}
	}

	return nil
}

func CleanupExpiredAssertions(ctx context.Context, conn pg.Conn) (int64, error) {
	return coredata.DeleteExpiredSAMLAssertions(ctx, conn, time.Now())
}

func CleanupExpiredRequests(ctx context.Context, conn pg.Conn) (int64, error) {
	return coredata.DeleteExpiredSAMLRequests(ctx, conn, time.Now())
}

func CleanupExpiredRelayStates(ctx context.Context, conn pg.Conn) (int64, error) {
	return coredata.DeleteExpiredSAMLRelayStates(ctx, conn, time.Now())
}
