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

package management_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/employeeportal/management"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

func TestPolicySet_CompliancePortalRolesCanOpenEmployeePortal(t *testing.T) {
	t.Parallel()

	organizationID := gid.New(gid.NewTenantID(), 1)
	portalID := gid.New(gid.NewTenantID(), 2)
	evaluator := policy.NewEvaluator()
	conditionContext := policy.ConditionContext{
		Principal: map[string]string{
			"organization_id": organizationID.String(),
		},
		Resource: map[string]string{
			"organization_id": organizationID.String(),
		},
	}

	roles := []string{
		"COMPLIANCE_PORTAL_MANAGER",
		"COMPLIANCE_PORTAL_ACCESS_MANAGER",
	}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			t.Parallel()

			policies := management.PolicySet().RolePolicies[role]
			result := evaluator.Evaluate(
				policy.AuthorizationRequest{
					Principal:        organizationID,
					Resource:         portalID,
					Action:           management.ActionEmployeePortalGet,
					ConditionContext: conditionContext,
				},
				policies,
			)

			assert.True(t, result.IsAllowed())
		})
	}
}
