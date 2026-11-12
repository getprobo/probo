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

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/bot"
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
							Message: bot.Message{
								ID:             gid.New(scope.GetTenantID(), coredata.SlackbotMessageEntityType),
								OrganizationID: organizationID,
								Type:           "ACCESS_REQUEST",
							},
							Intent:         bot.MessageIntent{FallbackText: "hello"},
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

	return newQueueTestOrganizationInTenant(t, gid.NewTenantID())
}

func newQueueTestOrganizationInTenant(
	t *testing.T,
	tenantID gid.TenantID,
) (*pg.Client, coredata.Scoper, gid.GID) {
	t.Helper()

	ctx := t.Context()
	pgClient := test.PGClient(t)
	scope := coredata.NewScope(tenantID, gid.Nil)
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
				organizationID,
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

func TestMessageService_GetMessageNilBackend(t *testing.T) {
	t.Parallel()

	service := NewMessageService(nil, nil, nil, nil, log.NewLogger())
	scope := coredata.NewScope(gid.NewTenantID(), gid.Nil)

	_, err := service.GetMessage(
		t.Context(),
		scope,
		gid.New(scope.GetTenantID(), coredata.SlackbotMessageEntityType),
	)
	require.EqualError(t, err, "slackbot message backend is unavailable")

	_, err = service.GetMessage(
		t.Context(),
		scope,
		gid.New(scope.GetTenantID(), coredata.SlackMessageEntityType),
	)
	require.EqualError(t, err, "legacy Slack message backend is unavailable")
}

func TestBotDeliveryDestination_UpsertIsolatesOrganizations(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	pgClient, firstScope, firstOrg := newQueueTestOrganizationInTenant(t, tenantID)
	_, secondScope, secondOrg := newQueueTestOrganizationInTenant(t, tenantID)

	target := probot.DeliveryTarget{
		Namespace: "shared_namespace",
		Key:       "shared-key",
	}
	insertQueueDestination(t, pgClient, firstScope, firstOrg, target, true)
	insertQueueDestination(t, pgClient, secondScope, secondOrg, target, false)

	var first, second coredata.BotDeliveryDestination

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				if err := first.LoadByTarget(
					ctx,
					conn,
					firstScope,
					firstOrg,
					ProviderName,
					target.Namespace,
					target.Key,
				); err != nil {
					return err
				}

				return second.LoadByTarget(
					ctx,
					conn,
					secondScope,
					secondOrg,
					ProviderName,
					target.Namespace,
					target.Key,
				)
			},
		),
	)
	assert.Equal(t, firstOrg, first.OrganizationID)
	assert.Equal(t, secondOrg, second.OrganizationID)
	assert.NotEqual(t, first.ID, second.ID)
	require.NotNil(t, first.VerifiedAt)
	assert.Nil(t, second.VerifiedAt)
}

func TestMessageService_RestoreDestination_CASPreservesConcurrentWrite(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	target := probot.DeliveryTarget{
		Namespace: "compliance_portal",
		Key:       "portal-" + organizationID.String(),
	}

	insertQueueDestinationWithExternalID(
		t,
		pgClient,
		scope,
		organizationID,
		target,
		"C-written",
		false,
	)

	// Concurrent selection replaced our write before restore.
	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				destination := coredata.NewBotDeliveryDestination(
					scope,
					organizationID,
					ProviderName,
					target.Namespace,
					target.Key,
				)
				destination.ExternalDestinationID = "C-other"
				destination.ExternalName = "other-channel"
				_, err := destination.Upsert(ctx, tx, scope)

				return err
			},
		),
	)

	service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
	verifiedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	restored, err := service.RestoreDestination(
		t.Context(),
		scope,
		organizationID,
		target,
		&coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-old",
			ExternalName:          "old-channel",
			VerifiedAt:            &verifiedAt,
		},
		"C-written",
	)
	require.NoError(t, err)
	assert.Nil(t, restored)

	current, err := service.GetDestination(t.Context(), scope, organizationID, target)
	require.NoError(t, err)
	assert.Equal(t, "C-other", current.ExternalDestinationID)
}

func TestMessageService_RestoreDestination_CASRestoresWhenUnchanged(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	target := probot.DeliveryTarget{
		Namespace: "compliance_portal",
		Key:       "portal-" + organizationID.String(),
	}

	insertQueueDestinationWithExternalID(
		t,
		pgClient,
		scope,
		organizationID,
		target,
		"C-written",
		false,
	)

	service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
	verifiedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	restored, err := service.RestoreDestination(
		t.Context(),
		scope,
		organizationID,
		target,
		&coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-old",
			ExternalName:          "old-channel",
			VerifiedAt:            &verifiedAt,
		},
		"C-written",
	)
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, "C-old", restored.ExternalDestinationID)
	require.NotNil(t, restored.VerifiedAt)
	assert.True(t, restored.VerifiedAt.Equal(verifiedAt))

	current, err := service.GetDestination(t.Context(), scope, organizationID, target)
	require.NoError(t, err)
	assert.Equal(t, "C-old", current.ExternalDestinationID)
	require.NotNil(t, current.VerifiedAt)
	assert.True(t, current.VerifiedAt.Equal(verifiedAt))
}

func TestMessageService_RestoreDestination_CASClearsWhenUnchanged(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	target := probot.DeliveryTarget{
		Namespace: "compliance_portal",
		Key:       "portal-" + organizationID.String(),
	}

	insertQueueDestinationWithExternalID(
		t,
		pgClient,
		scope,
		organizationID,
		target,
		"C-written",
		false,
	)

	service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
	restored, err := service.RestoreDestination(
		t.Context(),
		scope,
		organizationID,
		target,
		nil,
		"C-written",
	)
	require.NoError(t, err)
	assert.Nil(t, restored)

	_, err = service.GetDestination(t.Context(), scope, organizationID, target)
	require.ErrorIs(t, err, ErrSlackbotChannelNotFound)
}

func TestMessageService_RestoreDestination_CASSkipsClearWhenReplaced(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	target := probot.DeliveryTarget{
		Namespace: "compliance_portal",
		Key:       "portal-" + organizationID.String(),
	}

	insertQueueDestinationWithExternalID(
		t,
		pgClient,
		scope,
		organizationID,
		target,
		"C-other",
		false,
	)

	service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
	restored, err := service.RestoreDestination(
		t.Context(),
		scope,
		organizationID,
		target,
		nil,
		"C-written",
	)
	require.NoError(t, err)
	assert.Nil(t, restored)

	current, err := service.GetDestination(t.Context(), scope, organizationID, target)
	require.NoError(t, err)
	assert.Equal(t, "C-other", current.ExternalDestinationID)
}

func TestMessageService_UpdateWaitsUntilInitialPostIsDelivered(t *testing.T) {
	t.Parallel()

	t.Run("unprocessed post", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		insertBotPost(t, pgClient, scope, organizationID, namespace, key, false, false)
		service := newUpdateMessageService(pgClient, true, false)

		err := service.DeliverOutbound(
			t.Context(),
			updateDelivery(scope, organizationID, namespace, key),
		)
		require.ErrorIs(t, err, probot.ErrInitialPostPending)
		require.ErrorContains(t, err, "before the initial post is queued")
	})

	t.Run("queued slack message", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		post := insertBotPost(t, pgClient, scope, organizationID, namespace, key, true, false)
		insertSlackbotPost(t, pgClient, scope, organizationID, post.ID, false, false)
		service := newUpdateMessageService(pgClient, true, false)

		err := service.DeliverOutbound(
			t.Context(),
			updateDelivery(scope, organizationID, namespace, key),
		)
		require.ErrorIs(t, err, probot.ErrInitialPostPending)
		require.ErrorContains(t, err, "before the initial post is delivered")
	})
}

func TestMessageService_UpdateUnavailableWhenInitialPostCannotDeliver(t *testing.T) {
	t.Parallel()

	t.Run("missing post", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		service := newUpdateMessageService(pgClient, true, false)

		err := service.DeliverOutbound(
			t.Context(),
			updateDelivery(scope, organizationID, namespace, key),
		)
		require.ErrorIs(t, err, probot.ErrInitialPostUnavailable)
	})

	t.Run("dead-lettered post", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		insertBotPost(t, pgClient, scope, organizationID, namespace, key, false, true)
		service := newUpdateMessageService(pgClient, true, false)

		err := service.DeliverOutbound(
			t.Context(),
			updateDelivery(scope, organizationID, namespace, key),
		)
		require.ErrorIs(t, err, probot.ErrInitialPostUnavailable)
	})

	t.Run("terminal slack delivery", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		post := insertBotPost(t, pgClient, scope, organizationID, namespace, key, true, false)
		insertSlackbotPost(t, pgClient, scope, organizationID, post.ID, false, true)
		service := newUpdateMessageService(pgClient, true, false)

		err := service.DeliverOutbound(
			t.Context(),
			updateDelivery(scope, organizationID, namespace, key),
		)
		require.ErrorIs(t, err, probot.ErrInitialPostUnavailable)
		require.ErrorContains(t, err, "initial provider post failed")
	})
}

func TestMessageService_UpdateAppliesWhenInitialPostIsDeliveredWithoutSubject(t *testing.T) {
	t.Parallel()

	t.Run("modern slack message", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		post := insertBotPost(t, pgClient, scope, organizationID, namespace, key, true, false)
		insertSlackbotPost(t, pgClient, scope, organizationID, post.ID, true, false)
		service := newUpdateMessageService(pgClient, true, false)
		delivery := updateDelivery(scope, organizationID, namespace, key)

		require.NoError(t, service.DeliverOutbound(t.Context(), delivery))

		var revision coredata.SlackbotMessage
		require.NoError(
			t,
			pgClient.WithConn(
				t.Context(),
				func(ctx context.Context, conn pg.Querier) error {
					return revision.LoadBySourceEventID(ctx, conn, scope, delivery.SourceEventID)
				},
			),
		)
		assert.NotEqual(t, revision.ID, revision.InitialSlackbotMessageID)
		require.NotNil(t, revision.MessageTS)
		assert.Equal(t, "111.222", *revision.MessageTS)
	})

	t.Run("legacy message when modern queue is configured", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		post := insertBotPost(t, pgClient, scope, organizationID, namespace, key, true, false)
		insertLegacyPost(t, pgClient, scope, organizationID, post.ID)
		service := newUpdateMessageService(pgClient, true, true)
		delivery := updateDelivery(scope, organizationID, namespace, key)

		require.NoError(t, service.DeliverOutbound(t.Context(), delivery))

		var revision coredata.SlackMessage
		require.NoError(
			t,
			pgClient.WithConn(
				t.Context(),
				func(ctx context.Context, conn pg.Querier) error {
					return revision.LoadBySourceEventID(ctx, conn, scope, delivery.SourceEventID)
				},
			),
		)
		assert.NotEqual(t, revision.ID, revision.InitialSlackMessageID)
	})

	t.Run("legacy slack message", func(t *testing.T) {
		t.Parallel()

		pgClient, scope, organizationID := newQueueTestOrganization(t)
		namespace, key := newUpdateSubject(scope)
		post := insertBotPost(t, pgClient, scope, organizationID, namespace, key, true, false)
		insertLegacyPost(t, pgClient, scope, organizationID, post.ID)
		service := newUpdateMessageService(pgClient, false, true)
		delivery := updateDelivery(scope, organizationID, namespace, key)

		require.NoError(t, service.DeliverOutbound(t.Context(), delivery))

		var revision coredata.SlackMessage
		require.NoError(
			t,
			pgClient.WithConn(
				t.Context(),
				func(ctx context.Context, conn pg.Querier) error {
					return revision.LoadBySourceEventID(ctx, conn, scope, delivery.SourceEventID)
				},
			),
		)
		assert.NotEqual(t, revision.ID, revision.InitialSlackMessageID)
		require.NotNil(t, revision.MessageTS)
		assert.Equal(t, "111.222", *revision.MessageTS)
	})
}

func TestMessageService_UpdateUsesDeliveredThreadSubject(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	insertSlackbotPost(t, pgClient, scope, organizationID, gid.Nil, true, false)
	insertThreadSubject(t, pgClient, organizationID, "C-posted", "111.222", map[string]any{})
	service := newUpdateMessageService(pgClient, true, false)
	delivery := updateDelivery(scope, organizationID, "subject", "stable")

	require.NoError(t, service.DeliverOutbound(t.Context(), delivery))

	var revision coredata.SlackbotMessage
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return revision.LoadBySourceEventID(ctx, conn, scope, delivery.SourceEventID)
			},
		),
	)
	assert.NotEqual(t, revision.ID, revision.InitialSlackbotMessageID)
}

func newUpdateMessageService(
	pgClient *pg.Client,
	modern bool,
	legacy bool,
) *MessageService {
	var modernQueue ModernMessageQueue
	if modern {
		modernQueue = NewNotificationService(pgClient)
	}

	var legacyQueue LegacyMessageQueue
	if legacy {
		legacyQueue = legacyslack.NewService(pgClient, "test-secret", "https://slack.test", log.NewLogger())
	}

	return NewMessageService(pgClient, nil, modernQueue, legacyQueue, log.NewLogger())
}

func newUpdateSubject(scope coredata.Scoper) (string, string) {
	return "compliance_portal_access", gid.New(scope.GetTenantID(), coredata.BotMessageEntityType).String()
}

func updateDelivery(
	scope coredata.Scoper,
	organizationID gid.GID,
	namespace string,
	key string,
) probot.OutboundDelivery {
	return probot.OutboundDelivery{
		OrganizationID:   organizationID,
		Purpose:          coredata.BotMessagePurposeUpdate,
		SourceEventID:    gid.New(scope.GetTenantID(), coredata.BotMessageEntityType).String(),
		SubjectNamespace: namespace,
		SubjectKey:       key,
		Capability:       "compliance_access",
		MessageType:      "ACCESS",
		Result: probot.OutboundMessage{
			Message: bot.Message{
				OrganizationID: organizationID,
				Type:           "ACCESS",
			},
			Intent: bot.MessageIntent{FallbackText: "updated"},
		},
	}
}

func insertBotPost(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	namespace string,
	key string,
	processed bool,
	deadLettered bool,
) *coredata.BotMessage {
	t.Helper()

	var message *coredata.BotMessage

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				var err error
				message, err = bot.NewService(bot.ServiceConfig{}).EnqueueMessage(
					ctx,
					tx,
					scope,
					bot.MessageParams{
						OrganizationID:   organizationID,
						Capability:       "compliance_access",
						MessageType:      "ACCESS",
						SubjectNamespace: namespace,
						SubjectKey:       key,
						EventKey:         "post",
						Purpose:          coredata.BotMessagePurposePost,
					},
				)

				return err
			},
		),
	)
	require.NotNil(t, message)

	if !processed && !deadLettered {
		return message
	}

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`
UPDATE bot_messages
SET
	processed_at = CASE WHEN @processed THEN NOW() ELSE NULL END,
	dead_lettered_at = CASE WHEN @dead_lettered THEN NOW() ELSE NULL END,
	updated_at = NOW()
WHERE id = @id
`,
					pgx.StrictNamedArgs{
						"id":            message.ID,
						"processed":     processed,
						"dead_lettered": deadLettered,
					},
				)

				return err
			},
		),
	)

	return message
}

func insertSlackbotPost(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	postID gid.GID,
	delivered bool,
	terminal bool,
) {
	t.Helper()

	metadata := map[string]any{}
	if postID != gid.Nil {
		metadata[coredata.SlackbotMessageMetadataSourceEventID] = postID.String()
	}

	message := coredata.NewSlackbotMessage(
		scope,
		organizationID,
		"ACCESS",
		map[string]any{"text": "posted"},
		metadata,
	)
	message.ChannelID = new("C-posted")
	message.MessageTS = new("111.222")

	now := time.Now()
	if delivered {
		message.SentAt = &now
	}

	if terminal {
		message.Error = new("cannot create Slackbot message")
		message.SentAt = nil
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return message.Insert(ctx, tx, scope)
			},
		),
	)
}

func insertLegacyPost(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	postID gid.GID,
) {
	t.Helper()

	message := coredata.NewSlackMessage(
		scope,
		organizationID,
		coredata.SlackMessageTypeWelcome,
		map[string]any{"text": "posted"},
	)
	message.Metadata = map[string]any{
		coredata.SlackMessageMetadataSourceEventID: postID.String(),
	}
	message.ChannelID = new("C-posted")
	message.MessageTS = new("111.222")
	now := time.Now()
	message.SentAt = &now
	message.UpdatedAt = now

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if err := message.Insert(ctx, tx, scope); err != nil {
					return err
				}

				return message.Update(ctx, tx, scope)
			},
		),
	)
}

func insertQueueDestinationWithExternalID(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
	externalDestinationID string,
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
	destination.ExternalDestinationID = externalDestinationID
	destination.ExternalName = "channel-" + externalDestinationID

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
