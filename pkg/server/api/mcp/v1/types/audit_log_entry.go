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

func NewAuditLogEntry(e *coredata.AuditLogEntry) *AuditLogEntry {
	entry := &AuditLogEntry{
		ID:             e.ID,
		OrganizationID: e.OrganizationID,
		ActorID:        e.ActorID,
		ActorType:      AuditLogEntryActorType(e.ActorType),
		Action:         e.Action,
		ResourceType:   e.ResourceType,
		ResourceID:     e.ResourceID,
		CreatedAt:      e.CreatedAt,
	}

	if len(e.Metadata) > 0 {
		var m map[string]any
		if json.Unmarshal(e.Metadata, &m) == nil {
			entry.Metadata = &m
		}
	}

	return entry
}

func NewListAuditLogEntriesOutput(p *page.Page[*coredata.AuditLogEntry, coredata.AuditLogEntryOrderField]) ListAuditLogEntriesOutput {
	entries := make([]*AuditLogEntry, 0, len(p.Data))
	for _, e := range p.Data {
		entries = append(entries, NewAuditLogEntry(e))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAuditLogEntriesOutput{
		NextCursor:      nextCursor,
		AuditLogEntries: entries,
	}
}
