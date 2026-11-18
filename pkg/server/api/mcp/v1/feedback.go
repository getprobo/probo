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

package mcp_v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
	"go.probo.inc/probo/pkg/validator"
)

const (
	submitFeedbackToolName = "submit_feedback"
	feedbackMessageMaxLen  = 3000
	feedbackEventType      = "feedback:submitted"
)

// feedbackPayload is the webhook delivery body. submit_feedback has no
// subscription and no organization, so those fields are left out, and
// the call is not signed.
type (
	feedbackPayload struct {
		EventID   string        `json:"eventId"`
		EventType string        `json:"eventType"`
		CreatedAt time.Time     `json:"createdAt"`
		Actor     feedbackActor `json:"actor"`
		Data      feedbackData  `json:"data"`
	}

	feedbackActor struct {
		FullName     string    `json:"fullName"`
		EmailAddress mail.Addr `json:"emailAddress"`
	}

	feedbackData struct {
		Message string `json:"message"`
	}
)

// FeedbackConfig is the instance URL that receives submit_feedback.
// The tool is hidden when URL is empty.
type FeedbackConfig struct {
	URL string
}

func NewFeedbackConfig(rawURL string) FeedbackConfig {
	return FeedbackConfig{URL: strings.TrimSpace(rawURL)}
}

func (c FeedbackConfig) Enabled() bool {
	return c.URL != ""
}

func (c FeedbackConfig) client(logger *log.Logger) *http.Client {
	if !c.Enabled() {
		return nil
	}

	return httpclient.DefaultPooledClient(
		httpclient.WithLogger(logger),
		httpclient.WithSSRFProtection(),
	)
}

// feedbackDeliveryError drops the request URL. http.Client errors are
// *url.Error values whose text includes the full URL, and a feedback URL
// can carry a token in its path or query.
func feedbackDeliveryError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		return urlErr.Err
	}

	return err
}

func configureFeedbackTool(server *mcp.Server, cfg FeedbackConfig) {
	if cfg.Enabled() {
		return
	}

	server.RemoveTools(submitFeedbackToolName)
}

func (r *Resolver) submitFeedback(
	ctx context.Context,
	input *types.SubmitFeedbackInput,
) (*mcp.CallToolResult, types.SubmitFeedbackOutput, error) {
	if !r.feedback.Enabled() || r.feedbackClient == nil {
		return nil, types.SubmitFeedbackOutput{}, fmt.Errorf("feedback is not available")
	}

	identity := authn.IdentityFromContext(ctx)
	if identity == nil {
		return nil, types.SubmitFeedbackOutput{}, fmt.Errorf("missing identity")
	}

	message := ""
	if input != nil {
		message = input.Message
	}

	v := validator.New()
	v.Check(
		message,
		"message",
		validator.NotEmpty(),
		validator.MaxRunes(feedbackMessageMaxLen),
		validator.NoHTML(),
		validator.PrintableText(),
	)

	if err := v.Error(); err != nil {
		return nil, types.SubmitFeedbackOutput{}, err
	}

	eventID, err := uuid.NewV7()
	if err != nil {
		return nil, types.SubmitFeedbackOutput{}, fmt.Errorf("internal server error")
	}

	now := time.Now()

	body, err := json.Marshal(
		feedbackPayload{
			EventID:   eventID.String(),
			EventType: feedbackEventType,
			CreatedAt: now,
			Actor: feedbackActor{
				FullName:     identity.FullName,
				EmailAddress: identity.EmailAddress,
			},
			Data: feedbackData{Message: message},
		},
	)
	if err != nil {
		return nil, types.SubmitFeedbackOutput{}, fmt.Errorf("internal server error")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.feedback.URL, bytes.NewReader(body))
	if err != nil {
		return nil, types.SubmitFeedbackOutput{}, fmt.Errorf("internal server error")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Probo-Webhook-Event", feedbackEventType)
	req.Header.Set("X-Probo-Webhook-Timestamp", strconv.FormatInt(now.Unix(), 10))
	req.Header.Set("X-Probo-Webhook-Host", r.baseURL.String())

	resp, err := r.feedbackClient.Do(req)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot deliver feedback", log.Error(feedbackDeliveryError(err)))

		return nil, types.SubmitFeedbackOutput{}, fmt.Errorf("cannot deliver feedback")
	}

	defer func() { _ = resp.Body.Close() }()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.logger.ErrorCtx(ctx, "feedback endpoint rejected the message", log.Int("status", resp.StatusCode))

		return nil, types.SubmitFeedbackOutput{}, fmt.Errorf("cannot deliver feedback")
	}

	return nil, types.SubmitFeedbackOutput{Submitted: true}, nil
}
