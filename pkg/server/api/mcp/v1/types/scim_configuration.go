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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewSCIMConfiguration(c *coredata.SCIMConfiguration) *SCIMConfiguration {
	return &SCIMConfiguration{
		ID:             c.ID,
		OrganizationID: c.OrganizationID,
		BridgeID:       c.BridgeID,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func NewSCIMBridge(b *coredata.SCIMBridge) *SCIMBridge {
	return &SCIMBridge{
		ID:                  b.ID,
		OrganizationID:      b.OrganizationID,
		ScimConfigurationID: b.ScimConfigurationID,
		ConnectorID:         b.ConnectorID,
		Type:                b.Type,
		State:               b.State,
		ExcludedUserNames:   b.ExcludedUserNames,
		LastSyncedAt:        b.LastSyncedAt,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
	}
}

func NewSCIMEvent(e *coredata.SCIMEvent) *SCIMEvent {
	return &SCIMEvent{
		ID:                  e.ID,
		OrganizationID:      e.OrganizationID,
		ScimConfigurationID: e.SCIMConfigurationID,
		Method:              e.Method,
		Path:                e.Path,
		StatusCode:          e.StatusCode,
		RequestBody:         e.RequestBody,
		ResponseBody:        e.ResponseBody,
		ErrorMessage:        e.ErrorMessage,
		UserName:            e.UserName,
		IPAddress:           e.IPAddress.String(),
		CreatedAt:           e.CreatedAt,
	}
}

func NewListSCIMEventsOutput(p *page.Page[*coredata.SCIMEvent, coredata.SCIMEventOrderField]) ListSCIMEventsOutput {
	events := make([]*SCIMEvent, 0, len(p.Data))
	for _, e := range p.Data {
		events = append(events, NewSCIMEvent(e))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListSCIMEventsOutput{
		NextCursor: nextCursor,
		ScimEvents: events,
	}
}
