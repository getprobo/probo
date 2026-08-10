// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package slackbot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

const pendingInteractionLimit = 100

type interaction struct {
	InteractionID string
	EventType     string
	Payload       EventBody
}

func insertInteraction(
	ctx context.Context,
	pgClient *pg.Client,
	agentID, eventID, eventType string,
	payload EventBody,
) error {
	interactionID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("cannot generate interaction id: %w", err)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload: %w", err)
	}

	var eventIDVal *string
	if eventID != "" {
		eventIDVal = &eventID
	}

	row := coredata.SlackbotInteraction{
		InteractionID: interactionID.String(),
		AgentID:       agentID,
		EventID:       eventIDVal,
		EventType:     eventType,
		Payload:       data,
		CreatedAt:     time.Now(),
	}

	return pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := row.Insert(ctx, conn); err != nil {
				return err
			}

			return nil
		},
	)
}

// claimEventID records that an event has been handled. Returns false when the
// event was already claimed (Slack retry) so callers can skip side effects.
// Empty event IDs cannot be deduplicated and always claim successfully.
func claimEventID(ctx context.Context, pgClient *pg.Client, eventID string) (bool, error) {
	if eventID == "" {
		return true, nil
	}

	var claimed bool

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			event := coredata.SlackbotProcessedEvent{
				EventID:   eventID,
				CreatedAt: time.Now(),
			}

			var err error
			claimed, err = event.Claim(ctx, conn)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot claim slack event id: %w", err)
	}

	return claimed, nil
}

func listPending(ctx context.Context, pgClient *pg.Client, agentID string) ([]interaction, error) {
	var rows coredata.SlackbotInteractions

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := rows.LoadPendingByAgentID(ctx, conn, agentID, pendingInteractionLimit); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list pending interactions: %w", err)
	}

	interactions := make([]interaction, 0, len(rows))
	for _, row := range rows {
		var payload EventBody
		if err := json.Unmarshal(row.Payload, &payload); err != nil {
			return nil, fmt.Errorf("cannot unmarshal payload: %w", err)
		}

		interactions = append(
			interactions,
			interaction{
				InteractionID: row.InteractionID,
				EventType:     row.EventType,
				Payload:       payload,
			},
		)
	}

	return interactions, nil
}

func markProcessed(ctx context.Context, pgClient *pg.Client, interactionIDs []string) error {
	if len(interactionIDs) == 0 {
		return nil
	}

	return pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.MarkSlackbotInteractionsProcessed(
				ctx,
				conn,
				interactionIDs,
				time.Now(),
			); err != nil {
				return err
			}

			return nil
		},
	)
}
