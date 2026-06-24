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

// Package test provides shared Postgres test fixtures for packages that
// exercise the database layer against a real Postgres instance. Centralizing
// the connection bootstrap and agent_runs schema setup here keeps a single
// copy of that logic so it cannot drift between the coredata and agentrun
// test suites.
package test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/migrator"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// pgURLEnvVar points the integration tests at a migrated test database.
	pgURLEnvVar = "PROBO_TEST_PG_URL"

	// defaultPGURL targets the local compose Postgres so tests run with zero
	// configuration against a developer's stack. When the database is not
	// reachable (e.g. in CI, where `make test` runs without Postgres) the
	// connection check fails and the test is skipped.
	defaultPGURL = "postgres://probod:probod@localhost:5432/probod_test"
)

var (
	sharedPGClient *pg.Client
	pgOnce         sync.Once
	pgInitErr      error
	migrateOnce    sync.Once
	migrateErr     error
)

// PGClient returns a process-wide shared pg.Client connected to the test
// database described by the PROBO_TEST_PG_URL environment variable (falling
// back to a local compose Postgres), applying the agent_runs migrations on
// first use. The test is skipped when no database is reachable so `make test`
// stays a pure unit-test run.
func PGClient(t *testing.T) *pg.Client {
	t.Helper()

	pgOnce.Do(
		func() {
			dsn := os.Getenv(pgURLEnvVar)
			if dsn == "" {
				dsn = defaultPGURL
			}

			u, err := url.Parse(dsn)
			if err != nil {
				pgInitErr = fmt.Errorf("cannot parse %s: %w", pgURLEnvVar, err)
				return
			}

			opts := []pg.Option{pg.WithPoolSize(25)}

			if u.Host != "" {
				host := u.Host
				if u.Port() == "" {
					host = net.JoinHostPort(u.Hostname(), "5432")
				}

				opts = append(opts, pg.WithAddr(host))
			}

			if u.User != nil {
				opts = append(opts, pg.WithUser(u.User.Username()))
				if password, ok := u.User.Password(); ok {
					opts = append(opts, pg.WithPassword(password))
				}
			}

			if len(u.Path) > 1 {
				opts = append(opts, pg.WithDatabase(u.Path[1:]))
			}

			sharedPGClient, pgInitErr = pg.NewClient(opts...)
			if pgInitErr != nil {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			pgInitErr = sharedPGClient.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, "SELECT 1")
					return err
				},
			)
		},
	)

	if pgInitErr != nil {
		t.Skipf("cannot connect to test database: %v", pgInitErr)
	}

	migrateSchema(t, sharedPGClient)

	return sharedPGClient
}

// migrateSchema applies the full coredata migration set to the shared test
// database. The migrator is idempotent (it records applied versions in
// schema_versions and serializes through an advisory lock), so running it
// once per process brings any reachable database up to date regardless of
// its starting state.
func migrateSchema(t *testing.T, client *pg.Client) {
	t.Helper()

	migrateOnce.Do(
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			logger := log.NewLogger(log.WithOutput(io.Discard))

			migrateErr = migrator.
				NewMigrator(client, coredata.Migrations, logger).
				Run(ctx, "migrations")
		},
	)

	require.NoError(t, migrateErr, "cannot migrate test database schema")
}
