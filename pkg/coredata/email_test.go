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
	"go.probo.inc/probo/pkg/mail"
)

func TestResetStaleProcessingEmails_RequeuesOrFails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attemptCount  int
		startedAgo    time.Duration
		wantStatus    coredata.EmailStatus
		wantStarted   bool
		wantLastError bool
	}{
		{
			name:          "requeues while attempts remain",
			attemptCount:  3,
			startedAgo:    time.Hour,
			wantStatus:    coredata.EmailStatusPending,
			wantLastError: true,
		},
		{
			name:          "requeues on the last remaining attempt",
			attemptCount:  9,
			startedAgo:    time.Hour,
			wantStatus:    coredata.EmailStatusPending,
			wantLastError: true,
		},
		{
			name:          "fails after attempts are exhausted",
			attemptCount:  10,
			startedAgo:    time.Hour,
			wantStatus:    coredata.EmailStatusFailed,
			wantLastError: true,
		},
		{
			name:         "leaves a fresh lease alone",
			attemptCount: 10,
			startedAgo:   0,
			wantStatus:   coredata.EmailStatusProcessing,
			wantStarted:  true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				pgClient := test.PGClient(t)
				startedAt := time.Now().Add(-tt.startedAgo)
				email := insertProcessingEmailFixture(
					t,
					pgClient,
					tt.attemptCount,
					startedAt,
				)

				require.NoError(
					t,
					pgClient.WithConn(
						t.Context(),
						func(ctx context.Context, conn pg.Querier) error {
							return coredata.ResetStaleProcessingEmails(
								ctx,
								conn,
								10*time.Minute,
							)
						},
					),
				)

				persisted := loadEmailRecoveryState(t, pgClient, email.ID)
				assert.Equal(t, tt.wantStatus, persisted.Status)
				assert.Equal(t, tt.attemptCount, persisted.AttemptCount)
				assert.Equal(t, tt.wantStarted, persisted.ProcessingStartedAt != nil)
				assert.Equal(t, tt.wantLastError, persisted.LastError != nil)
			},
		)
	}
}

func insertProcessingEmailFixture(
	t *testing.T,
	pgClient *pg.Client,
	attemptCount int,
	startedAt time.Time,
) *coredata.Email {
	t.Helper()

	email := coredata.NewEmail(
		"Stale Recovery",
		mail.Addr("stale-recovery@example.test"),
		"Stale recovery test",
		"Stale recovery test",
		nil,
		nil,
	)

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if err := email.Insert(ctx, tx); err != nil {
					return err
				}

				email.Status = coredata.EmailStatusProcessing
				email.ProcessingStartedAt = &startedAt
				email.AttemptCount = attemptCount
				email.LastAttemptedAt = &startedAt

				return email.Update(ctx, tx)
			},
		),
	)
	t.Cleanup(
		func() {
			assert.NoError(
				t,
				pgClient.WithConn(
					context.Background(),
					func(ctx context.Context, conn pg.Querier) error {
						_, err := conn.Exec(
							ctx,
							"DELETE FROM emails WHERE id = $1",
							email.ID,
						)

						return err
					},
				),
			)
		},
	)

	return email
}

func loadEmailRecoveryState(
	t *testing.T,
	pgClient *pg.Client,
	id gid.GID,
) *coredata.Email {
	t.Helper()

	var email coredata.Email

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT status, attempt_count, processing_started_at, last_error
					FROM emails
					WHERE id = $1`,
					id,
				).Scan(
					&email.Status,
					&email.AttemptCount,
					&email.ProcessingStartedAt,
					&email.LastError,
				)
			},
		),
	)

	return &email
}
