// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

var (
	ErrDiscoveryInProgress = errors.New("a github discovery run is already in progress for this connector")
)

type (
	Service struct {
		pg *pg.Client
	}

	RunRequest struct {
		ConnectorID gid.GID
	}
)

func NewService(pgClient *pg.Client) *Service {
	return &Service{pg: pgClient}
}

func (req RunRequest) Validate() error {
	v := validator.New()

	v.Check(req.ConnectorID, "connector_id", validator.Required(), validator.GID(coredata.ConnectorEntityType))

	return v.Error()
}

func (s *Service) Run(
	ctx context.Context,
	scope coredata.Scoper,
	req RunRequest,
) (*coredata.AgentRun, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	run := &coredata.AgentRun{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			connector := &coredata.Connector{}

			if err := connector.LoadMetadataByID(ctx, tx, scope, req.ConnectorID); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			if connector.Provider != coredata.ConnectorProviderGitHub {
				return fmt.Errorf("connector provider is %s, expected GITHUB", connector.Provider)
			}

			if err := assertNoActiveDiscoveryRun(ctx, tx, req.ConnectorID); err != nil {
				return err
			}

			input := RunInput{
				ConnectorID:    req.ConnectorID,
				OrganizationID: connector.OrganizationID,
				RunKind:        "discovery",
			}

			inputJSON, err := json.Marshal(input)
			if err != nil {
				return fmt.Errorf("cannot marshal discovery run input: %w", err)
			}

			now := time.Now()
			*run = coredata.AgentRun{
				ID:             gid.New(scope.GetTenantID(), coredata.AgentRunEntityType),
				OrganizationID: connector.OrganizationID,
				StartAgentName: StartAgentName,
				Status:         coredata.AgentRunStatusPending,
				InputMessages:  inputJSON,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := run.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot enqueue github discovery run: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return run, nil
}

func assertNoActiveDiscoveryRun(ctx context.Context, tx pg.Tx, connectorID gid.GID) error {
	var active bool

	q := `
SELECT EXISTS (
	SELECT 1
	FROM agent_runs
	WHERE
		start_agent_name = @start_agent_name
		AND status IN ('PENDING', 'RUNNING')
		AND input_messages ->> 'connector_id' = @connector_id
);
`

	args := pgx.StrictNamedArgs{
		"start_agent_name": StartAgentName,
		"connector_id":     connectorID.String(),
	}

	if err := tx.QueryRow(ctx, q, args).Scan(&active); err != nil {
		return fmt.Errorf("cannot check active github discovery runs: %w", err)
	}

	if active {
		return ErrDiscoveryInProgress
	}

	return nil
}
