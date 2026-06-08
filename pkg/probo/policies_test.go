// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package probo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/probo"
)

func TestPublicFilePolicy(t *testing.T) {
	t.Parallel()

	evaluator := policy.NewEvaluator()
	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	anonID := coredata.AnonymousIdentityID

	t.Run("allows download-url for PUBLIC file", func(t *testing.T) {
		t.Parallel()

		req := policy.AuthorizationRequest{
			Principal: anonID,
			Resource:  fileID,
			Action:    probo.ActionFileDownloadUrl,
			ConditionContext: policy.ConditionContext{
				Principal: map[string]string{
					"id": anonID.String(),
				},
				Resource: map[string]string{
					"visibility": "PUBLIC",
				},
			},
		}

		result := evaluator.Evaluate(req, []*policy.Policy{probo.PublicFilePolicy})
		assert.True(t, result.IsAllowed(), "PUBLIC file should be allowed for anonymous")
	})

	t.Run("denies download-url for PRIVATE file", func(t *testing.T) {
		t.Parallel()

		req := policy.AuthorizationRequest{
			Principal: anonID,
			Resource:  fileID,
			Action:    probo.ActionFileDownloadUrl,
			ConditionContext: policy.ConditionContext{
				Principal: map[string]string{
					"id": anonID.String(),
				},
				Resource: map[string]string{
					"visibility": "PRIVATE",
				},
			},
		}

		result := evaluator.Evaluate(req, []*policy.Policy{probo.PublicFilePolicy})
		assert.False(t, result.IsAllowed(), "PRIVATE file should be denied for anonymous")
	})
}

func TestProboPolicySet_RegistersPublicFilePolicy(t *testing.T) {
	t.Parallel()

	ps := probo.ProboPolicySet()
	require.NotEmpty(t, ps.IdentityScopedPolicies, "ProboPolicySet should register at least one identity-scoped policy")
}

func TestAuditorPolicy_ProcessingActivityPageReadAccess(t *testing.T) {
	t.Parallel()

	organizationID := gid.New(gid.NewTenantID(), 1)
	evaluator := policy.NewEvaluator()
	conditionContext := policy.ConditionContext{
		Principal: map[string]string{
			"organization_id": organizationID.String(),
		},
		Resource: map[string]string{
			"organization_id": organizationID.String(),
		},
	}

	tests := []struct {
		name   string
		action string
	}{
		{
			name:   "list processing activities",
			action: probo.ActionProcessingActivityList,
		},
		{
			name:   "list data protection impact assessments",
			action: probo.ActionDataProtectionImpactAssessmentList,
		},
		{
			name:   "list transfer impact assessments",
			action: probo.ActionTransferImpactAssessmentList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := evaluator.Evaluate(
				policy.AuthorizationRequest{
					Principal:        organizationID,
					Resource:         organizationID,
					Action:           tt.action,
					ConditionContext: conditionContext,
				},
				[]*policy.Policy{probo.AuditorPolicy},
			)

			assert.True(t, result.IsAllowed())
		})
	}
}

func TestAuditorPolicy_OrganizationContextReadAccess(t *testing.T) {
	t.Parallel()

	organizationID := gid.New(gid.NewTenantID(), 1)
	evaluator := policy.NewEvaluator()
	conditionContext := policy.ConditionContext{
		Principal: map[string]string{
			"organization_id": organizationID.String(),
		},
		Resource: map[string]string{
			"organization_id": organizationID.String(),
		},
	}

	result := evaluator.Evaluate(
		policy.AuthorizationRequest{
			Principal:        organizationID,
			Resource:         organizationID,
			Action:           probo.ActionOrganizationContextGet,
			ConditionContext: conditionContext,
		},
		[]*policy.Policy{probo.AuditorPolicy},
	)

	assert.True(t, result.IsAllowed())
}
