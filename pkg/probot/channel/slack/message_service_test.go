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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	legacyslack "go.probo.inc/probo/pkg/slack"
)

type (
	recordingModernQueue struct {
		requests []QueueNotificationRequest
	}

	recordingLegacyQueue struct {
		requests []legacyslack.QueueRequest
	}

	queueRouteCase struct {
		name             string
		destination      bool
		verified         bool
		modern           bool
		legacy           bool
		requireVerified  bool
		incompleteTarget bool
		wantBackend      string
		wantErr          error
	}
)

func (q *recordingModernQueue) Queue(
	_ context.Context,
	_ coredata.Scoper,
	req QueueNotificationRequest,
) (*coredata.SlackbotMessage, error) {
	q.requests = append(q.requests, req)

	return &coredata.SlackbotMessage{
		ID:             req.ID,
		OrganizationID: req.OrganizationID,
		ChannelID:      new(req.ChannelID),
	}, nil
}

func (q *recordingModernQueue) QueueRevision(
	context.Context,
	coredata.Scoper,
	gid.GID,
	map[string]any,
	map[string]any,
) (*coredata.SlackbotMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingModernQueue) GetByID(
	context.Context,
	coredata.Scoper,
	gid.GID,
) (*coredata.SlackbotMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingModernQueue) GetBySourceEventID(
	context.Context,
	coredata.Scoper,
	string,
) (*coredata.SlackbotMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingModernQueue) GetInitialByOrganizationIDChannelAndTS(
	context.Context,
	coredata.Scoper,
	gid.GID,
	string,
	string,
) (*coredata.SlackbotMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingLegacyQueue) Queue(
	_ context.Context,
	_ coredata.Scoper,
	req legacyslack.QueueRequest,
) (*coredata.SlackMessage, error) {
	q.requests = append(q.requests, req)

	return &coredata.SlackMessage{
		ID:             req.ID,
		OrganizationID: req.OrganizationID,
	}, nil
}

func (q *recordingLegacyQueue) GetByID(
	context.Context,
	coredata.Scoper,
	gid.GID,
) (*coredata.SlackMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingLegacyQueue) GetBySourceEventID(
	context.Context,
	coredata.Scoper,
	string,
) (*coredata.SlackMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingLegacyQueue) GetInitialByOrganizationIDChannelAndTS(
	context.Context,
	coredata.Scoper,
	gid.GID,
	string,
	string,
) (*coredata.SlackMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingLegacyQueue) GetInitialByChannelAndTS(
	context.Context,
	coredata.Scoper,
	string,
	string,
) (*coredata.SlackMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func (q *recordingLegacyQueue) UpdateViaResponseURL(
	context.Context,
	coredata.Scoper,
	gid.GID,
	string,
	map[string]any,
	map[string]any,
) (*coredata.SlackMessage, error) {
	return nil, coredata.ErrResourceNotFound
}

func TestMessageService_QueueRoutesModernAndLegacy(t *testing.T) {
	t.Parallel()

	cases := []queueRouteCase{
		{
			name:            "verified destination uses modern queue",
			destination:     true,
			verified:        true,
			modern:          true,
			legacy:          true,
			requireVerified: true,
			wantBackend:     BackendSlackbot,
		},
		{
			name:            "unverified destination without legacy is rejected",
			destination:     true,
			verified:        false,
			modern:          true,
			requireVerified: true,
			wantErr:         ErrNoDeliveryDestination,
		},
		{
			name:            "unverified destination falls back to legacy",
			destination:     true,
			verified:        false,
			modern:          true,
			legacy:          true,
			requireVerified: true,
			wantBackend:     BackendLegacySlack,
		},
		{
			name:            "missing destination falls back to legacy",
			modern:          true,
			legacy:          true,
			requireVerified: true,
			wantBackend:     BackendLegacySlack,
		},
		{
			name:            "missing destination without legacy is rejected",
			modern:          true,
			requireVerified: true,
			wantErr:         ErrSlackbotChannelNotFound,
		},
		{
			name:            "verified destination without modern uses legacy",
			destination:     true,
			verified:        true,
			legacy:          true,
			requireVerified: true,
			wantBackend:     BackendLegacySlack,
		},
		{
			name:             "incomplete delivery target is rejected",
			modern:           true,
			legacy:           true,
			requireVerified:  true,
			incompleteTarget: true,
			wantErr:          errors.New("delivery target is incomplete"),
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				ctx := t.Context()
				pgClient, scope, organizationID := newQueueTestOrganization(t)
				modern := &recordingModernQueue{}
				legacy := &recordingLegacyQueue{}
				service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
				if tc.modern {
					service.modern = modern
				}
				if tc.legacy {
					service.legacy = legacy
				}

				target := probot.DeliveryTarget{
					Namespace: "compliance_portal",
					Key:       organizationID.String(),
				}
				if tc.incompleteTarget {
					target = probot.DeliveryTarget{}
				} else if tc.destination {
					insertQueueDestination(t, pgClient, scope, organizationID, target, tc.verified)
				}

				err := service.queue(
					ctx,
					probot.OutboundDelivery{
						OrganizationID: organizationID,
						Capability:     "compliance_access",
						MessageType:    "ACCESS_REQUEST",
						Result: probot.OutboundMessage{
							Message: probot.Message{
								ID:             gid.New(scope.GetTenantID(), coredata.SlackbotMessageEntityType),
								OrganizationID: organizationID,
								Type:           "ACCESS_REQUEST",
							},
							Intent:         probot.MessageIntent{FallbackText: "hello"},
							DeliveryTarget: target,
						},
					},
					tc.requireVerified,
				)

				if tc.wantErr != nil {
					require.Error(t, err)
					if errors.Is(tc.wantErr, ErrNoDeliveryDestination) ||
						errors.Is(tc.wantErr, ErrSlackbotChannelNotFound) {
						assert.ErrorIs(t, err, tc.wantErr)
					} else {
						assert.ErrorContains(t, err, tc.wantErr.Error())
					}
					assert.Empty(t, modern.requests)
					assert.Empty(t, legacy.requests)

					return
				}

				require.NoError(t, err)

				switch tc.wantBackend {
				case BackendSlackbot:
					require.Len(t, modern.requests, 1)
					assert.Empty(t, legacy.requests)
					assert.Equal(t, "C-modern", modern.requests[0].ChannelID)
				case BackendLegacySlack:
					require.Len(t, legacy.requests, 1)
					assert.Empty(t, modern.requests)
				default:
					t.Fatalf("unexpected backend %q", tc.wantBackend)
				}
			},
		)
	}
}

func newQueueTestOrganization(t *testing.T) (*pg.Client, coredata.Scoper, gid.GID) {
	t.Helper()

	ctx := t.Context()
	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var destination coredata.BotDeliveryDestination

			return destination.LoadByTarget(
				ctx,
				conn,
				scope,
				ProviderName,
				"missing",
				"missing",
			)
		},
	)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		t.Skipf("bot_delivery_destinations is unavailable in the test database: %v", err)
	}

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				organization := coredata.Organization{
					ID:        organizationID,
					TenantID:  tenantID,
					Name:      "slack-queue-" + organizationID.String(),
					CreatedAt: now,
					UpdatedAt: now,
				}

				return organization.Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					return (&coredata.Organization{}).Delete(ctx, tx, organizationID)
				},
			)
		},
	)

	return pgClient, scope, organizationID
}

func insertQueueDestination(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
	verified bool,
) {
	t.Helper()

	destination := coredata.NewBotDeliveryDestination(
		scope,
		organizationID,
		ProviderName,
		target.Namespace,
		target.Key,
	)
	destination.ExternalDestinationID = "C-modern"
	if verified {
		now := time.Now()
		destination.VerifiedAt = &now
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := destination.Upsert(ctx, tx, scope)

				return err
			},
		),
	)
}
