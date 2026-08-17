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
	"fmt"
	"maps"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const interactiveClaimStaleAfter = 5 * time.Minute

type (
	NotificationService struct {
		pg *pg.Client
	}

	QueueNotificationRequest struct {
		ID             gid.GID
		OrganizationID gid.GID
		ChannelID      string
		MessageType    string
		Body           map[string]any
		Metadata       map[string]any
		SourceEventID  *string
		DedupKey       *string
		DedupWindow    *time.Duration
	}
)

func NewNotificationService(pgClient *pg.Client) *NotificationService {
	return &NotificationService{
		pg: pgClient,
	}
}

func (s *NotificationService) Queue(
	ctx context.Context,
	scope coredata.Scoper,
	req QueueNotificationRequest,
) (*coredata.SlackbotMessage, error) {
	if err := validateDedup(req.DedupKey, req.DedupWindow); err != nil {
		return nil, err
	}

	metadata := cloneNotificationData(req.Metadata)
	if req.DedupKey != nil {
		metadata[coredata.SlackbotMessageMetadataDedupKey] = *req.DedupKey
	}

	if req.SourceEventID != nil {
		if *req.SourceEventID == "" {
			return nil, fmt.Errorf("source event ID must not be empty")
		}

		metadata[coredata.SlackbotMessageMetadataSourceEventID] = *req.SourceEventID
	}

	message := coredata.NewSlackbotMessage(
		scope,
		req.OrganizationID,
		req.MessageType,
		cloneNotificationData(req.Body),
		metadata,
	)
	if req.ID != gid.Nil {
		message.ID = req.ID
		message.InitialSlackbotMessageID = req.ID
	}

	message.ChannelID = new(req.ChannelID)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if req.SourceEventID != nil {
				var existing coredata.SlackbotMessage

				err := existing.LoadBySourceEventID(ctx, tx, scope, *req.SourceEventID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load Slackbot message by source event: %w", err)
				}

				if err == nil {
					*message = existing
					return nil
				}
			}

			if req.DedupKey != nil {
				var existing coredata.SlackbotMessage

				err := existing.LoadLatestByOrganizationIDChannelIDMessageTypeAndDedupKey(
					ctx,
					tx,
					scope,
					req.OrganizationID,
					req.ChannelID,
					req.MessageType,
					*req.DedupKey,
					message.CreatedAt.Add(-*req.DedupWindow),
				)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load deduplicated Slackbot message: %w", err)
				}

				if err == nil {
					message.InitialSlackbotMessageID = existing.InitialSlackbotMessageID
					message.ChannelID = existing.ChannelID
					message.MessageTS = existing.MessageTS
				}
			}

			if err := message.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) && req.SourceEventID != nil {
					var existing coredata.SlackbotMessage
					if loadErr := existing.LoadBySourceEventID(
						ctx,
						tx,
						scope,
						*req.SourceEventID,
					); loadErr != nil {
						return fmt.Errorf(
							"cannot reload Slackbot message after source event conflict: %w",
							loadErr,
						)
					}

					*message = existing

					return nil
				}

				return fmt.Errorf("cannot queue Slackbot message: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *NotificationService) QueueRevision(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
	body map[string]any,
	metadata map[string]any,
) (*coredata.SlackbotMessage, error) {
	var revision *coredata.SlackbotMessage

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var existing coredata.SlackbotMessage
			if err := existing.LoadByID(ctx, tx, scope, messageID); err != nil {
				return fmt.Errorf("cannot load Slackbot message for revision: %w", err)
			}

			revision = coredata.NewSlackbotMessage(
				scope,
				existing.OrganizationID,
				existing.MessageType,
				cloneNotificationData(body),
				cloneNotificationData(metadata),
			)
			revision.ChannelID = existing.ChannelID
			revision.MessageTS = existing.MessageTS
			revision.InitialSlackbotMessageID = existing.InitialSlackbotMessageID

			if err := revision.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot queue Slackbot message revision: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return revision, nil
}

func (s *NotificationService) GetByID(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
) (*coredata.SlackbotMessage, error) {
	message := &coredata.SlackbotMessage{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := message.LoadByID(ctx, conn, scope, messageID); err != nil {
				return fmt.Errorf("cannot load Slackbot message: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *NotificationService) GetBySourceEventID(
	ctx context.Context,
	scope coredata.Scoper,
	sourceEventID string,
) (*coredata.SlackbotMessage, error) {
	message := &coredata.SlackbotMessage{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := message.LoadBySourceEventID(ctx, conn, scope, sourceEventID); err != nil {
				return fmt.Errorf("cannot load Slackbot message by source event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *NotificationService) GetInitialByOrganizationIDChannelAndTS(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	channelID string,
	messageTS string,
) (*coredata.SlackbotMessage, error) {
	message := &coredata.SlackbotMessage{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := message.LoadLatestDeliveredByOrganizationIDChannelAndTS(
				ctx,
				conn,
				scope,
				organizationID,
				channelID,
				messageTS,
			); err != nil {
				return fmt.Errorf("cannot load initial Slackbot message by organization: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *NotificationService) ClaimInteractiveAction(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	interactionKey string,
) (string, bool, error) {
	processingToken, err := uuid.NewV7()
	if err != nil {
		return "", false, fmt.Errorf("cannot generate interactive claim token: %w", err)
	}

	var claimed bool

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			claim := coredata.NewSlackbotInteractiveClaim(organizationID, interactionKey)

			var err error

			claimed, err = claim.Claim(
				ctx,
				tx,
				scope,
				processingToken.String(),
				time.Now(),
				interactiveClaimStaleAfter,
			)
			if err != nil {
				return fmt.Errorf("cannot claim Slackbot interactive action: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", false, err
	}

	return processingToken.String(), claimed, nil
}

func (s *NotificationService) CompleteInteractiveAction(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	interactionKey string,
	processingToken string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			claim := coredata.NewSlackbotInteractiveClaim(organizationID, interactionKey)

			return claim.Complete(ctx, tx, scope, processingToken, time.Now())
		},
	)
}

func (s *NotificationService) ReleaseInteractiveAction(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	interactionKey string,
	processingToken string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			claim := coredata.NewSlackbotInteractiveClaim(organizationID, interactionKey)

			return claim.Release(ctx, tx, scope, processingToken)
		},
	)
}

func validateDedup(dedupKey *string, dedupWindow *time.Duration) error {
	if (dedupKey == nil) != (dedupWindow == nil) {
		return fmt.Errorf("dedup key and window must be provided together")
	}

	if dedupKey != nil && *dedupKey == "" {
		return fmt.Errorf("dedup key must not be empty")
	}

	if dedupWindow != nil && *dedupWindow <= 0 {
		return fmt.Errorf("dedup window must be positive")
	}

	return nil
}

func cloneNotificationData(data map[string]any) map[string]any {
	cloned := make(map[string]any, len(data))
	maps.Copy(cloned, data)

	return cloned
}
