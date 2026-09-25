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

package linear

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const (
	webhookFutureSkew  = 60 * time.Second
	webhookApplyMaxAge = 24 * time.Hour
)

type (
	WebhookEnvelope struct {
		Action           string          `json:"action"`
		Type             string          `json:"type"`
		Data             json.RawMessage `json:"data"`
		Actor            WebhookActor    `json:"actor"`
		URL              string          `json:"url"`
		WebhookTimestamp int64           `json:"webhookTimestamp"`
		WebhookID        string          `json:"webhookId"`
		OrganizationID   string          `json:"organizationId"`
	}

	WebhookActor struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	IssueWebhookData struct {
		ID          string `json:"id"`
		Identifier  string `json:"identifier"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    int    `json:"priority"`
		DueDate     string `json:"dueDate"`
		UpdatedAt   string `json:"updatedAt"`
		URL         string `json:"url"`
		State       struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"state"`
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
		Assignee *IssueWebhookAssignee `json:"assignee"`
		present  map[string]struct{}
	}

	IssueWebhookAssignee struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	CommentWebhookData struct {
		ID        string `json:"id"`
		Body      string `json:"body"`
		IssueID   string `json:"issueId"`
		UpdatedAt string `json:"updatedAt"`
		User      *struct {
			Email string `json:"email"`
		} `json:"user"`
		present map[string]struct{}
	}
)

func VerifySignature(secret, signature string, body []byte) bool {
	if secret == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

func ParseEnvelope(body []byte) (*WebhookEnvelope, error) {
	var envelope WebhookEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("cannot unmarshal Linear webhook envelope: %w", err)
	}

	return &envelope, nil
}

func (e *WebhookEnvelope) Timestamp(now time.Time) error {
	if e.WebhookTimestamp == 0 {
		return nil
	}

	sentAt := time.UnixMilli(e.WebhookTimestamp)
	if sentAt.After(now.Add(webhookFutureSkew)) {
		return fmt.Errorf("cannot verify Linear webhook timestamp: future")
	}

	if now.Sub(sentAt) > webhookApplyMaxAge {
		return fmt.Errorf("cannot verify Linear webhook timestamp: stale")
	}

	return nil
}

func (e *WebhookEnvelope) IssueData() (*IssueWebhookData, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(e.Data, &raw); err != nil {
		return nil, fmt.Errorf("cannot unmarshal Linear issue webhook data: %w", err)
	}

	var data IssueWebhookData
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal Linear issue webhook data: %w", err)
	}

	data.present = make(map[string]struct{}, len(raw))
	for key := range raw {
		data.present[key] = struct{}{}
	}

	return &data, nil
}

func (d *IssueWebhookData) Has(field string) bool {
	if d == nil || d.present == nil {
		return false
	}

	_, ok := d.present[field]

	return ok
}

func (e *WebhookEnvelope) CommentData() (*CommentWebhookData, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(e.Data, &raw); err != nil {
		return nil, fmt.Errorf("cannot unmarshal Linear comment webhook data: %w", err)
	}

	var data CommentWebhookData
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal Linear comment webhook data: %w", err)
	}

	data.present = make(map[string]struct{}, len(raw))
	for key := range raw {
		data.present[key] = struct{}{}
	}

	return &data, nil
}

func (d *CommentWebhookData) Has(field string) bool {
	if d == nil || d.present == nil {
		return false
	}

	_, ok := d.present[field]

	return ok
}
