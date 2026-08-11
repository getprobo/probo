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

package console_test

import (
	"fmt"
	"testing"

	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/journey"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// TestMembershipAccess_DisableSignupJourney covers the private-instance gate:
// when PROBOD_AUTH_DISABLE_SIGNUP is set, an authenticated identity without an
// active organization membership must not create organizations or reach the
// console API. Members who are not owners also cannot create organizations.
// Magic-link identity creation itself stays allowed so the compliance portal
// keeps working.
func TestMembershipAccess_DisableSignupJourney(t *testing.T) {
	t.Parallel()

	env := testutil.StartIsolatedEnv(
		t,
		testutil.IsolatedEnvOptions{DisableSignup: true},
	)

	world := journey.New(t)

	// Organizations are provisioned on the shared signup-enabled server, then
	// sessions are opened against the disable-signup instance. Production
	// private instances are bootstrapped the same way: orgs exist before the
	// flag is flipped.
	alice := world.NewActor("Alice", testutil.RoleOwner)
	bob := world.NewMemberActor("Bob", testutil.RoleViewer, alice)

	visitor := world.NewUnauthenticatedActorFor("Visitor", env.BaseURL)
	visitorEmail := factory.SafeEmail()

	visitor.Step(
		"signs in with a magic link without joining an organization",
		func() error {
			visitor.Client().SignInWithMagicLink(visitorEmail)
			return nil
		},
	)

	visitor.Step(
		"cannot create an organization",
		func() error {
			err := createOrganizationShouldFail(
				t,
				visitor.Client(),
				factory.SafeName("Blocked organization"),
			)
			testutil.RequireMembershipRequiredError(t, err)

			return nil
		},
	)

	visitor.Step(
		"cannot use the console",
		func() error {
			err := consoleProbeShouldFail(t, visitor.Client())
			testutil.RequireMembershipRequiredError(t, err)

			return nil
		},
	)

	aliceOnPrivate := world.NewUnauthenticatedActorFor(
		"Alice on private instance",
		env.BaseURL,
	)

	aliceOnPrivate.Step(
		"signs in with a magic link on the private instance",
		func() error {
			aliceOnPrivate.Client().SignInWithMagicLink(alice.Client().GetEmail())
			return nil
		},
	)

	aliceOnPrivate.Step(
		"creates another organization",
		func() error {
			organizationID, err := createOrganizationAs(
				t,
				aliceOnPrivate.Client(),
				factory.SafeName("Second organization"),
			)
			if err != nil {
				return fmt.Errorf("cannot create organization: %w", err)
			}

			if organizationID == "" {
				return fmt.Errorf("create organization returned an empty ID")
			}

			if organizationID == alice.Client().GetOrganizationID().String() {
				return fmt.Errorf("expected a new organization ID, got the existing one")
			}

			return nil
		},
	)

	bobOnPrivate := world.NewUnauthenticatedActorFor(
		"Bob on private instance",
		env.BaseURL,
	)

	bobOnPrivate.Step(
		"signs in with a magic link as a non-owner member",
		func() error {
			bobOnPrivate.Client().SignInWithMagicLink(bob.Client().GetEmail())
			return nil
		},
	)

	bobOnPrivate.Step(
		"cannot create an organization without an owner role",
		func() error {
			err := createOrganizationShouldFail(
				t,
				bobOnPrivate.Client(),
				factory.SafeName("Viewer organization"),
			)
			testutil.RequireForbiddenError(t, err)

			return nil
		},
	)

	alice.Step(
		"deactivates Bob's membership",
		func() error {
			return deactivateUser(t, alice.Client(), bob.Client())
		},
	)

	bobOnPrivate.Step(
		"signs in with a magic link after deactivation",
		func() error {
			bobOnPrivate.Client().SignInWithMagicLink(bob.Client().GetEmail())
			return nil
		},
	)

	bobOnPrivate.Step(
		"cannot create an organization with a deactivated membership",
		func() error {
			err := createOrganizationShouldFail(
				t,
				bobOnPrivate.Client(),
				factory.SafeName("Deactivated organization"),
			)
			testutil.RequireMembershipRequiredError(t, err)

			return nil
		},
	)

	bobOnPrivate.Step(
		"cannot use the console with a deactivated membership",
		func() error {
			err := consoleProbeShouldFail(t, bobOnPrivate.Client())
			testutil.RequireMembershipRequiredError(t, err)

			return nil
		},
	)
}
