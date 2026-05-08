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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

const (
	// cloudAccountDisconnectFailureThreshold is the number of
	// consecutive probe failures required before the worker
	// promotes a row from ERRORED to DISCONNECTED.
	cloudAccountDisconnectFailureThreshold = 3

	// cloudAccountDisconnectTimeGate is the minimum elapsed time
	// between first_probe_failure_at and the present moment before
	// the worker promotes a row to DISCONNECTED. Combined with the
	// failure-count threshold this prevents flapping a long-lived
	// account into DISCONNECTED from a single 5xx burst.
	cloudAccountDisconnectTimeGate = time.Hour
)

type (
	// CloudAccountWorkerConfig tunes the periodic-probe worker.
	// Zero-valued fields fall back to the defaults
	// (StaleAfter = 15 minutes).
	CloudAccountWorkerConfig struct {
		// StaleAfter is the elapsed time since last_probe_at that
		// makes a row eligible for re-probing.
		StaleAfter time.Duration
	}

	cloudAccountHandler struct {
		pg            *pg.Client
		registry      *cloudaccount.Registry
		encryptionKey cipher.EncryptionKey
		logger        *log.Logger
		staleAfter    time.Duration
	}
)

// NewCloudAccountWorker builds the periodic cloud-account probe
// worker. Defaults: WithInterval(15*time.Minute),
// WithMaxConcurrency(4). Caller (probod) overrides via
// probodconfig.CloudAccount.{ProbeInterval,ProbeMaxConcurrency}.
func NewCloudAccountWorker(
	pgClient *pg.Client,
	registry *cloudaccount.Registry,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
	cfg CloudAccountWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.CloudAccount] {
	staleAfter := cfg.StaleAfter
	if staleAfter == 0 {
		staleAfter = 15 * time.Minute
	}

	h := &cloudAccountHandler{
		pg:            pgClient,
		registry:      registry,
		encryptionKey: encryptionKey,
		logger:        logger,
		staleAfter:    staleAfter,
	}

	defaults := []worker.Option{
		worker.WithInterval(15 * time.Minute),
		worker.WithMaxConcurrency(4),
	}

	return worker.New(
		"cloud-account-worker",
		h,
		logger,
		append(defaults, opts...)...,
	)
}

// Claim returns the next cloud account whose last_probe_at is older
// than staleAfter (or NULL). The row is locked with FOR UPDATE SKIP
// LOCKED so concurrent worker goroutines cannot pick the same row.
// Returns worker.ErrNoTask when no row is due.
func (h *cloudAccountHandler) Claim(ctx context.Context) (coredata.CloudAccount, error) {
	var account coredata.CloudAccount

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := account.LoadNextStaleForUpdateSkipLocked(ctx, tx, h.staleAfter); err != nil {
				return err
			}

			now := time.Now()
			account.LastProbeAt = &now
			account.UpdatedAt = now
			if err := account.Update(ctx, tx, coredata.NewNoScope(), h.encryptionKey); err != nil {
				return fmt.Errorf("cannot update cloud account: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CloudAccount{}, worker.ErrNoTask
		}
		return coredata.CloudAccount{}, err
	}

	return account, nil
}

// Process probes the supplied cloud account out of any DB
// transaction, then opens a short status-transition tx. Logged
// fields are restricted to opaque IDs -- never log external_id, role
// ARN, SA email, client secret, or the raw SDK error payload (the
// error message is persisted on the row already).
func (h *cloudAccountHandler) Process(ctx context.Context, account coredata.CloudAccount) error {
	record := cloudAccountToRecord(&account)

	probeable, err := h.registry.BuildProbeable(record)
	if err != nil {
		return h.commitFailure(ctx, &account, err)
	}

	probeErr := probeable.Probe(ctx)

	if probeErr == nil {
		return h.commitSuccess(ctx, &account)
	}

	h.logger.WarnCtx(
		ctx,
		"cloud account probe failed",
		log.String("cloud_account_id", account.ID.String()),
	)

	return h.commitFailure(ctx, &account, probeErr)
}

// cloudAccountWebhookEvent names the webhook the worker should emit
// alongside a state transition. EmitNone means no webhook.
type cloudAccountWebhookEvent int

const (
	cloudAccountWebhookNone cloudAccountWebhookEvent = iota
	cloudAccountWebhookVerified
	cloudAccountWebhookDisconnected
)

// cloudAccountTransition is the pure-function output of
// computeCloudAccountTransition: the post-transition row state plus
// the webhook (if any) that should be emitted in the same tx.
//
// All time-dependent fields are written through the supplied `now`
// so deterministic tests can drive the state machine without a clock
// dependency.
type cloudAccountTransition struct {
	Status                   coredata.CloudAccountStatus
	ConsecutiveProbeFailures int
	FirstProbeFailureAt      *time.Time
	LastProbeAt              time.Time
	LastVerifiedAt           *time.Time
	LastProbeError           *string
	Webhook                  cloudAccountWebhookEvent
}

// computeCloudAccountTransition is the pure state-machine kernel of
// the periodic-probe worker. Given the row's current persisted
// fields, the latest probe outcome, the wall-clock `now`, and the
// disconnect thresholds, it returns the new status + bookkeeping
// fields and the webhook that must be emitted in the same tx as the
// row update.
//
// Callers (worker `Process` path) must mutate the loaded row from
// the returned transition and persist the row + webhook in a single
// pg.WithTx so the row update and webhook insert observe each other
// atomically.
func computeCloudAccountTransition(
	currentStatus coredata.CloudAccountStatus,
	consecutiveFailures int,
	firstFailureAt *time.Time,
	probeErr error,
	now time.Time,
	failureThreshold int,
	timeGate time.Duration,
) cloudAccountTransition {
	if probeErr == nil {
		wasErrored := currentStatus == coredata.CloudAccountStatusErrored ||
			currentStatus == coredata.CloudAccountStatusDisconnected
		verifiedAt := now
		webhook := cloudAccountWebhookNone
		if wasErrored {
			webhook = cloudAccountWebhookVerified
		}
		return cloudAccountTransition{
			Status:                   coredata.CloudAccountStatusVerified,
			ConsecutiveProbeFailures: 0,
			FirstProbeFailureAt:      nil,
			LastProbeAt:              now,
			LastVerifiedAt:           &verifiedAt,
			LastProbeError:           nil,
			Webhook:                  webhook,
		}
	}

	errMsg := probeErr.Error()

	switch currentStatus {
	case coredata.CloudAccountStatusPendingVerification:
		return cloudAccountTransition{
			Status:                   coredata.CloudAccountStatusPendingVerification,
			ConsecutiveProbeFailures: consecutiveFailures,
			FirstProbeFailureAt:      firstFailureAt,
			LastProbeAt:              now,
			LastVerifiedAt:           nil,
			LastProbeError:           &errMsg,
			Webhook:                  cloudAccountWebhookNone,
		}

	case coredata.CloudAccountStatusVerified:
		first := now
		return cloudAccountTransition{
			Status:                   coredata.CloudAccountStatusErrored,
			ConsecutiveProbeFailures: 1,
			FirstProbeFailureAt:      &first,
			LastProbeAt:              now,
			LastVerifiedAt:           nil,
			LastProbeError:           &errMsg,
			Webhook:                  cloudAccountWebhookNone,
		}

	case coredata.CloudAccountStatusErrored:
		newFailures := consecutiveFailures + 1
		first := firstFailureAt
		if first == nil {
			n := now
			first = &n
		}

		disconnect := newFailures >= failureThreshold && now.Sub(*first) >= timeGate

		nextStatus := coredata.CloudAccountStatusErrored
		webhook := cloudAccountWebhookNone
		if disconnect {
			nextStatus = coredata.CloudAccountStatusDisconnected
			webhook = cloudAccountWebhookDisconnected
		}

		return cloudAccountTransition{
			Status:                   nextStatus,
			ConsecutiveProbeFailures: newFailures,
			FirstProbeFailureAt:      first,
			LastProbeAt:              now,
			LastVerifiedAt:           nil,
			LastProbeError:           &errMsg,
			Webhook:                  webhook,
		}

	case coredata.CloudAccountStatusDisconnected:
		// Stay DISCONNECTED on further failures; surface the latest
		// error string only.
		return cloudAccountTransition{
			Status:                   coredata.CloudAccountStatusDisconnected,
			ConsecutiveProbeFailures: consecutiveFailures + 1,
			FirstProbeFailureAt:      firstFailureAt,
			LastProbeAt:              now,
			LastVerifiedAt:           nil,
			LastProbeError:           &errMsg,
			Webhook:                  cloudAccountWebhookNone,
		}

	default:
		// Unknown status: surface the error and let the caller decide
		// (the worker treats this as a no-op state and logs).
		return cloudAccountTransition{
			Status:                   currentStatus,
			ConsecutiveProbeFailures: consecutiveFailures,
			FirstProbeFailureAt:      firstFailureAt,
			LastProbeAt:              now,
			LastVerifiedAt:           nil,
			LastProbeError:           &errMsg,
			Webhook:                  cloudAccountWebhookNone,
		}
	}
}

func (h *cloudAccountHandler) commitSuccess(
	ctx context.Context,
	account *coredata.CloudAccount,
) error {
	scope := coredata.NewScopeFromObjectID(account.ID)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			fresh := &coredata.CloudAccount{}
			if err := fresh.LoadByID(ctx, tx, scope, account.ID, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot reload cloud account: %w", err)
			}

			now := time.Now()
			transition := computeCloudAccountTransition(
				fresh.Status,
				fresh.ConsecutiveProbeFailures,
				fresh.FirstProbeFailureAt,
				nil,
				now,
				cloudAccountDisconnectFailureThreshold,
				cloudAccountDisconnectTimeGate,
			)

			applyCloudAccountTransition(fresh, transition)

			if err := fresh.Update(ctx, tx, scope, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot update cloud account: %w", err)
			}

			if transition.Webhook == cloudAccountWebhookVerified {
				if err := webhook.InsertData(
					ctx,
					tx,
					scope,
					fresh.OrganizationID,
					coredata.WebhookEventTypeCloudAccountVerified,
					webhooktypes.NewCloudAccount(fresh),
				); err != nil {
					return fmt.Errorf("cannot insert cloud_account.verified webhook: %w", err)
				}
			}

			return nil
		},
	)
}

func (h *cloudAccountHandler) commitFailure(
	ctx context.Context,
	account *coredata.CloudAccount,
	probeErr error,
) error {
	scope := coredata.NewScopeFromObjectID(account.ID)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			fresh := &coredata.CloudAccount{}
			if err := fresh.LoadByID(ctx, tx, scope, account.ID, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot reload cloud account: %w", err)
			}

			now := time.Now()
			transition := computeCloudAccountTransition(
				fresh.Status,
				fresh.ConsecutiveProbeFailures,
				fresh.FirstProbeFailureAt,
				probeErr,
				now,
				cloudAccountDisconnectFailureThreshold,
				cloudAccountDisconnectTimeGate,
			)

			applyCloudAccountTransition(fresh, transition)

			if err := fresh.Update(ctx, tx, scope, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot update cloud account: %w", err)
			}

			if transition.Webhook == cloudAccountWebhookDisconnected {
				if err := webhook.InsertData(
					ctx,
					tx,
					scope,
					fresh.OrganizationID,
					coredata.WebhookEventTypeCloudAccountDisconnected,
					webhooktypes.NewCloudAccount(fresh),
				); err != nil {
					return fmt.Errorf("cannot insert cloud_account.disconnected webhook: %w", err)
				}
			}

			return nil
		},
	)
}

// applyCloudAccountTransition writes the post-transition fields onto
// the supplied row. The caller is responsible for persisting the row.
func applyCloudAccountTransition(account *coredata.CloudAccount, t cloudAccountTransition) {
	account.Status = t.Status
	account.ConsecutiveProbeFailures = t.ConsecutiveProbeFailures
	account.FirstProbeFailureAt = t.FirstProbeFailureAt
	probeAt := t.LastProbeAt
	account.LastProbeAt = &probeAt
	if t.LastVerifiedAt != nil {
		account.LastVerifiedAt = t.LastVerifiedAt
	}
	account.LastProbeError = t.LastProbeError
	account.UpdatedAt = t.LastProbeAt
}

// cloudAccountToRecord mirrors mapCloudAccountToRecord (private to
// cloud_account_service.go) so the worker can build a registry record
// without importing the service layer.
func cloudAccountToRecord(account *coredata.CloudAccount) cloudaccount.CloudAccountRecord {
	rec := cloudaccount.CloudAccountRecord{
		ID:                   account.ID.String(),
		Provider:             account.Provider,
		Kind:                 account.CredentialKind,
		ScopeKind:            account.ScopeKind,
		ScopeIdentifier:      account.ScopeIdentifier,
		DecryptedCredentials: account.DecryptedCredentials,
	}
	if account.ExternalID != nil {
		rec.ExternalID = *account.ExternalID
	}
	return rec
}
