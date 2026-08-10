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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
)

type PGCheckpointer struct {
	pg *pg.Client
}

func NewCheckpointer(pgClient *pg.Client) *PGCheckpointer {
	return &PGCheckpointer{pg: pgClient}
}

func (c *PGCheckpointer) Save(ctx context.Context, agentID string, cp *agent.Checkpoint) error {
	data, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("cannot marshal checkpoint: %w", err)
	}

	return c.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			agentRow := coredata.SlackbotAgent{
				AgentID:    agentID,
				Checkpoint: data,
			}
			if err := agentRow.UpdateCheckpoint(ctx, conn, time.Now()); err != nil {
				return err
			}

			return nil
		},
	)
}

func (c *PGCheckpointer) Load(ctx context.Context, agentID string) (*agent.Checkpoint, error) {
	var agentRow coredata.SlackbotAgent

	err := c.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := agentRow.LoadByAgentID(ctx, conn, agentID); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load checkpoint: %w", err)
	}
	if len(agentRow.Checkpoint) == 0 {
		return nil, nil
	}

	var cp agent.Checkpoint
	if err := json.Unmarshal(agentRow.Checkpoint, &cp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal checkpoint: %w", err)
	}

	return &cp, nil
}
