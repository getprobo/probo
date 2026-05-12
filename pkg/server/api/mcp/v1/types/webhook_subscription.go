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
