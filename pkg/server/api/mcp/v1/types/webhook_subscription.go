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

package types

import (
	"encoding/json"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewWebhookSubscription(w *coredata.WebhookSubscription) *WebhookSubscription {
	events := make([]coredata.WebhookEventType, len(w.SelectedEvents))
	copy(events, w.SelectedEvents)

	return &WebhookSubscription{
		ID:             w.ID,
		OrganizationID: w.OrganizationID,
		EndpointURL:    w.EndpointURL,
		SelectedEvents: events,
		CreatedAt:      w.CreatedAt,
		UpdatedAt:      w.UpdatedAt,
	}
}

func NewListWebhookSubscriptionsOutput(p *page.Page[*coredata.WebhookSubscription, coredata.WebhookSubscriptionOrderField]) ListWebhookSubscriptionsOutput {
	subscriptions := make([]*WebhookSubscription, 0, len(p.Data))
	for _, w := range p.Data {
		subscriptions = append(subscriptions, NewWebhookSubscription(w))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListWebhookSubscriptionsOutput{
		NextCursor:           nextCursor,
		WebhookSubscriptions: subscriptions,
	}
}

func NewWebhookEvent(e *coredata.WebhookEvent) *WebhookEvent {
	var response *string

	if len(e.Response) > 0 && string(e.Response) != "null" {
		s := string(json.RawMessage(e.Response))
		response = &s
	}

	return &WebhookEvent{
		ID:                    e.ID,
		WebhookSubscriptionID: e.WebhookSubscriptionID,
		Status:                e.Status,
		Response:              response,
		CreatedAt:             e.CreatedAt,
	}
}

func NewListWebhookEventsOutput(p *page.Page[*coredata.WebhookEvent, coredata.WebhookEventOrderField]) ListWebhookEventsOutput {
	events := make([]*WebhookEvent, 0, len(p.Data))
	for _, e := range p.Data {
		events = append(events, NewWebhookEvent(e))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListWebhookEventsOutput{
		NextCursor:    nextCursor,
		WebhookEvents: events,
	}
}
