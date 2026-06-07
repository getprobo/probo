// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package coredata_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

func insertPendingRun(
	t *testing.T,
	client *pg.Client,
	agentName string,
	inputMessages []llm.Message,
) coredata.AgentRun {
	t.Helper()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	runID := gid.New(tenantID, coredata.AgentRunEntityType)
	inputJSON, err := json.Marshal(inputMessages)
	require.NoError(t, err)

	now := time.Now()
	run := coredata.AgentRun{
		ID:             runID,
		OrganizationID: orgID,
		StartAgentName: agentName,
		Status:         coredata.AgentRunStatusPending,
		InputMessages:  inputJSON,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = client.WithTx(
		context.Background(),
		func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(
				ctx,
				`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
				orgID.String(), tenantID.String(), "test-org-"+orgID.String(), now, now,
			); err != nil {
				return fmt.Errorf("cannot insert placeholder organization: %w", err)
			}

			return run.Insert(ctx, tx, coredata.NewScope(tenantID))
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", orgID.String())
			return err
		})
	})

	return run
}
