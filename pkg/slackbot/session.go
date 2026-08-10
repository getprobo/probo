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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

type agentSession struct {
	pg      *pg.Client
	agentID string
}

func newAgentSession(pgClient *pg.Client, agentID string) *agentSession {
	return &agentSession{pg: pgClient, agentID: agentID}
}

func (s *agentSession) Load(ctx context.Context, _ string) ([]llm.Message, error) {
	var agent coredata.SlackbotAgent

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := agent.LoadByAgentID(ctx, conn, s.agentID); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load messages: %w", err)
	}
	if len(agent.Messages) == 0 {
		return nil, nil
	}

	var messages []llm.Message
	if err := json.Unmarshal(agent.Messages, &messages); err != nil {
		return nil, fmt.Errorf("cannot unmarshal messages: %w", err)
	}

	return messages, nil
}

func (s *agentSession) Save(ctx context.Context, _ string, messages []llm.Message) error {
	return saveMessages(ctx, s.pg, s.agentID, messages)
}
