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
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

var (
	sharedPGClientCoredata  *pg.Client
	pgOnceCoredata          sync.Once
	pgInitErrCoredata       error
	ensureTableOnceCoredata sync.Once
	ensureTableErrCoredata  error
)

func pgClient(t *testing.T) *pg.Client {
	t.Helper()

	pgOnceCoredata.Do(func() {
		addr := os.Getenv("PROBO_TEST_PG_ADDR")
		if addr == "" {
			addr = "localhost:5432"
		}

		user := os.Getenv("PROBO_TEST_PG_USER")
		if user == "" {
			user = "probod"
		}

		password := os.Getenv("PROBO_TEST_PG_PASSWORD")
		if password == "" {
			password = "probod"
		}

		database := os.Getenv("PROBO_TEST_PG_DATABASE")
		if database == "" {
			database = "probod_test"
		}

		sharedPGClientCoredata, pgInitErrCoredata = pg.NewClient(
			pg.WithAddr(addr),
			pg.WithUser(user),
			pg.WithPassword(password),
			pg.WithDatabase(database),
			pg.WithPoolSize(5),
		)
		if pgInitErrCoredata != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pgInitErrCoredata = sharedPGClientCoredata.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(ctx, "SELECT 1")
			return err
		})
	})

	if pgInitErrCoredata != nil {
		t.Skipf("cannot connect to test database: %v", pgInitErrCoredata)
	}

	ensureAgentRunsTable(t, sharedPGClientCoredata)

	return sharedPGClientCoredata
}

func ensureAgentRunsTable(t *testing.T, client *pg.Client) {
	t.Helper()

	ensureTableOnceCoredata.Do(func() {
		ctx := context.Background()
		ensureTableErrCoredata = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			var exists bool
			if err := conn.QueryRow(
				ctx,
				`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'agent_runs')`,
			).Scan(&exists); err != nil {
				return fmt.Errorf("cannot check agent_runs existence: %w", err)
			}

			if !exists {
				ddl, err := coredata.Migrations.ReadFile("migrations/20260424T173529Z.sql")
				if err != nil {
					return fmt.Errorf("cannot read agent_runs base migration: %w", err)
				}

				if _, err := conn.Exec(ctx, string(ddl)); err != nil {
					return fmt.Errorf("cannot apply agent_runs base migration: %w", err)
				}
			}

			var hasLeaseGeneration bool
			if err := conn.QueryRow(
				ctx,
				`SELECT EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_name = 'agent_runs'
						AND column_name = 'lease_generation'
				)`,
			).Scan(&hasLeaseGeneration); err != nil {
				return fmt.Errorf("cannot check lease_generation column: %w", err)
			}

			if !hasLeaseGeneration {
				ddl, err := coredata.Migrations.ReadFile("migrations/20260607T060000Z.sql")
				if err != nil {
					return fmt.Errorf("cannot read agent_runs lease generation migration: %w", err)
				}

				if _, err := conn.Exec(ctx, string(ddl)); err != nil {
					return fmt.Errorf("cannot apply agent_runs lease generation migration: %w", err)
				}
			}

			return nil
		})
	})

	require.NoError(t, ensureTableErrCoredata, "cannot ensure agent_runs table")
}

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
