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

// Package agentruntest provides shared test helpers for agent run
// integration tests that require a PostgreSQL database.
package agentruntest

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
	sharedPGClient  *pg.Client
	pgOnce          sync.Once
	pgInitErr       error
	ensureTableOnce sync.Once
	ensureTableErr  error
)

// PGClient returns a shared pg.Client connected to the test database.
// Skips the test if the database is not reachable.
func PGClient(t *testing.T) *pg.Client {
	t.Helper()

	pgOnce.Do(func() {
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

		sharedPGClient, pgInitErr = pg.NewClient(
			pg.WithAddr(addr),
			pg.WithUser(user),
			pg.WithPassword(password),
			pg.WithDatabase(database),
			pg.WithPoolSize(5),
		)
		if pgInitErr != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pgInitErr = sharedPGClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(ctx, "SELECT 1")
			return err
		})
	})

	if pgInitErr != nil {
		t.Skipf("cannot connect to test database: %v", pgInitErr)
	}

	EnsureAgentRunsTable(t, sharedPGClient)

	return sharedPGClient
}

// EnsureAgentRunsTable creates the agent_runs table if it does not
// exist, using the embedded migration.
func EnsureAgentRunsTable(t *testing.T, client *pg.Client) {
	t.Helper()

	ensureTableOnce.Do(func() {
		ctx := context.Background()
		ensureTableErr = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			var exists bool
			err := conn.QueryRow(
				ctx,
				`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'agent_runs')`,
			).Scan(&exists)
			if err != nil {
				return err
			}
			if exists {
				return nil
			}

			// Read from the embedded migration to avoid schema drift.
			ddl, err := coredata.Migrations.ReadFile("migrations/20260424T120000Z.sql")
			if err != nil {
				return fmt.Errorf("cannot read agent_runs migration: %w", err)
			}

			_, err = conn.Exec(ctx, string(ddl))
			return err
		})
	})
	require.NoError(t, ensureTableErr, "cannot ensure agent_runs table")
}

// CleanupAgentRun deletes an agent run by ID. Safe to call from
// t.Cleanup.
func CleanupAgentRun(client *pg.Client, id gid.GID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, "DELETE FROM agent_runs WHERE id = $1", id.String())
		return err
	})
}

// InsertPendingRun inserts a PENDING agent run and registers cleanup.
func InsertPendingRun(
	t *testing.T,
	client *pg.Client,
	agentName string,
	inputMessages []llm.Message,
) coredata.AgentRun {
	t.Helper()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, 1)
	runID := gid.New(tenantID, 2)

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
			return run.Insert(ctx, tx, coredata.NewScope(tenantID))
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		CleanupAgentRun(client, run.ID)
	})

	return run
}

// LoadAgentRun loads an agent run by ID, failing the test on error.
func LoadAgentRun(t *testing.T, client *pg.Client, id gid.GID) coredata.AgentRun {
	t.Helper()

	var run coredata.AgentRun
	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return run.LoadByID(ctx, conn, coredata.NewNoScope(), id)
		},
	)
	if err != nil {
		t.Fatalf("cannot load agent run %s: %v", id, err)
	}

	return run
}

// TryLoadAgentRun is a non-fatal variant safe for use inside
// require.Eventually callbacks (which recover panics).
func TryLoadAgentRun(client *pg.Client, id gid.GID) (coredata.AgentRun, error) {
	var run coredata.AgentRun
	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return run.LoadByID(ctx, conn, coredata.NewNoScope(), id)
		},
	)
	return run, err
}
