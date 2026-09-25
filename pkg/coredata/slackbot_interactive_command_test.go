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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestResetStaleSlackbotInteractiveCommands_RequeuesWhenAttemptsRemain(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	command := insertStaleSlackbotInteractiveCommand(t, pgClient, now, 1)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return coredata.ResetStaleSlackbotInteractiveCommands(
					ctx,
					conn,
					now,
					10*time.Minute,
				)
			},
		),
	)

	persisted := loadSlackbotInteractiveCommand(t, pgClient, command.RequestDigest)
	assert.Nil(t, persisted.ProcessingStartedAt)
	require.NotNil(t, persisted.NextAttemptAt)
	assert.WithinDuration(t, now, *persisted.NextAttemptAt, time.Second)
	assert.Nil(t, persisted.DeadLetteredAt)
	require.NotNil(t, persisted.LastError)
	assert.Equal(t, "Slack interactive command processing lease expired", *persisted.LastError)
}

func TestResetStaleSlackbotInteractiveCommands_DeadLettersWhenAttemptsExhausted(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	command := insertStaleSlackbotInteractiveCommand(
		t,
		pgClient,
		now,
		coredata.SlackbotInteractiveCommandDefaultMaxAttempts,
	)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return coredata.ResetStaleSlackbotInteractiveCommands(
					ctx,
					conn,
					now,
					10*time.Minute,
				)
			},
		),
	)

	persisted := loadSlackbotInteractiveCommand(t, pgClient, command.RequestDigest)
	assert.Nil(t, persisted.ProcessingStartedAt)
	assert.Nil(t, persisted.NextAttemptAt)
	require.NotNil(t, persisted.DeadLetteredAt)
	assert.WithinDuration(t, now, *persisted.DeadLetteredAt, time.Second)
	require.NotNil(t, persisted.LastError)
	assert.Equal(t, "Slack interactive command processing lease expired", *persisted.LastError)
}

func insertStaleSlackbotInteractiveCommand(
	t *testing.T,
	pgClient *pg.Client,
	now time.Time,
	attemptCount int,
) *coredata.SlackbotInteractiveCommand {
	t.Helper()

	digest := []byte(gid.New(gid.NilTenant, coredata.SlackbotInteractiveCommandEntityType).String())
	startedAt := now.Add(-time.Hour)
	command := coredata.NewSlackbotInteractiveCommand(digest, []byte("payload"))
	command.ProcessingStartedAt = &startedAt
	command.AttemptCount = attemptCount

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := command.Insert(ctx, conn)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(
						ctx,
						`DELETE FROM slackbot_interactive_commands WHERE id = $1`,
						command.ID,
					)

					return err
				},
			)
		},
	)

	return command
}

func loadSlackbotInteractiveCommand(
	t *testing.T,
	pgClient *pg.Client,
	requestDigest []byte,
) *coredata.SlackbotInteractiveCommand {
	t.Helper()

	var persisted coredata.SlackbotInteractiveCommand

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return persisted.LoadByRequestDigest(ctx, conn, requestDigest)
			},
		),
	)

	return &persisted
}
