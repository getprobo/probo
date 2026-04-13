// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package probo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

type (
	AgentRunSupervisor struct {
		pg             *pg.Client
		store          *coredata.PGCheckpointer
		registry       agent.AgentRegistry
		logger         *log.Logger
		interval       time.Duration
		leaseDuration  time.Duration
		maxConcurrency int
		workerID       string

		mu      sync.Mutex
		running map[string]*runHandle // runID -> handle
	}

	// runHandle wraps a stop channel with a sync.Once to prevent double-close panics.
	runHandle struct {
		stopCh chan struct{}
		once   sync.Once
	}

	AgentRunSupervisorOption func(*AgentRunSupervisor)
)

var (
	ErrAgentRunHeartbeatFailed = errors.New("agent run heartbeat failed")
	ErrAgentRunLeaseLost       = errors.New("agent run lease lost")
)

func (h *runHandle) stop() {
	h.once.Do(func() { close(h.stopCh) })
}

func WithAgentRunSupervisorInterval(d time.Duration) AgentRunSupervisorOption {
	return func(s *AgentRunSupervisor) {
		if d > 0 {
			s.interval = d
		}
	}
}

func WithAgentRunSupervisorLeaseDuration(d time.Duration) AgentRunSupervisorOption {
	return func(s *AgentRunSupervisor) {
		if d > 0 {
			s.leaseDuration = d
		}
	}
}

func WithAgentRunSupervisorMaxConcurrency(n int) AgentRunSupervisorOption {
	return func(s *AgentRunSupervisor) {
		if n > 0 {
			s.maxConcurrency = n
		}
	}
}

func NewAgentRunSupervisor(
	pgClient *pg.Client,
	store *coredata.PGCheckpointer,
	registry agent.AgentRegistry,
	logger *log.Logger,
	opts ...AgentRunSupervisorOption,
) *AgentRunSupervisor {
	s := &AgentRunSupervisor{
		pg:             pgClient,
		store:          store,
		registry:       registry,
		logger:         logger,
		interval:       10 * time.Second,
		leaseDuration:  5 * time.Minute,
		maxConcurrency: 5,
		workerID:       uuid.MustNewV4().String(),
		running:        make(map[string]*runHandle),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *AgentRunSupervisor) Run(ctx context.Context) error {
	var (
		wg     sync.WaitGroup
		sem    = make(chan struct{}, s.maxConcurrency)
		ticker = time.NewTicker(s.interval)
	)
	defer ticker.Stop()
	defer wg.Wait()
	defer s.stopAll()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			nonCancelableCtx := context.WithoutCancel(ctx)
			s.recoverStaleRuns(nonCancelableCtx)
			s.checkStopRequests(nonCancelableCtx)

			for {
				if err := s.processNext(ctx, sem, &wg); err != nil {
					if !errors.Is(err, coredata.ErrResourceNotFound) {
						s.logger.ErrorCtx(nonCancelableCtx, "cannot claim agent run", log.Error(err))
					}
					break
				}
			}
		}
	}
}

func (s *AgentRunSupervisor) processNext(
	ctx context.Context,
	sem chan struct{},
	wg *sync.WaitGroup,
) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		run              = coredata.AgentRun{}
		now              = time.Now()
		leaseOwner       = s.workerID
		leaseExpiresAt   = now.Add(s.leaseDuration)
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := s.pg.WithTx(
		nonCancelableCtx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			run.Status = coredata.AgentRunStatusRunning
			run.StartedAt = &now
			run.LeaseOwner = &leaseOwner
			run.LeaseExpiresAt = &leaseExpiresAt
			run.UpdatedAt = now

			if err := run.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update agent run: %w", err)
			}

			return nil
		},
	); err != nil {
		<-sem
		return err
	}

	handle := &runHandle{stopCh: make(chan struct{})}

	s.mu.Lock()
	s.running[run.ID.String()] = handle
	s.mu.Unlock()

	wg.Add(1)
	go func(run coredata.AgentRun) {
		defer wg.Done()
		defer func() { <-sem }()
		defer func() {
			s.mu.Lock()
			delete(s.running, run.ID.String())
			s.mu.Unlock()
		}()

		runCtx, cancelRun := context.WithCancelCause(nonCancelableCtx)
		defer cancelRun(nil)

		heartbeatCtx, cancelHeartbeat := context.WithCancel(nonCancelableCtx)
		defer cancelHeartbeat()
		go s.heartbeatLease(heartbeatCtx, run.ID.String(), cancelRun)

		runCtx = agent.WithStopSignal(runCtx, handle.stopCh)
		s.executeRun(runCtx, &run)
	}(run)

	return nil
}

func (s *AgentRunSupervisor) heartbeatLease(
	ctx context.Context,
	runID string,
	cancelRun context.CancelCauseFunc,
) {
	ticker := time.NewTicker(s.leaseDuration / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expiresAt := time.Now().Add(s.leaseDuration)

			if err := s.pg.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					rowsAffected, err := coredata.HeartbeatAgentRunLease(ctx, conn, runID, s.workerID, expiresAt)
					if err != nil {
						return err
					}

					if rowsAffected == 0 {
						return ErrAgentRunLeaseLost
					}

					return nil
				},
			); err != nil {
				s.logger.ErrorCtx(ctx, "cannot heartbeat agent run lease", log.Error(err))
				if errors.Is(err, ErrAgentRunLeaseLost) {
					cancelRun(ErrAgentRunLeaseLost)
				} else {
					cancelRun(fmt.Errorf("%w: %w", ErrAgentRunHeartbeatFailed, err))
				}
				return
			}
		}
	}
}

func (s *AgentRunSupervisor) executeRun(ctx context.Context, run *coredata.AgentRun) {
	runID := run.ID.String()

	var (
		result *agent.Result
		runErr error
	)

	if run.Checkpoint != nil {
		// Resume from checkpoint.
		s.logger.InfoCtx(ctx, "resuming agent run", log.String("run_id", runID))
		result, runErr = agent.Restore(ctx, s.store, runID, s.registry)
	} else {
		// Start fresh.
		s.logger.InfoCtx(ctx, "starting agent run", log.String("run_id", runID))

		a, err := s.registry.Agent(run.StartAgentName)
		if err != nil {
			runErr = fmt.Errorf("cannot resolve agent %q: %w", run.StartAgentName, err)
		} else {
			var inputMsgs []llm.Message
			if err := json.Unmarshal(run.InputMessages, &inputMsgs); err != nil {
				runErr = fmt.Errorf("cannot unmarshal input messages: %w", err)
			} else {
				result, runErr = a.RunWithOpts(
					ctx,
					inputMsgs,
					agent.WithCheckpointer(s.store, runID),
				)
			}
		}
	}

	// Heartbeat loss: another worker may have taken over. Do not commit
	// any status — stale recovery will handle the row.
	if cause := context.Cause(ctx); errors.Is(cause, ErrAgentRunLeaseLost) || errors.Is(cause, ErrAgentRunHeartbeatFailed) {
		s.logger.WarnCtx(
			context.WithoutCancel(ctx),
			"agent run stopped after heartbeat failure; leaving status for stale recovery",
			log.String("run_id", runID),
			log.Error(cause),
		)
		return
	}

	// Infrastructure-triggered suspension (graceful shutdown): leave the
	// row as RUNNING so stale recovery resets it to PENDING on restart.
	// The checkpoint was already saved by coreLoop before returning
	// SuspendedError, so Restore will pick up where it left off.
	if runErr != nil {
		if _, ok := errors.AsType[*agent.SuspendedError](runErr); ok {
			s.logger.InfoCtx(
				context.WithoutCancel(ctx),
				"agent run suspended by infrastructure; leaving for stale recovery",
				log.String("run_id", runID),
			)
			return
		}
	}

	// Update run status based on outcome.
	now := time.Now()
	run.UpdatedAt = now
	run.StartedAt = nil
	run.LeaseOwner = nil
	run.LeaseExpiresAt = nil

	switch {
	case runErr == nil:
		run.Status = coredata.AgentRunStatusCompleted
		if result != nil {
			data, err := json.Marshal(result)
			if err != nil {
				s.logger.ErrorCtx(ctx, "cannot marshal agent run result", log.Error(err))
			} else {
				run.Result = data
			}
		}
		run.StopRequested = false

	default:
		run.Status = coredata.AgentRunStatusFailed
		msg := runErr.Error()
		run.ErrorMessage = &msg
	}

	commitCtx := context.WithoutCancel(ctx)

	if err := s.pg.WithTx(
		commitCtx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return err
			}

			if run.Status == coredata.AgentRunStatusCompleted {
				if err := run.ClearCheckpoint(ctx, tx, coredata.NewNoScope()); err != nil {
					return err
				}
			}

			return nil
		},
	); err != nil {
		s.logger.ErrorCtx(commitCtx, "cannot commit agent run status", log.Error(err))
	}
}

func (s *AgentRunSupervisor) checkStopRequests(ctx context.Context) {
	var ids []string

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			ids, err = coredata.LoadRunningStopRequestedIDs(ctx, conn)
			return err
		},
	); err != nil {
		s.logger.ErrorCtx(ctx, "cannot check stop requests", log.Error(err))
		return
	}

	for _, id := range ids {
		s.mu.Lock()
		if h, ok := s.running[id]; ok {
			h.stop()
		}
		s.mu.Unlock()
	}
}

func (s *AgentRunSupervisor) stopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, h := range s.running {
		h.stop()
	}
}

func (s *AgentRunSupervisor) recoverStaleRuns(ctx context.Context) {
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleAgentRuns(ctx, conn); err != nil {
				return fmt.Errorf("cannot reset stale agent runs: %w", err)
			}

			return nil
		},
	); err != nil {
		s.logger.ErrorCtx(ctx, "cannot recover stale agent runs", log.Error(err))
	}
}
