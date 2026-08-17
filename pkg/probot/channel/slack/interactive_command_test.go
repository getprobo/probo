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

package slack_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

func TestInteractiveCommandInbox_RequeuesDeadLetteredDigest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	client := test.PGClient(t)
	if !interactiveCommandsAvailable(t, client) {
		t.Skip("slackbot_interactive_commands is unavailable in the test database")
	}

	key := cipher.EncryptionKey{1, 2, 3}
	inbox := slackchannel.NewInteractiveCommandInbox(client, key)
	payload := uniqueInteractivePayload(t)
	digest := sha256.Sum256(payload)

	t.Cleanup(func() { deleteInteractiveCommand(t, client, digest[:]) })

	inserted, err := inbox.Enqueue(ctx, payload)
	require.NoError(t, err)
	assert.True(t, inserted)

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				now := time.Now()
				_, err := conn.Exec(
					ctx,
					`UPDATE slackbot_interactive_commands
					 SET dead_lettered_at = $1,
					     attempt_count = max_attempts,
					     last_error = 'permanent',
					     updated_at = $1
					 WHERE request_digest = $2`,
					now,
					digest[:],
				)

				return err
			},
		),
	)

	requeued, err := inbox.Enqueue(ctx, payload)
	require.NoError(t, err)
	assert.True(t, requeued)

	command := loadInteractiveCommand(t, client, digest[:])
	assert.Nil(t, command.DeadLetteredAt)
	assert.Nil(t, command.ProcessedAt)
	assert.Nil(t, command.ProcessingStartedAt)
	assert.Nil(t, command.NextAttemptAt)
	assert.Equal(t, 0, command.AttemptCount)
}

func TestInteractiveCommandInbox_DoesNotResetProcessedDigest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	client := test.PGClient(t)
	if !interactiveCommandsAvailable(t, client) {
		t.Skip("slackbot_interactive_commands is unavailable in the test database")
	}

	key := cipher.EncryptionKey{1, 2, 3}
	inbox := slackchannel.NewInteractiveCommandInbox(client, key)
	payload := uniqueInteractivePayload(t)
	digest := sha256.Sum256(payload)

	t.Cleanup(func() { deleteInteractiveCommand(t, client, digest[:]) })

	inserted, err := inbox.Enqueue(ctx, payload)
	require.NoError(t, err)
	assert.True(t, inserted)

	processedAt := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`UPDATE slackbot_interactive_commands
					 SET processed_at = $1,
					     processing_started_at = NULL,
					     updated_at = $1
					 WHERE request_digest = $2`,
					processedAt,
					digest[:],
				)

				return err
			},
		),
	)

	replayed, err := inbox.Enqueue(ctx, payload)
	require.NoError(t, err)
	assert.False(t, replayed)

	command := loadInteractiveCommand(t, client, digest[:])
	require.NotNil(t, command.ProcessedAt)
	assert.WithinDuration(t, processedAt, *command.ProcessedAt, time.Second)
	assert.Nil(t, command.DeadLetteredAt)
}

func uniqueInteractivePayload(t *testing.T) []byte {
	t.Helper()

	return fmt.Appendf(
		nil,
		`{"team":{"id":"T-%s"},"user":{"id":"U1"}}`,
		gid.New(gid.NilTenant, coredata.SlackbotInteractiveCommandEntityType),
	)
}

func loadInteractiveCommand(
	t *testing.T,
	client *pg.Client,
	digest []byte,
) coredata.SlackbotInteractiveCommand {
	t.Helper()

	var command coredata.SlackbotInteractiveCommand

	require.NoError(
		t,
		client.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				return command.LoadByRequestDigest(ctx, conn, digest)
			},
		),
	)

	return command
}

func deleteInteractiveCommand(t *testing.T, client *pg.Client, digest []byte) {
	t.Helper()

	_ = client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`DELETE FROM slackbot_interactive_commands WHERE request_digest = $1`,
				digest,
			)

			return err
		},
	)
}

func interactiveCommandsAvailable(t *testing.T, client *pg.Client) bool {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			var command coredata.SlackbotInteractiveCommand

			return command.LoadByRequestDigest(ctx, conn, []byte("missing"))
		},
	)

	return err == nil || errors.Is(err, coredata.ErrResourceNotFound)
}
