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

package probot

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type retentionHandler struct {
	pg                        *pg.Client
	logger                    *log.Logger
	retention                 time.Duration
	deadLetterRetention       time.Duration
	operationReceiptRetention time.Duration
	batchSize                 int
	now                       func() time.Time
}

const (
	DefaultReliabilityRetention      = 30 * 24 * time.Hour
	DefaultDeadLetterRetention       = 90 * 24 * time.Hour
	DefaultOperationReceiptRetention = 120 * 24 * time.Hour
	DefaultReliabilityCleanupBatch   = 1000
	DefaultReliabilityCleanupPeriod  = time.Hour
)

func NewRetentionWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.PeriodicWorker {
	workerOpts := append(
		[]worker.Option{worker.WithInterval(DefaultReliabilityCleanupPeriod)},
		opts...,
	)

	return worker.NewPeriodic(
		"probot-reliability-retention-worker",
		&retentionHandler{
			pg:                        pgClient,
			logger:                    logger,
			retention:                 DefaultReliabilityRetention,
			deadLetterRetention:       DefaultDeadLetterRetention,
			operationReceiptRetention: DefaultOperationReceiptRetention,
			batchSize:                 DefaultReliabilityCleanupBatch,
			now:                       time.Now,
		},
		logger,
		workerOpts...,
	)
}

type retentionCleanup struct {
	name string
	err  string
	run  func(ctx context.Context, tx pg.Tx) (int64, error)
}

func (h *retentionHandler) Run(ctx context.Context) error {
	before := h.now().Add(-h.retention)
	deadLetterBefore := h.now().Add(-h.deadLetterRetention)
	receiptBefore := h.now().Add(-h.operationReceiptRetention)
	cleanups := []retentionCleanup{
		{
			name: "event_inbox",
			err:  "cannot clean Slack event inbox",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteProcessedSlackbotEventsBeforeBatch(
					ctx,
					tx,
					before,
					h.batchSize,
				)
			},
		},
		{
			name: "event_dead_letters",
			err:  "cannot clean Slack event dead letters",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteDeadLetteredSlackbotEventsBeforeBatch(
					ctx,
					tx,
					deadLetterBefore,
					h.batchSize,
				)
			},
		},
		{
			name: "event_dedupe",
			err:  "cannot clean Slack event dedupe ledger",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteSlackbotProcessedEventsBeforeBatch(
					ctx,
					tx,
					before,
					h.batchSize,
				)
			},
		},
		{
			name: "command_inbox",
			err:  "cannot clean Slack command inbox",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteCompletedSlackbotInteractiveCommandsBefore(
					ctx,
					tx,
					before,
					h.batchSize,
				)
			},
		},
		{
			name: "command_dead_letters",
			err:  "cannot clean Slack command dead letters",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteDeadLetteredSlackbotInteractiveCommandsBefore(
					ctx,
					tx,
					deadLetterBefore,
					h.batchSize,
				)
			},
		},
		{
			name: "notification_messages",
			err:  "cannot clean Slack notification messages",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteRetiredSlackbotMessageThreadsBefore(
					ctx,
					tx,
					before,
					deadLetterBefore,
					h.batchSize,
				)
			},
		},
		{
			name: "delivery_operations",
			err:  "cannot clean Slack delivery operations",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteCompletedSlackDeliveryOperationsBefore(
					ctx,
					tx,
					before,
					h.batchSize,
				)
			},
		},
		{
			name: "delivery_dead_letters",
			err:  "cannot clean Slack delivery dead letters",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteDeadLetteredSlackDeliveryOperationsBefore(
					ctx,
					tx,
					deadLetterBefore,
					h.batchSize,
				)
			},
		},
		{
			name: "agent_inputs",
			err:  "cannot clean agent inputs",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteRetiredAgentInputsBefore(
					ctx,
					tx,
					before,
					deadLetterBefore,
					h.batchSize,
				)
			},
		},
		{
			name: "bot_messages",
			err:  "cannot clean bot messages",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteProcessedBotMessagesBeforeBatch(
					ctx,
					tx,
					before,
					h.batchSize,
				)
			},
		},
		{
			name: "bot_message_dead_letters",
			err:  "cannot clean bot message dead letters",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteDeadLetteredBotMessagesBeforeBatch(
					ctx,
					tx,
					deadLetterBefore,
					h.batchSize,
				)
			},
		},
		{
			name: "agent_executions",
			err:  "cannot clean agent executions and anchors",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				// Agent executions own their provider anchors.
				// Deleting an idle/dead execution cascades its anchors
				// and remaining inputs.
				return coredata.DeleteRetiredAgentExecutionsBefore(
					ctx,
					tx,
					deadLetterBefore,
					deadLetterBefore,
					h.batchSize,
				)
			},
		},
		{
			name: "operation_receipts",
			err:  "cannot clean operation receipts",
			run: func(ctx context.Context, tx pg.Tx) (int64, error) {
				return coredata.DeleteOperationReceiptsBefore(
					ctx,
					tx,
					receiptBefore,
					h.batchSize,
				)
			},
		},
	}
	counts := make(map[string]int64, len(cleanups))

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, cleanup := range cleanups {
				count, err := cleanup.run(ctx, tx)
				if err != nil {
					return fmt.Errorf("%s: %w", cleanup.err, err)
				}

				counts[cleanup.name] = count
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot run Probot reliability retention: %w", err)
	}

	for queue, count := range counts {
		if count > 0 {
			h.logger.InfoCtx(
				ctx,
				"deleted expired Probot reliability records",
				log.String("queue", queue),
				log.Int64("count", count),
			)
		}
	}

	return nil
}
