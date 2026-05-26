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

package deviceagent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/deviceagent/checks"
	"go.probo.inc/probo/pkg/deviceagent/update"
)

const (
	hostInfoRefreshInterval = 6 * time.Hour
	perCheckTimeout         = 15 * time.Second

	pendingFlushBackoffMin = 15 * time.Second
	pendingFlushBackoffMax = 30 * time.Minute

	updateCheckTimeout = 10 * time.Minute
)

// ErrRestartRequired is returned by Agent.Run after a successful
// in-place upgrade of the agent binary. Callers should exit cleanly
// so the OS service supervisor relaunches the new binary.
var ErrRestartRequired = errors.New("agent: restart required after self-update")

// Agent runs enrollment, heartbeat, and posture sync loops.
type Agent struct {
	Dir       string
	Version   string
	UserAgent string
	Logger    *log.Logger

	// Updater performs binary self-update. When nil, auto-update is
	// disabled (e.g. dev builds, --no-auto-update at install time).
	Updater *update.Updater

	cfg     *Config
	client  *Client
	revoked bool

	collectHostInfo     func() HostInfo
	hostInfo            HostInfo
	hostInfoCollectedAt time.Time

	now        func() time.Time
	randInt63n func(int64) int64

	pendingFlushBackoff time.Duration
	pendingFlushRetryAt time.Time
}

// New creates an agent instance.
func New(dir, version string, logger *log.Logger) *Agent {
	if logger == nil {
		logger = log.NewLogger(log.WithName("device-agent"))
	}

	return &Agent{
		Dir:       dir,
		Version:   version,
		UserAgent: fmt.Sprintf("probo-agent/%s", version),
		Logger:    logger,
		collectHostInfo: func() HostInfo {
			return CollectHostInfo()
		},
		now: time.Now,
		randInt63n: func(n int64) int64 {
			return rand.Int63n(n)
		},
	}
}

// EnrollNewDevice enrolls and persists local config and key state.
func (a *Agent) EnrollNewDevice(
	ctx context.Context,
	serverURL, enrollmentToken string,
) (*EnrollResponse, error) {
	if serverURL == "" {
		return nil, errors.New("server URL is required")
	}

	if enrollmentToken == "" {
		return nil, errors.New("enrollment token is required")
	}

	host := a.currentHostInfo(time.Now())
	client := NewClient(serverURL, "", a.UserAgent)

	resp, err := client.Enroll(
		ctx,
		EnrollRequest{
			EnrollmentToken: enrollmentToken,
			HardwareUUID:    host.HardwareUUID,
			SerialNumber:    host.SerialNumber,
			Hostname:        host.Hostname,
			Platform:        host.Platform,
			OSVersion:       host.OSVersion,
			AgentVersion:    a.Version,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot enroll device: %w", err)
	}

	cfg := &Config{
		ServerURL:         serverURL,
		DeviceID:          resp.DeviceID,
		HeartbeatInterval: time.Duration(resp.HeartbeatSeconds) * time.Second,
		PostureInterval:   time.Duration(resp.PostureSeconds) * time.Second,
	}
	if err := SaveConfig(a.Dir, cfg); err != nil {
		return nil, fmt.Errorf("cannot save config: %w", err)
	}

	if err := SaveAPIKey(a.Dir, resp.APIKey); err != nil {
		return nil, fmt.Errorf("cannot save api key: %w", err)
	}

	if err := clearPendingPostureBatches(a.Dir); err != nil {
		a.Logger.Warn("cannot clear pending posture queue after enrollment", log.Error(err))
	}

	a.cfg = cfg
	a.client = NewClient(serverURL, resp.APIKey, a.UserAgent)

	return resp, nil
}

// LoadLocalState loads persisted config and API key.
func (a *Agent) LoadLocalState() error {
	cfg, err := LoadConfig(a.Dir)
	if err != nil {
		return err
	}

	key, err := LoadAPIKey(a.Dir)
	if err != nil {
		return err
	}

	if cfg.ServerURL == "" {
		return errors.New("config has no server URL")
	}

	a.cfg = cfg
	a.client = NewClient(cfg.ServerURL, key, a.UserAgent)

	return nil
}

// Run starts the long-running heartbeat and posture loops. It returns
// ErrRestartRequired after a successful self-update so the caller can
// exit cleanly and let the service supervisor restart the new binary.
func (a *Agent) Run(ctx context.Context) error {
	if a.cfg == nil || a.client == nil {
		if err := a.LoadLocalState(); err != nil {
			return fmt.Errorf("cannot load agent state: %w", err)
		}
	}

	a.Logger = a.Logger.With(log.String("device_id", a.cfg.DeviceID))

	a.Logger.InfoCtx(
		ctx,
		"agent starting",
		log.String("server", a.cfg.ServerURL),
		log.Duration("heartbeat_interval", a.cfg.HeartbeatInterval),
		log.Duration("posture_interval", a.cfg.PostureInterval),
		log.Duration("host_info_refresh_interval", hostInfoRefreshInterval),
		log.Bool("auto_update_enabled", a.autoUpdateEnabled()),
		log.Duration("update_interval", a.cfg.UpdateInterval),
	)

	_, _ = a.doHeartbeat(ctx)
	a.doPostures(ctx)

	heartbeatTicker := time.NewTicker(a.cfg.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	postureTicker := time.NewTicker(a.cfg.PostureInterval)
	defer postureTicker.Stop()

	updateTicker, updateChan := a.newUpdateTicker()
	if updateTicker != nil {
		defer updateTicker.Stop()
	}

	for {
		select {
		case <-ctx.Done():
			a.Logger.InfoCtx(ctx, "agent stopping", log.Error(ctx.Err()))
			return ctx.Err()
		case <-heartbeatTicker.C:
			heartbeatIntervalChanged, postureIntervalChanged := a.doHeartbeat(ctx)
			if heartbeatIntervalChanged {
				heartbeatTicker.Reset(a.cfg.HeartbeatInterval)
			}

			if postureIntervalChanged {
				postureTicker.Reset(a.cfg.PostureInterval)
			}
		case <-postureTicker.C:
			a.doPostures(ctx)
		case <-updateChan:
			if a.tryAutoUpdate(ctx) {
				return ErrRestartRequired
			}
		}
	}
}

// autoUpdateEnabled reports whether the agent should periodically
// self-update. Disabled when no Updater is wired in or the operator
// flipped UpdatesDisabled in config.
func (a *Agent) autoUpdateEnabled() bool {
	if a.cfg == nil {
		return false
	}

	if a.cfg.UpdatesDisabled {
		return false
	}

	return a.Updater != nil
}

// newUpdateTicker returns the periodic ticker used to drive
// auto-update checks. When auto-update is disabled we return a
// (nil, nil) channel pair so the select in Run never fires.
func (a *Agent) newUpdateTicker() (*time.Ticker, <-chan time.Time) {
	if !a.autoUpdateEnabled() {
		return nil, nil
	}

	t := time.NewTicker(a.cfg.UpdateInterval)

	return t, t.C
}

// tryAutoUpdate runs one auto-update cycle and returns true when the
// agent binary was successfully replaced and the process should
// restart.
func (a *Agent) tryAutoUpdate(parent context.Context) bool {
	if !a.autoUpdateEnabled() {
		return false
	}

	ctx, cancel := context.WithTimeout(parent, updateCheckTimeout)
	defer cancel()

	rel, err := a.Updater.CheckLatest(ctx)
	if err != nil {
		if errors.Is(err, update.ErrNoUpdateAvailable) {
			a.Logger.DebugCtx(ctx, "no agent update available")
			return false
		}

		a.Logger.WarnCtx(ctx, "agent update check failed", log.Error(err))

		return false
	}

	a.Logger.InfoCtx(
		ctx,
		"agent update available, applying",
		log.String("from_version", a.Version),
		log.String("to_version", rel.Version),
	)

	if err := a.Updater.Apply(ctx, rel); err != nil {
		a.Logger.ErrorCtx(ctx, "cannot apply agent update", log.Error(err), log.String("to_version", rel.Version))
		return false
	}

	return true
}

// CollectOnce executes checks without pushing results to the server.
func (a *Agent) CollectOnce(ctx context.Context) []checks.Result {
	now := time.Now()
	results := make([]checks.Result, 0)

	for _, c := range checks.All() {
		select {
		case <-ctx.Done():
			return results
		default:
		}

		checkCtx, cancel := context.WithTimeout(ctx, perCheckTimeout)
		r := c.Run(checkCtx)

		cancel()

		if r.ObservedAt.IsZero() {
			r.ObservedAt = now
		}

		if r.CheckKey == "" {
			r.CheckKey = c.Key()
		}

		results = append(results, r)
	}

	return results
}

// Unenroll revokes server state best-effort and clears local credentials.
func (a *Agent) Unenroll(ctx context.Context) error {
	if a.cfg == nil || a.client == nil {
		if err := a.LoadLocalState(); err != nil {
			return err
		}
	}

	if err := a.client.Unenroll(ctx); err != nil {
		a.Logger.WarnCtx(
			ctx,
			"unenroll server-side revocation failed, continuing with local wipe",
			log.Error(err),
		)
	}

	if err := DeleteAPIKey(a.Dir); err != nil {
		return err
	}

	if err := clearPendingPostureBatches(a.Dir); err != nil {
		return err
	}

	return nil
}

func (a *Agent) doHeartbeat(ctx context.Context) (bool, bool) {
	if a.revoked {
		return false, false
	}

	oldHeartbeatInterval := a.cfg.HeartbeatInterval
	oldPostureInterval := a.cfg.PostureInterval

	host := a.currentHostInfo(time.Now())

	resp, err := a.client.Heartbeat(
		ctx,
		HeartbeatRequest{
			AgentVersion: a.Version,
			Hostname:     host.Hostname,
			OSVersion:    host.OSVersion,
		},
	)
	if err != nil {
		a.Logger.ErrorCtx(ctx, "heartbeat failed", log.Error(err))

		if IsUnauthorized(err) {
			a.handleUnauthorized()
		}

		return false, false
	}

	if resp.HeartbeatSeconds > 0 {
		next := normalizeHeartbeatInterval(time.Duration(resp.HeartbeatSeconds) * time.Second)
		if next != a.cfg.HeartbeatInterval {
			a.cfg.HeartbeatInterval = next
		}
	}

	if resp.PostureSeconds > 0 {
		next := normalizePostureInterval(time.Duration(resp.PostureSeconds) * time.Second)
		if next != a.cfg.PostureInterval {
			a.cfg.PostureInterval = next
		}
	}

	a.flushQueuedPostures(ctx)

	heartbeatChanged := a.cfg.HeartbeatInterval != oldHeartbeatInterval

	postureChanged := a.cfg.PostureInterval != oldPostureInterval
	if heartbeatChanged || postureChanged {
		if err := SaveConfig(a.Dir, a.cfg); err != nil {
			a.Logger.WarnCtx(
				ctx,
				"cannot persist updated agent intervals",
				log.Error(err),
			)
		}
	}

	return heartbeatChanged, postureChanged
}

func (a *Agent) doPostures(ctx context.Context) {
	if a.revoked {
		return
	}

	start := time.Now()

	results := a.CollectOnce(ctx)
	if len(results) == 0 {
		return
	}

	var (
		passCount          int
		failCount          int
		unknownCount       int
		notApplicableCount int
	)

	for _, r := range results {
		switch r.Status {
		case checks.StatusPass:
			passCount++
		case checks.StatusFail:
			failCount++
		case checks.StatusUnknown:
			unknownCount++
		case checks.StatusNotApplicable:
			notApplicableCount++
		}
	}

	a.Logger.InfoCtx(
		ctx,
		"posture checks completed",
		log.Int("checks", len(results)),
		log.Int("pass_count", passCount),
		log.Int("fail_count", failCount),
		log.Int("unknown_count", unknownCount),
		log.Int("not_applicable_count", notApplicableCount),
		log.Duration("elapsed", time.Since(start)),
		log.Duration("per_check_timeout", perCheckTimeout),
	)

	payload := make([]PostureResultPayload, 0, len(results))
	for _, r := range results {
		payload = append(
			payload,
			PostureResultPayload{
				CheckKey:   r.CheckKey,
				Status:     string(r.Status),
				Evidence:   checks.EvidenceJSON(r.Evidence),
				ObservedAt: r.ObservedAt,
			},
		)
	}

	a.flushQueuedPostures(ctx)

	if a.revoked {
		return
	}

	if err := a.client.PushPostures(ctx, payload); err != nil {
		a.Logger.ErrorCtx(ctx, "posture push failed", log.Error(err))

		if IsUnauthorized(err) {
			a.handleUnauthorized()
			return
		}

		dropped, enqueueErr := enqueuePendingPostureBatch(a.Dir, payload, a.currentTime())
		if enqueueErr != nil {
			a.Logger.ErrorCtx(ctx, "cannot queue posture batch after failed push", log.Error(enqueueErr))
			return
		}

		a.Logger.WarnCtx(
			ctx,
			"queued posture batch for retry",
			log.Int("queued_results", len(payload)),
			log.Int("dropped_old_batches", dropped),
		)
	}
}

func (a *Agent) flushQueuedPostures(ctx context.Context) {
	if a.revoked || a.client == nil {
		return
	}

	now := a.currentTime()
	if !a.pendingFlushRetryAt.IsZero() && now.Before(a.pendingFlushRetryAt) {
		return
	}

	batches, err := loadPendingPostureBatches(a.Dir)
	if err != nil {
		a.Logger.WarnCtx(ctx, "cannot load pending posture batches", log.Error(err))
		return
	}

	if len(batches) == 0 {
		a.resetPendingFlushRetry()
		return
	}

	for i, batch := range batches {
		if err := a.client.PushPostures(ctx, batch.Results); err != nil {
			if IsUnauthorized(err) {
				a.handleUnauthorized()
				return
			}

			if saveErr := savePendingPostureBatches(a.Dir, batches[i:]); saveErr != nil {
				a.Logger.ErrorCtx(ctx, "cannot persist pending posture batches", log.Error(saveErr))
			}

			retryIn := a.schedulePendingFlushRetry(now)
			a.Logger.WarnCtx(
				ctx,
				"cannot flush pending posture batch",
				log.Error(err),
				log.Int("remaining_batches", len(batches)-i),
				log.Duration("retry_in", retryIn),
			)

			return
		}
	}

	if err := clearPendingPostureBatches(a.Dir); err != nil {
		a.Logger.ErrorCtx(ctx, "cannot clear pending posture batches", log.Error(err))
		return
	}

	a.resetPendingFlushRetry()
	a.Logger.InfoCtx(ctx, "flushed pending posture batches", log.Int("batches", len(batches)))
}

func (a *Agent) currentHostInfo(now time.Time) HostInfo {
	if a.hostInfoCollectedAt.IsZero() || now.Sub(a.hostInfoCollectedAt) >= hostInfoRefreshInterval {
		collector := a.collectHostInfo
		if collector == nil {
			collector = CollectHostInfo
		}

		a.hostInfo = collector()
		a.hostInfoCollectedAt = now
	}

	return a.hostInfo
}

func (a *Agent) currentTime() time.Time {
	if a.now != nil {
		return a.now()
	}

	return time.Now()
}

func (a *Agent) randomInt63n(n int64) int64 {
	if n <= 1 {
		return 0
	}

	if a.randInt63n != nil {
		return a.randInt63n(n)
	}

	return rand.Int63n(n)
}

func (a *Agent) schedulePendingFlushRetry(now time.Time) time.Duration {
	nextBase := a.pendingFlushBackoff
	if nextBase <= 0 {
		nextBase = pendingFlushBackoffMin
	} else {
		nextBase *= 2
		if nextBase > pendingFlushBackoffMax {
			nextBase = pendingFlushBackoffMax
		}
	}

	a.pendingFlushBackoff = nextBase

	jitterRange := nextBase / 5

	jitter := time.Duration(0)
	if jitterRange > 0 {
		jitter = time.Duration(a.randomInt63n(int64(jitterRange)*2+1)) - jitterRange
	}

	retryIn := max(nextBase+jitter, time.Second)
	a.pendingFlushRetryAt = now.Add(retryIn)

	return retryIn
}

func (a *Agent) resetPendingFlushRetry() {
	a.pendingFlushBackoff = 0
	a.pendingFlushRetryAt = time.Time{}
}

// handleUnauthorized wipes local auth state after a 401 response.
func (a *Agent) handleUnauthorized() {
	if a.revoked {
		return
	}

	a.revoked = true
	if a.client != nil {
		a.client.APIKey = ""
	}

	a.Logger.Warn("agent API returned 401, wiping local key and requiring re-enrollment")

	if err := DeleteAPIKey(a.Dir); err != nil {
		a.Logger.Error("cannot delete local key after 401", log.Error(err))
	}

	if err := clearPendingPostureBatches(a.Dir); err != nil {
		a.Logger.Error("cannot delete pending posture queue after 401", log.Error(err))
	}

	a.resetPendingFlushRetry()
}
