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
	"testing"

	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestControlApplicability_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1FrameworkID := factory.CreateFramework(org1Owner)
	org1ControlID := factory.CreateControl(org1Owner, org1FrameworkID)
	org1SOAID := factory.NewStatementOfApplicability(org1Owner).
		WithName(factory.SafeName("Org1 SOA tenant")).
		Create()
	org1ASID := factory.CreateApplicabilityStatement(org1Owner, org1SOAID, org1ControlID, true, nil)
	org1ObligationID := createObligationLeaf(
		t,
		org1Owner,
		factory.SafeName("Org1 obligation tenant"),
		"LEGAL",
	)
	createControlObligationMappingLeaf(t, org1Owner, org1ControlID, org1ObligationID)

	org2SOAID := factory.NewStatementOfApplicability(org2Owner).
		WithName(factory.SafeName("Org2 SOA tenant")).
		Create()

	t.Run(
		"cannot create applicability statement for another organization's control",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := createApplicabilityStatementExpectError(
				t,
				org2,
				org2SOAID,
				org1ControlID,
				true,
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not create applicability for another organization's control",
			)
		},
	)

	t.Run(
		"cannot update applicability statement in another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := updateApplicabilityStatementExpectError(t, org2, org1ASID, false)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not update applicability statement from another organization",
			)
		},
	)

	t.Run(
		"cannot delete applicability statement in another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := deleteApplicabilityStatementExpectError(t, org2, org1ASID)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete applicability statement from another organization",
			)
		},
	)

	t.Run(
		"cannot create control obligation mapping on another organization's control",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)
			org2ObligationID := createObligationLeaf(
				t,
				org2,
				factory.SafeName("Org2 obligation tenant"),
				"LEGAL",
			)

			err := createControlObligationMappingExpectError(
				t,
				org2,
				org1ControlID,
				org2ObligationID,
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not map obligations onto another organization's control",
			)
		},
	)

	t.Run(
		"cannot delete control obligation mapping on another organization's control",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := deleteControlObligationMappingExpectError(
				t,
				org2,
				org1ControlID,
				org1ObligationID,
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete mapping on another organization's control",
			)
		},
	)
}
