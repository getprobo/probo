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

package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type (
	webhookHandler struct {
		pg             *pg.Client
		logger         *log.Logger
		httpClient     *http.Client
		encryptionKey  cipher.EncryptionKey
		host           string
		cache          sync.Map
		cacheMu        sync.Mutex
		cacheCreatedAt time.Time
		cacheTTL       time.Duration
		timeout        time.Duration
		staleAfter     time.Duration
		retryBase      time.Duration
		retryMax       time.Duration
		now            func() time.Time
		jitter         func(time.Duration) time.Duration
	}

	cachedSecret struct {
		encryptedSecret []byte
		plaintext       string
	}

	Config struct {
		Interval       time.Duration
		Timeout        time.Duration
		CacheTTL       time.Duration
		StaleAfter     time.Duration
		RetryBase      time.Duration
		RetryMax       time.Duration
		MaxConcurrency int
		EncryptionKey  cipher.EncryptionKey
		Host           string
	}

	webhookTask struct {
		event        coredata.WebhookEvent
		webhookData  coredata.WebhookData
		subscription coredata.WebhookSubscription
	}

	webhookDeliveryError struct {
		message    string
		transient  bool
		retryAfter *time.Duration
	}
)

const (
	defaultTimeout        = 15 * time.Second
	defaultStaleAfter     = 5 * time.Minute
	defaultRetryBase      = 30 * time.Second
	defaultRetryMax       = 4 * time.Hour
	defaultMaxConcurrency = 5
	maxResponseBodySize   = 64 * 1024
)

var (
	_ worker.Handler[webhookTask] = (*webhookHandler)(nil)
	_ worker.StaleRecoverer       = (*webhookHandler)(nil)
)

func (e *webhookDeliveryError) Error() string {
	return e.message
}

func NewWebhookWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	cfg Config,
	opts ...worker.Option,
) *worker.Worker[webhookTask] {
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 24 * time.Hour
	}
	if cfg.StaleAfter <= 0 {
		cfg.StaleAfter = defaultStaleAfter
	}
	if cfg.RetryBase <= 0 {
		cfg.RetryBase = defaultRetryBase
	}
	if cfg.RetryMax <= 0 {
		cfg.RetryMax = defaultRetryMax
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = defaultMaxConcurrency
	}

	h := &webhookHandler{
		pg:             pgClient,
		logger:         logger,
		httpClient:     httpclient.DefaultPooledClient(httpclient.WithLogger(logger), httpclient.WithSSRFProtection()),
		encryptionKey:  cfg.EncryptionKey,
		host:           cfg.Host,
		cacheCreatedAt: time.Now(),
		cacheTTL:       cfg.CacheTTL,
		timeout:        cfg.Timeout,
		staleAfter:     cfg.StaleAfter,
		retryBase:      cfg.RetryBase,
		retryMax:       cfg.RetryMax,
		now:            time.Now,
		jitter: func(max time.Duration) time.Duration {
			if max <= 0 {
				return 0
			}

			return time.Duration(rand.Int64N(int64(max) + 1))
		},
	}

	workerOpts := append(
		[]worker.Option{
			worker.WithInterval(cfg.Interval),
			worker.WithMaxConcurrency(cfg.MaxConcurrency),
		},
		opts...,
	)

	return worker.New("webhook-sender", h, logger, workerOpts...)
}

func (h *webhookHandler) Claim(ctx context.Context) (webhookTask, error) {
	h.refreshSecretCache()

	var task webhookTask
	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := h.materializeLegacyWebhookData(ctx, tx); err != nil {
				return fmt.Errorf("cannot materialize legacy webhook data: %w", err)
			}

			if err := task.event.ClaimNextForUpdateSkipLocked(ctx, tx, h.now()); err != nil {
				return fmt.Errorf("cannot claim next webhook event: %w", err)
			}

			scope := coredata.NewScopeFromObjectID(task.event.ID)
			if err := task.webhookData.LoadByID(ctx, tx, scope, task.event.WebhookDataID); err != nil {
				return fmt.Errorf("cannot load webhook data: %w", err)
			}
			if err := task.subscription.LoadByID(ctx, tx, scope, task.event.WebhookSubscriptionID); err != nil {
				return fmt.Errorf("cannot load webhook subscription: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return webhookTask{}, worker.ErrNoTask
		}

		return webhookTask{}, err
	}

	return task, nil
}

func (h *webhookHandler) Process(ctx context.Context, task webhookTask) error {
	response, deliveryErr := h.deliver(ctx, &task)
	now := h.now()

	task.event.Response = response
	task.event.ProcessingStartedAt = nil
	task.event.UpdatedAt = now
	if deliveryErr == nil {
		task.event.Status = coredata.WebhookEventStatusSucceeded
		task.event.CompletedAt = &now
		task.event.NextAttemptAt = nil
		task.event.LastError = nil
	} else {
		task.event.LastError = new(deliveryErr.Error())
		if deliveryErr.transient && task.event.AttemptCount < task.event.MaxAttempts {
			task.event.Status = coredata.WebhookEventStatusPending
			nextAttemptAt := now.Add(h.retryDelay(task.event.AttemptCount, deliveryErr))
			task.event.NextAttemptAt = &nextAttemptAt
		} else {
			task.event.Status = coredata.WebhookEventStatusFailed
			task.event.NextAttemptAt = nil
			task.event.DeadLetteredAt = &now
		}
	}

	scope := coredata.NewScopeFromObjectID(task.event.ID)
	err := h.pg.WithConn(
		context.WithoutCancel(ctx),
		func(ctx context.Context, conn pg.Querier) error {
			return task.event.UpdateDeliveryState(ctx, conn, scope)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrProcessingLeaseLost) {
			h.logger.InfoCtx(
				ctx,
				"lost webhook event processing lease",
				log.String("event_id", task.event.ID.String()),
			)

			return nil
		}

		return fmt.Errorf("cannot persist webhook delivery result: %w", err)
	}

	if deliveryErr != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot deliver webhook",
			log.String("event_id", task.event.ID.String()),
			log.String("webhook_data_id", task.webhookData.ID.String()),
			log.String("subscription_id", task.subscription.ID.String()),
			log.Int("attempt_count", task.event.AttemptCount),
			log.Bool("dead_lettered", task.event.DeadLetteredAt != nil),
			log.Error(deliveryErr),
		)

		return deliveryErr
	}

	h.logger.InfoCtx(
		ctx,
		"delivered webhook",
		log.String("event_id", task.event.ID.String()),
		log.Int("attempt_count", task.event.AttemptCount),
	)

	return nil
}

func (h *webhookHandler) RecoverStale(ctx context.Context) error {
	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleWebhookEvents(ctx, conn, h.now(), h.staleAfter)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot recover stale webhook events: %w", err)
	}

	return nil
}

func (h *webhookHandler) materializeLegacyWebhookData(ctx context.Context, tx pg.Tx) error {
	var webhookData coredata.WebhookData
	if err := webhookData.LoadNextUnprocessedForUpdate(ctx, tx); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load next unprocessed webhook data: %w", err)
	}

	scope := coredata.NewScopeFromObjectID(webhookData.ID)
	var subscriptions coredata.WebhookSubscriptions
	if err := subscriptions.LoadMatchingByOrganizationIDAndEventType(
		ctx,
		tx,
		scope,
		webhookData.OrganizationID,
		webhookData.EventType,
	); err != nil {
		return fmt.Errorf("cannot load matching webhook subscriptions: %w", err)
	}

	now := h.now()
	for _, subscription := range subscriptions {
		event := coredata.WebhookEvent{
			ID:                    gid.New(webhookData.ID.TenantID(), coredata.WebhookEventEntityType),
			WebhookDataID:         webhookData.ID,
			WebhookSubscriptionID: subscription.ID,
			Status:                coredata.WebhookEventStatusPending,
			MaxAttempts:           coredata.WebhookEventDefaultMaxAttempts,
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		if err := event.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert webhook event: %w", err)
		}
	}

	webhookData.ProcessedAt = &now
	if err := webhookData.UpdateProcessedAt(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot mark webhook data processed: %w", err)
	}

	return nil
}

func (h *webhookHandler) deliver(
	ctx context.Context,
	task *webhookTask,
) (json.RawMessage, *webhookDeliveryError) {
	signingSecret, err := h.getSigningSecret(
		task.subscription.ID.String(),
		task.subscription.EncryptedSigningSecret,
	)
	if err != nil {
		return nil, &webhookDeliveryError{message: "cannot decrypt webhook signing secret"}
	}

	return h.doHTTPCall(
		ctx,
		task.event.ID,
		task.subscription.EndpointURL,
		&task.webhookData,
		task.subscription.ID,
		signingSecret,
	)
}

func (h *webhookHandler) retryDelay(
	attempt int,
	deliveryErr *webhookDeliveryError,
) time.Duration {
	if deliveryErr.retryAfter != nil {
		delay := *deliveryErr.retryAfter
		if delay > h.retryMax {
			return h.retryMax
		}

		return delay
	}

	delay := h.retryBase
	for idx := 1; idx < attempt && delay < h.retryMax; idx++ {
		if delay > h.retryMax/2 {
			delay = h.retryMax
			break
		}

		delay *= 2
	}
	if delay > h.retryMax {
		delay = h.retryMax
	}

	return h.jitter(delay)
}

func (h *webhookHandler) refreshSecretCache() {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	if h.now().Sub(h.cacheCreatedAt) < h.cacheTTL {
		return
	}

	h.cache = sync.Map{}
	h.cacheCreatedAt = h.now()
}

func (h *webhookHandler) getSigningSecret(
	webhookSubscriptionID string,
	encryptedSigningSecret []byte,
) (string, error) {
	if cached, ok := h.cache.Load(webhookSubscriptionID); ok {
		entry := cached.(*cachedSecret)
		if bytes.Equal(entry.encryptedSecret, encryptedSigningSecret) {
			return entry.plaintext, nil
		}
	}

	plaintext, err := cipher.Decrypt(encryptedSigningSecret, h.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("cannot decrypt signing secret: %w", err)
	}

	signingSecret := string(plaintext)
	h.cache.Store(
		webhookSubscriptionID,
		&cachedSecret{
			encryptedSecret: encryptedSigningSecret,
			plaintext:       signingSecret,
		},
	)

	return signingSecret, nil
}

func (h *webhookHandler) doHTTPCall(
	ctx context.Context,
	eventID gid.GID,
	endpointURL string,
	webhookData *coredata.WebhookData,
	subscriptionID gid.GID,
	signingSecret string,
) (json.RawMessage, *webhookDeliveryError) {
	payload := Payload{
		EventID:        eventID.String(),
		SubscriptionID: subscriptionID.String(),
		OrganizationID: webhookData.OrganizationID.String(),
		EventType:      webhookData.EventType.String(),
		CreatedAt:      webhookData.CreatedAt,
		Data:           webhookData.Data,
		UpdatedFrom:    webhookData.UpdatedFrom,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &webhookDeliveryError{message: "cannot marshal webhook payload"}
	}

	reqCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, &webhookDeliveryError{message: "cannot create webhook request"}
	}

	timestamp := strconv.FormatInt(h.now().Unix(), 10)
	signature := computeSignature(signingSecret, timestamp, body)
	deliveryID := eventID.String()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", deliveryID)
	req.Header.Set("X-Probo-Webhook-Delivery-Id", deliveryID)
	req.Header.Set("X-Probo-Webhook-Event", webhookData.EventType.String())
	req.Header.Set("X-Probo-Webhook-Organization-Id", webhookData.OrganizationID.String())
	req.Header.Set("X-Probo-Webhook-Timestamp", timestamp)
	req.Header.Set("X-Probo-Webhook-Signature", signature)
	req.Header.Set("X-Probo-Webhook-Host", h.host)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, &webhookDeliveryError{
			message:   "webhook network request failed",
			transient: true,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	response := buildResponseJSON(resp, respBody)
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return response, nil
	}

	deliveryErr := &webhookDeliveryError{
		message:   fmt.Sprintf("webhook endpoint returned status %d", resp.StatusCode),
		transient: isRetryableStatus(resp.StatusCode),
	}
	if deliveryErr.transient {
		if retryAfter, ok := parseRetryAfter(resp.Header.Get("Retry-After"), h.now()); ok {
			deliveryErr.retryAfter = &retryAfter
		}
	}

	return response, deliveryErr
}

func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
		return true
	default:
		return statusCode >= http.StatusInternalServerError
	}
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds < 0 {
			return 0, false
		}

		return time.Duration(seconds) * time.Second, true
	}

	retryAt, err := http.ParseTime(value)
	if err != nil || retryAt.Before(now) {
		return 0, false
	}

	return retryAt.Sub(now), true
}

func buildResponseJSON(resp *http.Response, body []byte) json.RawMessage {
	headers := make(map[string]any, len(resp.Header))
	for k, v := range resp.Header {
		if len(v) == 1 {
			headers[k] = v[0]
		} else {
			headers[k] = v
		}
	}

	var bodyValue any
	if json.Valid(body) {
		bodyValue = json.RawMessage(body)
	} else {
		bodyValue = string(body)
	}

	respObj := map[string]any{
		"proto":       resp.Proto,
		"status_code": resp.StatusCode,
		"headers":     headers,
		"body":        bodyValue,
	}

	if len(resp.Trailer) > 0 {
		trailers := make(map[string]any, len(resp.Trailer))
		for k, v := range resp.Trailer {
			if len(v) == 1 {
				trailers[k] = v[0]
			} else {
				trailers[k] = v
			}
		}

		respObj["trailers"] = trailers
	}

	data, _ := json.Marshal(respObj)

	return data
}

func computeSignature(signingSecret, timestamp string, body []byte) string {
	h := hmac.New(sha256.New, []byte(signingSecret))
	_, _ = fmt.Fprintf(h, "%s:%s", timestamp, body)

	return hex.EncodeToString(h.Sum(nil))
}
