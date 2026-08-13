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
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type (
	QueueRequest struct {
		ID             gid.GID
		OrganizationID gid.GID
		MessageType    coredata.SlackMessageType
		Body           map[string]any
		Metadata       map[string]any
		RequesterEmail *mail.Addr
		SourceEventID  *string
		DedupKey       *string
		DedupWindow    *time.Duration
	}
)

func (s *Service) Queue(
	ctx context.Context,
	scope coredata.Scoper,
	req QueueRequest,
) (*coredata.SlackMessage, error) {
	if err := validateDedup(req.DedupKey, req.DedupWindow); err != nil {
		return nil, err
	}

	message := coredata.NewSlackMessage(scope, req.OrganizationID, req.MessageType, cloneData(req.Body))
	if req.ID != gid.Nil {
		message.ID = req.ID
		message.InitialSlackMessageID = req.ID
	}

	message.Metadata = cloneData(req.Metadata)

	message.RequesterEmail = req.RequesterEmail
	if req.DedupKey != nil {
		message.Metadata[coredata.SlackMessageMetadataDedupKey] = *req.DedupKey
	}

	if req.SourceEventID != nil {
		if *req.SourceEventID == "" {
			return nil, fmt.Errorf("source event ID must not be empty")
		}

		message.Metadata[coredata.SlackMessageMetadataSourceEventID] = *req.SourceEventID
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if req.SourceEventID != nil {
				var existing coredata.SlackMessage

				err := existing.LoadBySourceEventID(ctx, tx, scope, *req.SourceEventID)
				if err != nil && !errors.Is(err, coredata.ErrSlackMessageNotFound{}) {
					return fmt.Errorf("cannot load legacy Slack message by source event: %w", err)
				}

				if err == nil {
					*message = existing
					return nil
				}
			}

			if req.DedupKey != nil {
				var existing coredata.SlackMessage

				err := existing.LoadLatestByOrganizationIDMessageTypeAndDedupKey(
					ctx,
					tx,
					scope,
					req.OrganizationID,
					req.MessageType,
					*req.DedupKey,
					message.CreatedAt.Add(-*req.DedupWindow),
				)
				if err != nil && !errors.Is(err, coredata.ErrSlackMessageNotFound{}) {
					return fmt.Errorf("cannot load deduplicated legacy Slack message: %w", err)
				}

				if err == nil {
					message.InitialSlackMessageID = existing.InitialSlackMessageID
					message.ChannelID = existing.ChannelID
					message.MessageTS = existing.MessageTS
				}
			}

			if err := message.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot queue legacy Slack message: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *Service) QueueRevision(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
	body map[string]any,
	metadata map[string]any,
) (*coredata.SlackMessage, error) {
	var revision *coredata.SlackMessage

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var existing coredata.SlackMessage
			if err := existing.LoadById(ctx, tx, scope, messageID); err != nil {
				return fmt.Errorf("cannot load legacy Slack message for revision: %w", err)
			}

			revision = coredata.NewSlackMessage(
				scope,
				existing.OrganizationID,
				existing.Type,
				cloneData(body),
			)
			revision.Metadata = cloneData(metadata)
			revision.RequesterEmail = existing.RequesterEmail
			revision.ChannelID = existing.ChannelID
			revision.MessageTS = existing.MessageTS
			revision.InitialSlackMessageID = existing.InitialSlackMessageID

			if err := revision.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot queue legacy Slack message revision: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return revision, nil
}

func (s *Service) GetByID(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
) (*coredata.SlackMessage, error) {
	message := &coredata.SlackMessage{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return message.LoadById(ctx, conn, scope, messageID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load legacy Slack message: %w", err)
	}

	return message, nil
}

func (s *Service) GetBySourceEventID(
	ctx context.Context,
	scope coredata.Scoper,
	sourceEventID string,
) (*coredata.SlackMessage, error) {
	message := &coredata.SlackMessage{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return message.LoadBySourceEventID(ctx, conn, scope, sourceEventID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load legacy Slack message by source event: %w", err)
	}

	return message, nil
}

func (s *Service) GetInitialByChannelAndTS(
	ctx context.Context,
	scope coredata.Scoper,
	channelID string,
	messageTS string,
) (*coredata.SlackMessage, error) {
	message := &coredata.SlackMessage{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return message.LoadLatestDeliveredByChannelAndTS(ctx, conn, scope, channelID, messageTS)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load initial legacy Slack message: %w", err)
	}

	return message, nil
}

func (s *Service) GetInitialByOrganizationIDChannelAndTS(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	channelID string,
	messageTS string,
) (*coredata.SlackMessage, error) {
	message := &coredata.SlackMessage{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return message.LoadLatestDeliveredByOrganizationIDChannelAndTS(
				ctx,
				conn,
				scope,
				organizationID,
				channelID,
				messageTS,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load initial legacy Slack message by organization: %w", err)
	}

	return message, nil
}

func (s *Service) UpdateViaResponseURL(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
	responseURL string,
	body map[string]any,
	metadata map[string]any,
) (*coredata.SlackMessage, error) {
	if responseURL == "" {
		return s.QueueRevision(ctx, scope, messageID, body, metadata)
	}

	existing, err := s.GetByID(ctx, scope, messageID)
	if err != nil {
		return nil, err
	}

	if err := s.GetSlackClient().UpdateInteractiveMessage(ctx, responseURL, body); err != nil {
		return nil, fmt.Errorf("cannot update legacy Slack message via response URL: %w", err)
	}

	now := time.Now()
	revision := coredata.NewSlackMessage(
		scope,
		existing.OrganizationID,
		existing.Type,
		cloneData(body),
	)
	revision.Metadata = cloneData(metadata)
	revision.RequesterEmail = existing.RequesterEmail
	revision.ChannelID = existing.ChannelID
	revision.MessageTS = existing.MessageTS
	revision.InitialSlackMessageID = existing.InitialSlackMessageID
	revision.SentAt = &now

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return revision.Insert(ctx, tx, scope)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot record legacy Slack response revision: %w", err)
	}

	return revision, nil
}

func validateDedup(key *string, window *time.Duration) error {
	if (key == nil) != (window == nil) {
		return fmt.Errorf("dedup key and window must be provided together")
	}

	if key != nil && *key == "" {
		return fmt.Errorf("dedup key must not be empty")
	}

	if window != nil && *window <= 0 {
		return fmt.Errorf("dedup window must be positive")
	}

	return nil
}

func cloneData(data map[string]any) map[string]any {
	cloned := make(map[string]any, len(data))
	for key, value := range data {
		cloned[key] = value
	}

	return cloned
}
