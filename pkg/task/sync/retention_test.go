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

package tasksync

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRetentionHandlerDeletesExpiredLinearWebhookEvents(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Round(time.Microsecond)
	expiredProcessedAt := now.Add(-2 * time.Hour)
	keptProcessedAt := now.Add(-30 * time.Minute)
	expiredDeadLetteredAt := now.Add(-4 * time.Hour)

	expiredProcessed := coredata.NewLinearWebhookEvent("retention-processed-expired", []byte(`{}`))
	expiredProcessed.ProcessedAt = &expiredProcessedAt
	expiredProcessed.UpdatedAt = expiredProcessedAt

	keptProcessed := coredata.NewLinearWebhookEvent("retention-processed-kept", []byte(`{}`))
	keptProcessed.ProcessedAt = &keptProcessedAt
	keptProcessed.UpdatedAt = keptProcessedAt

	expiredDeadLetter := coredata.NewLinearWebhookEvent("retention-dead-letter-expired", []byte(`{}`))
	expiredDeadLetter.DeadLetteredAt = &expiredDeadLetteredAt
	expiredDeadLetter.UpdatedAt = expiredDeadLetteredAt

	pending := coredata.NewLinearWebhookEvent("retention-pending", []byte(`{}`))

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				for _, event := range []*coredata.LinearWebhookEvent{
					expiredProcessed,
					keptProcessed,
					expiredDeadLetter,
					pending,
				} {
					if _, err := event.Insert(ctx, tx); err != nil {
						return err
					}
				}

				return nil
			},
		),
	)

	h := &retentionHandler{
		pg:                  pgClient,
		logger:              log.NewLogger(log.WithOutput(io.Discard)),
		retention:           time.Hour,
		deadLetterRetention: 3 * time.Hour,
		batchSize:           10,
		now:                 func() time.Time { return now },
	}
	require.NoError(t, h.Run(t.Context()))

	var remaining []string

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				rows, err := conn.Query(
					ctx,
					`SELECT delivery_id
					FROM linear_webhook_events
					WHERE delivery_id IN (
						@expired_processed,
						@kept_processed,
						@expired_dead_letter,
						@pending
					)
					ORDER BY delivery_id`,
					pgx.StrictNamedArgs{
						"expired_processed":   expiredProcessed.DeliveryID,
						"kept_processed":      keptProcessed.DeliveryID,
						"expired_dead_letter": expiredDeadLetter.DeliveryID,
						"pending":             pending.DeliveryID,
					},
				)
				if err != nil {
					return err
				}

				remaining, err = pgx.CollectRows(rows, pgx.RowTo[string])

				return err
			},
		),
	)
	require.Equal(t, []string{pending.DeliveryID, keptProcessed.DeliveryID}, remaining)
}
