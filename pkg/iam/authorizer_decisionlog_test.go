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

package iam_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"

	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

func TestAuthorizer_DecisionLogging(t *testing.T) {
	t.Parallel()

	t.Run("allow writes audit and decision log", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()

		var logOutput bytes.Buffer

		authorizer := newTestAuthorizerWithLogger(client, action, nil, &logOutput)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{fixture.frameworkID1},
			},
		)
		require.NoError(t, err)

		output := logOutput.String()
		assert.Contains(t, output, "authz decision")
		assert.Contains(t, output, "allow")
		assert.Contains(t, output, action)
		assert.Contains(t, output, fixture.identityID.String())
		assert.Contains(t, output, fixture.frameworkID1.String())
		assert.Contains(t, output, "allow-test-action")
		assert.Equal(t, 1, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("deny writes decision log without audit row", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()

		var logOutput bytes.Buffer

		authorizer := newTestAuthorizerWithLogger(
			client,
			action,
			nil,
			&logOutput,
			policy.Deny(action).WithSID("deny-test-action"),
		)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{fixture.frameworkID1},
			},
		)
		require.Error(t, err)

		output := logOutput.String()
		assert.Contains(t, output, "authz decision")
		assert.Contains(t, output, "deny")
		assert.Contains(t, output, "deny-test-action")
		assert.Contains(t, output, "explicit deny by statement deny-test-action")
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("implicit deny logs no_match without policy id", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()

		var logOutput bytes.Buffer

		authorizer := newTestAuthorizerWithLogger(
			client,
			"core:other:action",
			nil,
			&logOutput,
		)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{fixture.frameworkID1},
			},
		)
		require.Error(t, err)

		output := logOutput.String()
		assert.Contains(t, output, "authz decision")
		assert.Contains(t, output, "no_match")
		assert.NotContains(t, output, "policy_id")
		assert.True(t, strings.Contains(output, "implicit deny"))
	})
}

func newTestAuthorizerWithLogger(
	client *pg.Client,
	action string,
	allowResourceID *gid.GID,
	logOutput *bytes.Buffer,
	extraStatements ...policy.Statement,
) *iam.Authorizer {
	statements := []policy.Statement{
		policy.Allow(action).WithSID("allow-test-action"),
	}
	if allowResourceID != nil {
		statements[0] = statements[0].When(policy.Equals("resource.id", allowResourceID.String()))
	}

	statements = append(statements, extraStatements...)

	authorizer := iam.NewAuthorizer(client, log.NewLogger(log.WithOutput(logOutput)))
	authorizer.RegisterPolicySet(
		iam.NewPolicySet().AddRolePolicy(
			string(coredata.MembershipRoleOwner),
			policy.NewPolicy("batch-authorize-test", "Batch Authorize Test", statements...),
		),
	)

	return authorizer
}
