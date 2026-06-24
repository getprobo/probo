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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// TestSecurity_WriteGap_PublishRiskListApproverIDs covers a write-gap found
// while auditing GHSA-c74x-79w6-63jh's blast radius: GeneratedDocumentService's
// shared publishOrRequestApproval helper (used by every generated-document
// Publish* method: risk list, processing activity list, third party list,
// obligation list, finding list, data list, DPIA/TIA lists, asset list,
// statement of applicability, framework/audit report) persisted
// caller-supplied approverIds as DocumentDefaultApprovers and
// DocumentVersionApprovalDecision rows without validating they belong to the
// caller's own organization. This test exercises one call site
// (publishRiskList); the fix (validateApproverProfileIDs) lives in the single
// shared helper so it covers all of them.
func TestSecurity_WriteGap_PublishRiskListApproverIDs(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	factory.CreateRisk(org1Owner, factory.Attrs{"name": "Org1 Risk for publish"})

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Approver"})

	_, err := org1Owner.Do(`
		mutation($input: PublishRiskListInput!) {
			publishRiskList(input: $input) {
				documentEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": org1Owner.GetOrganizationID().String(),
			"approverIds":    []string{org2ProfileID},
			"minor":          false,
		},
	})
	require.Error(t, err, "must not accept an approverId belonging to another organization")
}

// TestSecurity_WriteGap_TrustCenterAccessDocuments covers a write-gap found
// while auditing GHSA-c74x-79w6-63jh's blast radius: TrustCenterAccessService.Update
// persisted caller-supplied document/report-file/trust-center-file ids into
// trust_center_document_accesses (via coredata's MergeDocumentAccesses/
// MergeReportFileAccesses/MergeTrustCenterFileAccesses) without validating
// they belong to the trust center's own organization -- the DB-level FK check
// alone doesn't catch this because those primary keys are globally unique,
// not per-tenant.
//
// TrustCenterAccess rows are normally created through the trust/v1 public
// portal's visitor request flow (requestAllAccesses), which needs a
// separate authenticated visitor identity and NDA acceptance. To keep this
// test focused on the fix under test (the Update mutation's FK validation)
// rather than that unrelated flow, the access row's prerequisite state is
// seeded directly via SQL against the same Postgres database the e2e probod
// instance runs against, then the real updateTrustCenterAccess mutation is
// exercised through the live GraphQL API.
func TestSecurity_WriteGap_TrustCenterAccessDocuments(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1TrustCenterID := trustCenterID(t, org1Owner)
	org1DocumentID := factory.NewDocument(org1Owner).WithTitle("Org1 Document for trust center access").Create()
	org2DocumentID := factory.NewDocument(org2Owner).WithTitle("Org2 Secret Document").Create()

	accessID := seedTrustCenterAccess(t, org1Owner, org1TrustCenterID)

	t.Run("cannot grant access to a document from another organization", func(t *testing.T) {
		_, err := org1Owner.Do(`
			mutation($input: UpdateTrustCenterAccessInput!) {
				updateTrustCenterAccess(input: $input) {
					trustCenterAccess { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":        accessID,
				"documents": []map[string]any{{"id": org2DocumentID, "status": "GRANTED"}},
			},
		})
		require.Error(t, err, "must not accept a documentId belonging to another organization")
	})

	t.Run("can grant access to a document from the same organization", func(t *testing.T) {
		_, err := org1Owner.Do(`
			mutation($input: UpdateTrustCenterAccessInput!) {
				updateTrustCenterAccess(input: $input) {
					trustCenterAccess { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":        accessID,
				"documents": []map[string]any{{"id": org1DocumentID, "status": "GRANTED"}},
			},
		})
		require.NoError(t, err)
	})
}

// seedTrustCenterAccess inserts a minimal trust_center_accesses row directly
// via SQL, bypassing the trust/v1 visitor request flow (which requires a
// separate authenticated visitor identity and NDA acceptance) so that
// updateTrustCenterAccess -- the mutation under test -- can be exercised in
// isolation. owner's own identity id is reused to satisfy the row's
// identity_id foreign key; which identity it is doesn't matter for this test.
func seedTrustCenterAccess(t *testing.T, owner *testutil.Client, trustCenterID string) string {
	t.Helper()

	tcID, err := gid.ParseGID(trustCenterID)
	require.NoError(t, err)

	tenantID := owner.GetOrganizationID().TenantID()
	accessID := gid.New(tenantID, coredata.TrustCenterAccessEntityType)
	now := time.Now().UTC()

	client := test.PGClient(t)
	ctx := context.Background()

	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, `
			INSERT INTO trust_center_accesses (id, tenant_id, organization_id, trust_center_id, identity_id, email, name, state, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'ACTIVE', $8, $8)
		`,
			accessID.String(), tenantID.String(), owner.GetOrganizationID().String(), tcID.String(),
			owner.GetUserID().String(), factory.SafeEmail(), "Test Access", now,
		)

		return err
	})
	require.NoError(t, err, "test setup: cannot seed trust_center_accesses row")

	return accessID.String()
}
