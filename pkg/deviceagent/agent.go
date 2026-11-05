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

package deviceagent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/deviceagent/checks"
	"go.probo.inc/probo/pkg/deviceagent/update"
	"go.probo.inc/probo/pkg/gid"
)

const (
	hostInfoRefreshInterval = 6 * time.Hour
	perCheckTimeout         = 150 * time.Second

	pendingFlushBackoffMin = 15 * time.Second
	pendingFlushBackoffMax = 30 * time.Minute

	updateCheckTimeout = 10 * time.Minute
)

// ErrRestartRequired is returned by Agent.Run after a successful
// in-place upgrade of the agent binary. Callers should exit cleanly
// so the OS service supervisor relaunches the new binary.
var ErrRestartRequired = errors.New("agent: restart required after self-update")

type trackedEnrollment struct {
	Enrollment
	client *Client
}

// Agent runs enrollment, heartbeat, and posture sync loops.
type Agent struct {
	Dir       string
	Version   string
	UserAgent string
	Logger    *log.Logger

	// Updater performs binary self-update. When nil, auto-update is
	// disabled (e.g. dev builds, --no-auto-update at install time).
	Updater *update.Updater

	cfg         *Config
	enrollments []trackedEnrollment

	collectHostInfo     func() HostInfo
	checkSet            func() []checks.Check
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
		checkSet: checks.All,
		now:      time.Now,
		randInt63n: func(n int64) int64 {
			return rand.Int63n(n)
		},
	}
}

// AddEnrollment stores an API key and activates that organization
// enrollment with a heartbeat. A retryable heartbeat failure keeps the
// key so the one-shot token is not lost. A hardware conflict or 401
// removes the key.
func (a *Agent) AddEnrollment(
	ctx context.Context,
	serverURL, apiKey string,
) (*HeartbeatResponse, error) {
	if serverURL == "" {
		return nil, errors.New("server URL is required")
	}

	if apiKey == "" {
		return nil, errors.New("api key is required")
	}

	normalized, err := NormalizeServerURL(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	if err := ensureSecureAgentDir(a.Dir); err != nil {
		return nil, err
	}

	items, err := LoadEnrollments(a.Dir)
	if err != nil {
		return nil, err
	}

	cfg, err := loadOrCreateConfig(a.Dir)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(items, func(item Enrollment) bool {
		return item.ServerURL == normalized && item.APIKey == apiKey
	})
	if idx < 0 {
		items = append(items, Enrollment{ServerURL: normalized, APIKey: apiKey})

		idx = len(items) - 1
		if err := SaveEnrollments(a.Dir, items); err != nil {
			return nil, fmt.Errorf("cannot save enrollments: %w", err)
		}
	}

	client := NewClient(normalized, apiKey, a.UserAgent)
	host := a.currentHostInfo(time.Now())

	resp, err := client.Heartbeat(ctx, a.heartbeatRequest(host))
	if err != nil {
		if IsHardwareConflict(err) || IsUnauthorized(err) {
			items = slices.Delete(items, idx, idx+1)
			if saveErr := SaveEnrollments(a.Dir, items); saveErr != nil {
				return nil, fmt.Errorf("cannot remove enrollment: %w", saveErr)
			}

			if IsHardwareConflict(err) {
				return nil, ErrAlreadyEnrolledInOrganization
			}

			return nil, fmt.Errorf("cannot activate device: %w", err)
		}

		a.cfg = cfg
		a.setEnrollments(items)

		return nil, fmt.Errorf("cannot activate device: %w", err)
	}

	items[idx].DeviceID = resp.DeviceID
	if err := SaveEnrollments(a.Dir, items); err != nil {
		return nil, fmt.Errorf("cannot save enrollment device id: %w", err)
	}

	cfg.HeartbeatInterval = serverInterval(
		cfg.HeartbeatInterval,
		time.Duration(resp.HeartbeatSeconds)*time.Second,
		MinHeartbeatInterval,
	)
	cfg.PostureInterval = serverInterval(
		cfg.PostureInterval,
		time.Duration(resp.PostureSeconds)*time.Second,
		MinPostureInterval,
	)

	cfg.ServerURL = ""
	cfg.DeviceID = ""

	if err := SaveConfig(a.Dir, cfg); err != nil {
		return nil, fmt.Errorf("cannot save config: %w", err)
	}

	if err := syncPublicEnrollmentState(a.Dir, items); err != nil {
		return nil, fmt.Errorf("cannot update enrollment state: %w", err)
	}

	a.cfg = cfg
	a.setEnrollments(items)

	return resp, nil
}

// LoadLocalState loads persisted config and enrollments.
func (a *Agent) LoadLocalState() error {
	if err := ensureSecureAgentDir(a.Dir); err != nil {
		return err
	}

	items, err := LoadEnrollments(a.Dir)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return errors.New("no enrollments")
	}

	cfg, err := LoadConfig(a.Dir)
	if err != nil {
		return err
	}

	a.cfg = cfg
	a.setEnrollments(items)

	if len(a.enrollments) == 0 {
		return errors.New("no enrollments")
	}

	return nil
}

func (a *Agent) setEnrollments(items []Enrollment) {
	tracked := make([]trackedEnrollment, 0, len(items))
	for _, item := range items {
		if item.ServerURL == "" || item.APIKey == "" {
			continue
		}

		tracked = append(tracked, trackedEnrollment{
			Enrollment: item,
			client:     NewClient(item.ServerURL, item.APIKey, a.UserAgent),
		})
	}

	a.enrollments = tracked
}

func (a *Agent) enrollmentItems() []Enrollment {
	items := make([]Enrollment, len(a.enrollments))
	for i, enrollment := range a.enrollments {
		items[i] = enrollment.Enrollment
	}

	return items
}

func (a *Agent) dropEnrollment(apiKey string) {
	if apiKey == "" {
		return
	}

	var deviceID string

	kept := make([]Enrollment, 0, len(a.enrollments))
	for _, enrollment := range a.enrollments {
		if enrollment.APIKey == apiKey {
			deviceID = enrollment.DeviceID
			continue
		}

		kept = append(kept, enrollment.Enrollment)
	}

	a.setEnrollments(kept)

	if deviceID != "" {
		if err := clearPendingPostureBatches(a.Dir, deviceID); err != nil {
			a.Logger.Error(
				"cannot delete pending posture queue",
				log.Error(err),
				log.String("device_id", deviceID),
			)
		}
	}

	if err := SaveEnrollments(a.Dir, kept); err != nil {
		a.Logger.Error("cannot save enrollments", log.Error(err))
	}

	if err := syncPublicEnrollmentState(a.Dir, kept); err != nil {
		a.Logger.Error("cannot update enrollment state", log.Error(err))
	}
}

// Run starts the long-running heartbeat and posture loops. It returns
// ErrRestartRequired after a successful self-update so the caller can
// exit cleanly and let the service supervisor restart the new binary.
func (a *Agent) Run(ctx context.Context) error {
	if a.cfg == nil || len(a.enrollments) == 0 {
		if err := a.LoadLocalState(); err != nil {
			return fmt.Errorf("cannot load agent state: %w", err)
		}
	}

	if err := syncPublicEnrollmentState(a.Dir, a.enrollmentItems()); err != nil {
		return fmt.Errorf("cannot sync enrollment marker: %w", err)
	}

	a.Logger.InfoCtx(
		ctx,
		"agent starting",
		log.Int("enrollments", len(a.enrollments)),
		log.Duration("heartbeat_interval", a.cfg.HeartbeatInterval),
		log.Duration("posture_interval", a.cfg.PostureInterval),
		log.Duration("host_info_refresh_interval", hostInfoRefreshInterval),
		log.Bool("auto_update_enabled", a.autoUpdateEnabled()),
		log.Bool("allow_prereleases", a.cfg.AllowPrereleases),
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

// CollectOnce executes checks without pushing results to the server. A
// non-nil error means the run was cut short: the set may be incomplete, and a
// check caught by the cancellation reports a failure of the shutdown rather
// than of the host, so callers must not persist the results as a report.
func (a *Agent) CollectOnce(ctx context.Context) ([]checks.Result, error) {
	now := time.Now()
	all := a.checks()
	results := make([]checks.Result, 0, len(all))

	for _, c := range all {
		select {
		case <-ctx.Done():
			return results, fmt.Errorf("cannot complete posture collection: %w", ctx.Err())
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

	// A check already running when the context ends still returns, so the loop
	// can finish on its own with only that last result poisoned by the
	// shutdown. Counting the results cannot catch it; the context can.
	if err := ctx.Err(); err != nil {
		return results, fmt.Errorf("cannot complete posture collection: %w", err)
	}

	return a.rememberChecks(ctx, results), nil
}

func (a *Agent) rememberChecks(ctx context.Context, results []checks.Result) []checks.Result {
	memory, err := loadCheckMemory(a.Dir)
	if err != nil {
		a.Logger.WarnCtx(ctx, "cannot load check memory", log.Error(err))

		// An unreadable file must not be replaced with a partial map.
		return results
	}

	if applyCheckMemory(results, memory) {
		if err := saveCheckMemory(a.Dir, memory); err != nil {
			a.Logger.WarnCtx(ctx, "cannot persist check memory", log.Error(err))
		}
	}

	return results
}

func (a *Agent) checks() []checks.Check {
	if a.checkSet == nil {
		return checks.All()
	}

	return a.checkSet()
}

// Unenroll asks the server to revoke this device. Local wipe is left to
// RemoveLocalState (called by uninstall after service/tray teardown).
// Missing credentials or server failures are logged and ignored so
// uninstall can continue.
func (a *Agent) Unenroll(ctx context.Context) error {
	if a.cfg == nil || len(a.enrollments) == 0 {
		if err := a.LoadLocalState(); err != nil {
			a.Logger.WarnCtx(ctx, "cannot load local state for unenroll", log.Error(err))
			return nil
		}
	}

	for _, enrollment := range a.enrollments {
		if err := enrollment.client.Unenroll(ctx); err != nil {
			a.Logger.WarnCtx(
				ctx,
				"unenroll server-side revocation failed",
				log.Error(err),
				log.String("device_id", enrollment.DeviceID),
			)
		}
	}

	return nil
}

func (a *Agent) doHeartbeat(ctx context.Context) (bool, bool) {
	if len(a.enrollments) == 0 {
		return false, false
	}

	oldHeartbeatInterval := a.cfg.HeartbeatInterval
	oldPostureInterval := a.cfg.PostureInterval

	host := a.currentHostInfo(time.Now())
	req := a.heartbeatRequest(host)

	var (
		dropped       []string
		identityDirty bool
	)

	for i := range a.enrollments {
		enrollment := &a.enrollments[i]

		resp, err := enrollment.client.Heartbeat(ctx, req)
		if err != nil {
			a.Logger.ErrorCtx(
				ctx,
				"heartbeat failed",
				log.Error(err),
				log.String("device_id", enrollment.DeviceID),
				log.String("server", enrollment.ServerURL),
			)

			if IsUnauthorized(err) || IsHardwareConflict(err) {
				dropped = append(dropped, enrollment.APIKey)
			}

			continue
		}

		if resp.DeviceID != "" && enrollment.DeviceID != resp.DeviceID {
			enrollment.DeviceID = resp.DeviceID
			identityDirty = true
		}

		a.cfg.HeartbeatInterval = serverInterval(
			a.cfg.HeartbeatInterval,
			time.Duration(resp.HeartbeatSeconds)*time.Second,
			MinHeartbeatInterval,
		)
		a.cfg.PostureInterval = serverInterval(
			a.cfg.PostureInterval,
			time.Duration(resp.PostureSeconds)*time.Second,
			MinPostureInterval,
		)
	}

	if identityDirty && len(dropped) == 0 {
		if err := SaveEnrollments(a.Dir, a.enrollmentItems()); err != nil {
			a.Logger.WarnCtx(ctx, "cannot persist enrollment device ids", log.Error(err))
		}
	}

	for _, apiKey := range dropped {
		a.dropEnrollment(apiKey)
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
	if len(a.enrollments) == 0 {
		return
	}

	start := time.Now()

	results, err := a.CollectOnce(ctx)
	if err != nil {
		a.Logger.WarnCtx(
			ctx,
			"discarding truncated posture run",
			log.Error(err),
			log.Int("collected_checks", len(results)),
			log.Int("expected_checks", len(a.checks())),
		)

		return
	}

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

	a.flushQueuedPostures(ctx)

	enrollments := append([]trackedEnrollment(nil), a.enrollments...)
	for _, enrollment := range enrollments {
		if enrollment.DeviceID == "" || enrollment.client == nil {
			continue
		}

		payload, err := posturePayload(enrollment.DeviceID, results)
		if err != nil {
			a.Logger.ErrorCtx(
				ctx,
				"cannot build posture payload",
				log.Error(err),
				log.String("device_id", enrollment.DeviceID),
			)

			continue
		}

		if err := enrollment.client.PushPostures(
			ctx,
			PosturesRequest{AgentVersion: a.Version, Results: payload},
		); err != nil {
			a.Logger.ErrorCtx(
				ctx,
				"posture push failed",
				log.Error(err),
				log.String("device_id", enrollment.DeviceID),
			)

			if IsUnauthorized(err) {
				a.dropEnrollment(enrollment.APIKey)
				continue
			}

			dropped, enqueueErr := enqueuePendingPostureBatch(
				a.Dir,
				enrollment.DeviceID,
				a.Version,
				payload,
				a.currentTime(),
			)
			if enqueueErr != nil {
				a.Logger.ErrorCtx(
					ctx,
					"cannot queue posture batch after failed push",
					log.Error(enqueueErr),
					log.String("device_id", enrollment.DeviceID),
				)

				continue
			}

			a.Logger.WarnCtx(
				ctx,
				"queued posture batch for retry",
				log.Int("queued_results", len(payload)),
				log.Int("dropped_old_batches", dropped),
				log.String("device_id", enrollment.DeviceID),
			)
		}
	}
}

func posturePayload(deviceID string, results []checks.Result) ([]PostureResultPayload, error) {
	parsed, err := gid.ParseGID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("cannot parse device id: %w", err)
	}

	correlationID := gid.New(
		parsed.TenantID(),
		coredata.DevicePostureReportEntityType,
	).String()

	payload := make([]PostureResultPayload, 0, len(results))
	for _, result := range results {
		payload = append(
			payload,
			PostureResultPayload{
				CheckKey:      result.CheckKey,
				Status:        string(result.Status),
				Evidence:      checks.EvidenceJSON(result.Evidence),
				ObservedAt:    result.ObservedAt,
				CorrelationID: correlationID,
			},
		)
	}

	return payload, nil
}

func (a *Agent) flushQueuedPostures(ctx context.Context) {
	if len(a.enrollments) == 0 {
		return
	}

	now := a.currentTime()
	if !a.pendingFlushRetryAt.IsZero() && now.Before(a.pendingFlushRetryAt) {
		return
	}

	queues, err := loadPostureQueues(a.Dir)
	if err != nil {
		a.Logger.WarnCtx(ctx, "cannot load pending posture batches", log.Error(err))
		return
	}

	if len(queues) == 0 {
		a.resetPendingFlushRetry()
		return
	}

	failed := false
	flushed := 0

	enrollments := append([]trackedEnrollment(nil), a.enrollments...)
	for _, enrollment := range enrollments {
		if enrollment.DeviceID == "" || enrollment.client == nil {
			continue
		}

		batches := queues[enrollment.DeviceID]
		if len(batches) == 0 {
			continue
		}

		remainingFrom := -1
		unauthorized := false

		for i, batch := range batches {
			err := enrollment.client.PushPostures(
				ctx,
				PosturesRequest{AgentVersion: batch.AgentVersion, Results: batch.Results},
			)
			if err != nil {
				if IsUnauthorized(err) {
					unauthorized = true
					break
				}

				remainingFrom = i
				failed = true

				a.Logger.WarnCtx(
					ctx,
					"cannot flush pending posture batch",
					log.Error(err),
					log.String("device_id", enrollment.DeviceID),
					log.Int("remaining_batches", len(batches)-i),
				)

				break
			}

			flushed++
		}

		if unauthorized {
			a.dropEnrollment(enrollment.APIKey)
			continue
		}

		if remainingFrom >= 0 {
			if saveErr := savePendingPostureBatches(a.Dir, enrollment.DeviceID, batches[remainingFrom:]); saveErr != nil {
				a.Logger.ErrorCtx(
					ctx,
					"cannot persist pending posture batches",
					log.Error(saveErr),
					log.String("device_id", enrollment.DeviceID),
				)
			}

			continue
		}

		if err := clearPendingPostureBatches(a.Dir, enrollment.DeviceID); err != nil {
			a.Logger.ErrorCtx(
				ctx,
				"cannot clear pending posture batches",
				log.Error(err),
				log.String("device_id", enrollment.DeviceID),
			)
		}
	}

	if failed {
		retryIn := a.schedulePendingFlushRetry(now)
		a.Logger.WarnCtx(ctx, "posture flush will retry", log.Duration("retry_in", retryIn))

		return
	}

	a.resetPendingFlushRetry()

	if flushed > 0 {
		a.Logger.InfoCtx(ctx, "flushed pending posture batches", log.Int("batches", flushed))
	}
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

func (a *Agent) heartbeatRequest(host HostInfo) HeartbeatRequest {
	return HeartbeatRequest{
		HardwareUUID: host.HardwareUUID,
		SerialNumber: host.SerialNumber,
		Hostname:     host.Hostname,
		Platform:     host.Platform,
		OSVersion:    host.OSVersion,
		AgentVersion: a.Version,
	}
}
