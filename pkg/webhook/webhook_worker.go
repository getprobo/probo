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
	"net/http"
	"strconv"
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
		cacheCreatedAt time.Time
		cacheTTL       time.Duration
		timeout        time.Duration
	}

	cachedSecret struct {
		encryptedSecret []byte
		plaintext       string
	}

	Config struct {
		Interval      time.Duration
		Timeout       time.Duration
		CacheTTL      time.Duration
		EncryptionKey cipher.EncryptionKey
		Host          string
	}

	pendingDelivery struct {
		Event  *coredata.WebhookEvent
		Config *coredata.WebhookSubscription
	}

	webhookTask struct {
		webhookData *coredata.WebhookData
		deliveries  []pendingDelivery
	}
)

const maxResponseBodySize = 64 * 1024 // 64KB

var _ worker.Handler[webhookTask] = (*webhookHandler)(nil)

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
		cfg.Timeout = 30 * time.Second
	}

	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 24 * time.Hour
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
	}

	workerOpts := append(
		[]worker.Option{
			worker.WithInterval(cfg.Interval),
			worker.WithMaxConcurrency(1),
		},
		opts...,
	)

	return worker.New(
		"webhook-sender",
		h,
		logger,
		workerOpts...,
	)
}

func (h *webhookHandler) Claim(ctx context.Context) (webhookTask, error) {
	if time.Since(h.cacheCreatedAt) >= h.cacheTTL {
		h.cache = sync.Map{}
		h.cacheCreatedAt = time.Now()
	}

	webhookData, deliveries, err := h.claimNextWebhookData(ctx)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return webhookTask{}, worker.ErrNoTask
		}

		return webhookTask{}, fmt.Errorf("cannot claim next webhook data: %w", err)
	}

	return webhookTask{webhookData: webhookData, deliveries: deliveries}, nil
}

func (h *webhookHandler) Process(ctx context.Context, task webhookTask) error {
	h.processDeliveries(ctx, task.webhookData, task.deliveries)

	return nil
}

func (h *webhookHandler) claimNextWebhookData(ctx context.Context) (*coredata.WebhookData, []pendingDelivery, error) {
	var (
		webhookData coredata.WebhookData
		deliveries  []pendingDelivery
	)

	err := h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := webhookData.LoadNextUnprocessedForUpdate(ctx, tx); err != nil {
			return fmt.Errorf("cannot load next unprocessed webhook data: %w", err)
		}

		scope := coredata.NewScopeFromObjectID(webhookData.ID)

		var configs coredata.WebhookSubscriptions
		if err := configs.LoadMatchingByOrganizationIDAndEventType(
			ctx,
			tx,
			scope,
			webhookData.OrganizationID,
			webhookData.EventType,
		); err != nil {
			return fmt.Errorf("cannot load matching webhook subscriptions: %w", err)
		}

		now := time.Now()

		for _, config := range configs {
			event := &coredata.WebhookEvent{
				ID:                    gid.New(webhookData.ID.TenantID(), coredata.WebhookEventEntityType),
				WebhookDataID:         webhookData.ID,
				WebhookSubscriptionID: config.ID,
				Status:                coredata.WebhookEventStatusPending,
				CreatedAt:             now,
			}

			if err := event.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			deliveries = append(
				deliveries,
				pendingDelivery{
					Event:  event,
					Config: config,
				},
			)
		}

		webhookData.ProcessedAt = &now
		if err := webhookData.UpdateProcessedAt(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update webhook data processed_at: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return &webhookData, deliveries, nil
}

func (h *webhookHandler) processDeliveries(ctx context.Context, webhookData *coredata.WebhookData, deliveries []pendingDelivery) {
	for _, d := range deliveries {
		h.deliver(ctx, webhookData, d)
	}
}

func (h *webhookHandler) deliver(ctx context.Context, webhookData *coredata.WebhookData, d pendingDelivery) {
	scope := coredata.NewScopeFromObjectID(d.Event.ID)

	signingSecret, err := h.getSigningSecret(d.Config.ID.String(), d.Config.EncryptedSigningSecret)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot get signing secret",
			log.Error(err),
			log.String("webhook_data_id", webhookData.ID.String()),
			log.String("subscription_id", d.Config.ID.String()),
		)
		h.updateEventStatus(ctx, d.Event, scope, coredata.WebhookEventStatusFailed, nil)

		return
	}

	response, sendErr := h.doHTTPCall(ctx, d.Event.ID, d.Config.EndpointURL, webhookData, d.Config.ID, signingSecret)

	eventStatus := coredata.WebhookEventStatusSucceeded
	if sendErr != nil {
		eventStatus = coredata.WebhookEventStatusFailed

		h.logger.ErrorCtx(
			ctx,
			"error delivering webhook",
			log.Error(sendErr),
			log.String("webhook_data_id", webhookData.ID.String()),
			log.String("event_id", d.Event.ID.String()),
		)
	}

	h.updateEventStatus(ctx, d.Event, scope, eventStatus, response)
}

func (h *webhookHandler) updateEventStatus(
	ctx context.Context,
	event *coredata.WebhookEvent,
	scope coredata.Scoper,
	status coredata.WebhookEventStatus,
	response json.RawMessage,
) {
	event.Status = status
	event.Response = response

	err := h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return event.UpdateStatus(ctx, tx, scope)
	})
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot update webhook event status",
			log.Error(err),
			log.String("event_id", event.ID.String()),
			log.String("target_status", status.String()),
		)
	}
}

func (h *webhookHandler) getSigningSecret(webhookSubscriptionID string, encryptedSigningSecret []byte) (string, error) {
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
) (json.RawMessage, error) {
	payload := Payload{
		EventID:        eventID.String(),
		SubscriptionID: subscriptionID.String(),
		OrganizationID: webhookData.OrganizationID.String(),
		EventType:      webhookData.EventType.String(),
		CreatedAt:      webhookData.CreatedAt,
		Data:           webhookData.Data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal webhook payload: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := computeSignature(signingSecret, timestamp, body)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Probo-Webhook-Event", webhookData.EventType.String())
	req.Header.Set("X-Probo-Webhook-Organization-Id", webhookData.OrganizationID.String())
	req.Header.Set("X-Probo-Webhook-Timestamp", timestamp)
	req.Header.Set("X-Probo-Webhook-Signature", signature)
	req.Header.Set("X-Probo-Webhook-Host", h.host)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))

	response := buildResponseJSON(resp, respBody)

	switch resp.StatusCode {
	case http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent:
		return response, nil
	default:
		return response, fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
	}
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
