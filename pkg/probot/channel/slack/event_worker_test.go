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

package slack

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	recordingEventProcessor struct {
		calls int
		errs  []error
	}

	eventWorkerFixture struct {
		pg *pg.Client
	}
)

func (p *recordingEventProcessor) ProcessEvent(context.Context, Envelope) error {
	p.calls++
	if len(p.errs) == 0 {
		return nil
	}

	err := p.errs[0]
	p.errs = p.errs[1:]

	return err
}

func newEventWorkerFixture(t *testing.T) eventWorkerFixture {
	t.Helper()

	slackEventTestMu.Lock()
	t.Cleanup(slackEventTestMu.Unlock)

	return eventWorkerFixture{pg: test.PGClient(t)}
}

func (f eventWorkerFixture) queue(t *testing.T, maxAttempts int) *coredata.SlackbotEvent {
	t.Helper()

	eventID := "E-worker-" + gid.New(gid.NilTenant, coredata.SlackbotEventEntityType).String()
	envelope, err := json.Marshal(
		Envelope{
			Type:    EnvelopeTypeEventCallback,
			EventID: eventID,
			Event:   &EventBody{Type: EventTypeMessage},
		},
	)
	require.NoError(t, err)

	event := coredata.NewSlackbotEvent(eventID, envelope)
	event.MaxAttempts = maxAttempts

	require.NoError(
		t,
		f.pg.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := event.Insert(ctx, conn)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = f.pg.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					return event.Delete(ctx, conn)
				},
			)
		},
	)

	return event
}

func (f eventWorkerFixture) load(t *testing.T, eventID string) coredata.SlackbotEvent {
	t.Helper()

	var event coredata.SlackbotEvent

	require.NoError(
		t,
		f.pg.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return event.LoadByEventID(ctx, conn, eventID)
			},
		),
	)

	return event
}

func newTestEventWorkerHandler(
	f eventWorkerFixture,
	processor EventProcessor,
) *eventWorkerHandler {
	return &eventWorkerHandler{
		pg:         f.pg,
		processor:  processor,
		logger:     log.NewLogger(log.WithName("test")),
		staleAfter: 5 * time.Minute,
		retryBase:  time.Second,
		retryMax:   time.Minute,
		now:        time.Now,
	}
}

func TestEventWorkerHandler_ProcessesPersistedEvent(t *testing.T) {
	t.Parallel()

	fixture := newEventWorkerFixture(t)
	queued := fixture.queue(t, 5)
	processor := &recordingEventProcessor{}
	handler := newTestEventWorkerHandler(fixture, processor)

	claimed, err := handler.Claim(t.Context())
	require.NoError(t, err)
	assert.Equal(t, queued.EventID, claimed.EventID)
	assert.Equal(t, 1, claimed.AttemptCount)
	require.NoError(t, handler.Process(t.Context(), claimed))

	processed := fixture.load(t, queued.EventID)
	assert.NotNil(t, processed.ProcessedAt)
	assert.Nil(t, processed.ProcessingStartedAt)
	assert.Nil(t, processed.LastError)
	assert.Equal(t, 1, processor.calls)
}

func TestEventWorkerHandler_RetriesTransientFailure(t *testing.T) {
	t.Parallel()

	fixture := newEventWorkerFixture(t)
	queued := fixture.queue(t, 3)
	processor := &recordingEventProcessor{errs: []error{errors.New("temporary failure")}}
	handler := newTestEventWorkerHandler(fixture, processor)
	now := time.Now().UTC()
	handler.now = func() time.Time { return now }

	claimed, err := handler.Claim(t.Context())
	require.NoError(t, err)
	require.Error(t, handler.Process(t.Context(), claimed))

	failed := fixture.load(t, queued.EventID)
	assert.Nil(t, failed.ProcessedAt)
	assert.NotNil(t, failed.LastError)
	require.NotNil(t, failed.NextAttemptAt)
	assert.True(t, failed.NextAttemptAt.Equal(now.Add(time.Second)))

	_, err = handler.Claim(t.Context())
	require.ErrorIs(t, err, worker.ErrNoTask)

	now = now.Add(time.Second)
	claimed, err = handler.Claim(t.Context())
	require.NoError(t, err)
	require.NoError(t, handler.Process(t.Context(), claimed))

	processed := fixture.load(t, queued.EventID)
	assert.NotNil(t, processed.ProcessedAt)
	assert.Equal(t, 2, processor.calls)
}

func TestEventWorkerHandler_DeadLettersPermanentFailure(t *testing.T) {
	t.Parallel()

	fixture := newEventWorkerFixture(t)
	queued := fixture.queue(t, 5)
	processor := &recordingEventProcessor{
		errs: []error{
			&permanentEventError{err: errors.New("malformed event")},
		},
	}
	handler := newTestEventWorkerHandler(fixture, processor)

	claimed, err := handler.Claim(t.Context())
	require.NoError(t, err)
	require.Error(t, handler.Process(t.Context(), claimed))

	failed := fixture.load(t, queued.EventID)
	assert.Nil(t, failed.ProcessedAt)
	assert.Nil(t, failed.NextAttemptAt)
	assert.NotNil(t, failed.LastError)
	assert.NotNil(t, failed.DeadLetteredAt)
	assert.Equal(t, failed.MaxAttempts, failed.AttemptCount)

	var deadLettered coredata.SlackbotEvent

	require.NoError(
		t,
		fixture.pg.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return deadLettered.LoadDeadLetteredByEventID(ctx, conn, queued.EventID)
			},
		),
	)
	assert.Equal(t, queued.EventID, deadLettered.EventID)

	_, err = handler.Claim(t.Context())
	require.ErrorIs(t, err, worker.ErrNoTask)
}

func TestEventWorkerHandler_RecoversStaleClaim(t *testing.T) {
	t.Parallel()

	fixture := newEventWorkerFixture(t)
	queued := fixture.queue(t, 3)
	processor := &recordingEventProcessor{}
	handler := newTestEventWorkerHandler(fixture, processor)
	claimTime := time.Now().UTC().Add(-10 * time.Minute)
	handler.now = func() time.Time { return claimTime }

	_, err := handler.Claim(t.Context())
	require.NoError(t, err)

	recoveryTime := claimTime.Add(10 * time.Minute)
	handler.now = func() time.Time { return recoveryTime }
	require.NoError(t, handler.RecoverStale(t.Context()))

	recovered := fixture.load(t, queued.EventID)
	assert.Nil(t, recovered.ProcessingStartedAt)
	require.NotNil(t, recovered.NextAttemptAt)
	assert.True(t, recovered.NextAttemptAt.Equal(recoveryTime))
	assert.NotNil(t, recovered.LastError)
}

func TestService_ProcessEvent_DisablesUninstalledWorkspace(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	installations := newTestInstallationService(t, pgClient, "")
	handler := &Service{installations: installations}
	envelope := Envelope{
		Type:    EnvelopeTypeEventCallback,
		EventID: "E-uninstall-" + organizationID.String(),
		TeamID:  teamID,
		Event:   &EventBody{Type: EventTypeAppUninstalled},
	}

	require.NoError(t, handler.ProcessEvent(t.Context(), envelope))
	loaded, err := installations.GetByOrganizationID(
		t.Context(),
		coredata.NewScopeFromObjectID(organizationID),
		organizationID,
	)
	require.NoError(t, err)
	assert.Equal(t, coredata.SlackbotInstallationStatusDisabled, loaded.Status)

	require.NoError(t, handler.ProcessEvent(t.Context(), envelope))
	loaded, err = installations.GetByOrganizationID(
		t.Context(),
		coredata.NewScopeFromObjectID(organizationID),
		organizationID,
	)
	require.NoError(t, err)
	assert.Equal(t, coredata.SlackbotInstallationStatusDisabled, loaded.Status)
}
