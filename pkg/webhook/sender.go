// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Sender struct {
		pg             *pg.Client
		logger         *log.Logger
		httpClient     *http.Client
		encryptionKey  cipher.EncryptionKey
		cache          sync.Map
		cacheCreatedAt time.Time
		cacheTTL       time.Duration
		interval       time.Duration
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
	}

	pendingDelivery struct {
		Event  *coredata.WebhookEvent
		Config *coredata.WebhookSubscription
	}
)

const maxResponseBodySize = 64 * 1024 // 64KB

func NewSender(pg *pg.Client, logger *log.Logger, cfg Config) *Sender {
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 24 * time.Hour
	}

	return &Sender{
		pg:             pg,
		logger:         logger,
		httpClient:     httpclient.DefaultPooledClient(httpclient.WithLogger(logger)),
		encryptionKey:  cfg.EncryptionKey,
		cacheCreatedAt: time.Now(),
		cacheTTL:       cfg.CacheTTL,
		interval:       cfg.Interval,
		timeout:        cfg.Timeout,
	}
}

func (s *Sender) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.interval):
			if err := s.processEvents(ctx); err != nil {
				s.logger.ErrorCtx(ctx, "cannot process webhook events", log.Error(err))
			}
		}
	}
}

func (s *Sender) processEvents(ctx context.Context) error {
	if time.Since(s.cacheCreatedAt) >= s.cacheTTL {
		s.cache = sync.Map{}
		s.cacheCreatedAt = time.Now()
	}

	for {
		webhookData, deliveries, err := s.claimNextWebhookData(ctx)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil
			}
			return fmt.Errorf("cannot claim next webhook data: %w", err)
		}

		s.processDeliveries(ctx, webhookData, deliveries)
	}
}

func (s *Sender) claimNextWebhookData(ctx context.Context) (*coredata.WebhookData, []pendingDelivery, error) {
	var webhookData coredata.WebhookData
	var deliveries []pendingDelivery

	err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
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

			deliveries = append(deliveries, pendingDelivery{
				Event:  event,
				Config: config,
			})
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

func (s *Sender) processDeliveries(ctx context.Context, webhookData *coredata.WebhookData, deliveries []pendingDelivery) {
	for _, d := range deliveries {
		s.deliver(ctx, webhookData, d)
	}
}

func (s *Sender) deliver(ctx context.Context, webhookData *coredata.WebhookData, d pendingDelivery) {
	scope := coredata.NewScopeFromObjectID(d.Event.ID)

	signingSecret, err := s.getSigningSecret(d.Config.ID.String(), d.Config.EncryptedSigningSecret)
	if err != nil {
		s.logger.ErrorCtx(
			ctx,
			"cannot get signing secret",
			log.Error(err),
			log.String("webhook_data_id", webhookData.ID.String()),
			log.String("subscription_id", d.Config.ID.String()),
		)
		s.updateEventStatus(ctx, d.Event, scope, coredata.WebhookEventStatusFailed, nil)
		return
	}

	response, sendErr := s.doHTTPCall(ctx, d.Event.ID, d.Config.EndpointURL, webhookData, d.Config.ID, signingSecret)

	eventStatus := coredata.WebhookEventStatusSucceeded
	if sendErr != nil {
		eventStatus = coredata.WebhookEventStatusFailed
		s.logger.ErrorCtx(
			ctx,
			"error delivering webhook",
			log.Error(sendErr),
			log.String("webhook_data_id", webhookData.ID.String()),
			log.String("event_id", d.Event.ID.String()),
		)
	}

	s.updateEventStatus(ctx, d.Event, scope, eventStatus, response)
}

func (s *Sender) updateEventStatus(
	ctx context.Context,
	event *coredata.WebhookEvent,
	scope coredata.Scoper,
	status coredata.WebhookEventStatus,
	response json.RawMessage,
) {
	event.Status = status
	event.Response = response

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return event.UpdateStatus(ctx, conn, scope)
	})
	if err != nil {
		s.logger.ErrorCtx(
			ctx,
			"cannot update webhook event status",
			log.Error(err),
			log.String("event_id", event.ID.String()),
			log.String("target_status", status.String()),
		)
	}
}

func (s *Sender) getSigningSecret(webhookSubscriptionID string, encryptedSigningSecret []byte) (string, error) {
	if cached, ok := s.cache.Load(webhookSubscriptionID); ok {
		entry := cached.(*cachedSecret)
		if bytes.Equal(entry.encryptedSecret, encryptedSigningSecret) {
			return entry.plaintext, nil
		}
	}

	plaintext, err := cipher.Decrypt(encryptedSigningSecret, s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("cannot decrypt signing secret: %w", err)
	}

	signingSecret := string(plaintext)
	s.cache.Store(
		webhookSubscriptionID,
		&cachedSecret{
			encryptedSecret: encryptedSigningSecret,
			plaintext:       signingSecret,
		},
	)

	return signingSecret, nil
}

func (s *Sender) doHTTPCall(
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

	reqCtx, cancel := context.WithTimeout(ctx, s.timeout)
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

	resp, err := s.httpClient.Do(req)
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
