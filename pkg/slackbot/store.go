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
	"go.probo.inc/probo/pkg/llm"
)

type (
	resumableAgent struct {
		AgentID     string
		SessionID   string
		Channel     string
		ThreadTS    string
		SlackUserID string
	}

	agentScope struct {
		Channel     string
		ThreadTS    string
		SessionID   string
		SlackUserID string
	}
)

func upsertAgent(
	ctx context.Context,
	pgClient *pg.Client,
	sessionID, channel, threadTS, slackUserID string,
) (string, error) {
	agentID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("cannot generate agent id: %w", err)
	}

	now := time.Now()
	agent := coredata.SlackbotAgent{
		AgentID:     agentID.String(),
		SessionID:   sessionID,
		Channel:     channel,
		ThreadTS:    threadTS,
		SlackUserID: slackUserID,
		Status:      coredata.SlackbotAgentStatusAvailable,
		Messages:    json.RawMessage("[]"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if _, err := agent.UpsertBySessionID(ctx, conn); err != nil {
				return fmt.Errorf("cannot upsert agent: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return agent.AgentID, nil
}

func loadAgentScope(ctx context.Context, pgClient *pg.Client, agentID string) (agentScope, error) {
	var agent coredata.SlackbotAgent

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := agent.LoadByAgentID(ctx, conn, agentID); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return agentScope{}, fmt.Errorf("cannot load agent scope: %w", err)
	}

	return agentScope{
		Channel:     agent.Channel,
		ThreadTS:    agent.ThreadTS,
		SessionID:   agent.SessionID,
		SlackUserID: agent.SlackUserID,
	}, nil
}

func claimAgent(ctx context.Context, pgClient *pg.Client, agentID string) (bool, error) {
	var claimed bool

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			agent := coredata.SlackbotAgent{AgentID: agentID}

			var err error
			claimed, err = agent.ClaimAvailable(ctx, conn, time.Now())
			if err != nil {
				return err
			}

			return nil
		},
	)

	return claimed, err
}

func markAvailable(ctx context.Context, pgClient *pg.Client, agentID string) error {
	return pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			agent := coredata.SlackbotAgent{AgentID: agentID}
			if err := agent.MarkAvailable(ctx, conn, time.Now()); err != nil {
				return err
			}

			return nil
		},
	)
}

func saveMessages(ctx context.Context, pgClient *pg.Client, agentID string, messages []llm.Message) error {
	if messages == nil {
		messages = []llm.Message{}
	}

	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("cannot marshal messages: %w", err)
	}

	return pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			agent := coredata.SlackbotAgent{
				AgentID:  agentID,
				Messages: data,
			}
			if err := agent.UpdateMessages(ctx, conn, time.Now()); err != nil {
				return err
			}

			return nil
		},
	)
}

func listResumable(ctx context.Context, pgClient *pg.Client) ([]resumableAgent, error) {
	var agents coredata.SlackbotAgents

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := agents.ClaimResumable(ctx, conn, time.Now()); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list resumable agents: %w", err)
	}

	out := make([]resumableAgent, 0, len(agents))
	for _, a := range agents {
		out = append(
			out,
			resumableAgent{
				AgentID:     a.AgentID,
				SessionID:   a.SessionID,
				Channel:     a.Channel,
				ThreadTS:    a.ThreadTS,
				SlackUserID: a.SlackUserID,
			},
		)
	}

	return out, nil
}
