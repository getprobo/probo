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
		pg            *pg.Client
		logger        *log.Logger
		httpClient    *http.Client
		encryptionKey cipher.EncryptionKey
		cache         sync.Map
		interval      time.Duration
		timeout       time.Duration
	}

	Config struct {
		Interval      time.Duration
		Timeout       time.Duration
		EncryptionKey cipher.EncryptionKey
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

	return &Sender{
		pg:            pg,
		logger:        logger,
		httpClient:    httpclient.DefaultPooledClient(httpclient.WithLogger(logger)),
		encryptionKey: cfg.EncryptionKey,
		interval:      cfg.Interval,
		timeout:       cfg.Timeout,
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
	for {
		event, err := s.claimNextEvent(ctx)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil
			}
			return err
		}

		s.processEvent(ctx, event)
	}
}

func (s *Sender) claimNextEvent(ctx context.Context) (*coredata.WebhookEvent, error) {
	var event coredata.WebhookEvent

	err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		if err := event.LoadNextPendingForUpdate(ctx, tx); err != nil {
			return err
		}

		scope := coredata.NewScopeFromObjectID(event.ID)

		event.Status = coredata.WebhookEventStatusProcessing

		if err := event.UpdateStatus(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update webhook event to processing: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Sender) processEvent(ctx context.Context, event *coredata.WebhookEvent) {
	scope := coredata.NewScopeFromObjectID(event.ID)

	var configs coredata.WebhookConfigurations
	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return configs.LoadMatchingByOrganizationIDAndEventType(
			ctx, conn, scope, event.OrganizationID, event.EventType,
		)
	})
	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot load matching webhook configurations",
			log.Error(err),
			log.String("event_id", event.ID.String()),
		)
		return
	}

	for _, config := range configs {
		s.deliverToConfiguration(ctx, event, config, scope)
	}

	now := time.Now()
	event.Status = coredata.WebhookEventStatusDelivered
	event.ProcessedAt = &now

	err = s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return event.UpdateStatus(ctx, conn, scope)
	})
	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot update webhook event to delivered",
			log.Error(err),
			log.String("event_id", event.ID.String()),
		)
	}
}

func (s *Sender) deliverToConfiguration(
	ctx context.Context,
	event *coredata.WebhookEvent,
	config *coredata.WebhookConfiguration,
	scope coredata.Scoper,
) {
	callID := gid.New(event.ID.TenantID(), coredata.WebhookCallEntityType)

	signingSecret, err := s.getSigningSecret(config.ID.String(), config.EncryptedSigningSecret)
	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot get signing secret",
			log.Error(err),
			log.String("event_id", event.ID.String()),
			log.String("configuration_id", config.ID.String()),
		)
		s.recordCall(ctx, callID, event, config, scope, coredata.WebhookCallStatusFailed, nil)
		return
	}

	response, sendErr := s.doHTTPCall(ctx, callID, config.EndpointURL, event, config.ID, signingSecret)

	callStatus := coredata.WebhookCallStatusSucceeded
	if sendErr != nil {
		callStatus = coredata.WebhookCallStatusFailed
		s.logger.ErrorCtx(ctx, "error delivering webhook event",
			log.Error(sendErr),
			log.String("event_id", event.ID.String()),
			log.String("endpoint_url", config.EndpointURL),
		)
	}

	s.recordCall(ctx, callID, event, config, scope, callStatus, response)
}

func (s *Sender) getSigningSecret(webhookConfigurationID string, encryptedSigningSecret []byte) (string, error) {
	if cached, ok := s.cache.Load(webhookConfigurationID); ok {
		return cached.(string), nil
	}

	plaintext, err := cipher.Decrypt(encryptedSigningSecret, s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("cannot decrypt signing secret: %w", err)
	}

	signingSecret := string(plaintext)
	s.cache.Store(webhookConfigurationID, signingSecret)

	return signingSecret, nil
}

func (s *Sender) doHTTPCall(
	ctx context.Context,
	callID gid.GID,
	endpointURL string,
	event *coredata.WebhookEvent,
	configurationID gid.GID,
	signingSecret string,
) (json.RawMessage, error) {
	payload := map[string]any{
		"eventId":         event.ID.String(),
		"callId":          callID.String(),
		"configurationId": configurationID.String(),
		"eventType":       event.EventType.String(),
		"createdAt":       event.CreatedAt,
		"data":            event.Data,
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
	req.Header.Set("X-Probo-Webhook-Event", event.EventType.String())
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

func (s *Sender) recordCall(
	ctx context.Context,
	callID gid.GID,
	event *coredata.WebhookEvent,
	config *coredata.WebhookConfiguration,
	scope coredata.Scoper,
	status coredata.WebhookCallStatus,
	response json.RawMessage,
) {
	call := coredata.WebhookCall{
		ID:                     callID,
		WebhookEventID:         event.ID,
		WebhookConfigurationID: config.ID,
		EndpointURL:            config.EndpointURL,
		Status:                 status,
		Response:               response,
		CreatedAt:              time.Now(),
	}

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return call.Insert(ctx, conn, scope)
	})
	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot insert webhook call",
			log.Error(err),
			log.String("event_id", event.ID.String()),
			log.String("configuration_id", config.ID.String()),
		)
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
