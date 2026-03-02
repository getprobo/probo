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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
)

func NewElectronicSignature(es *coredata.ElectronicSignature) *ElectronicSignature {
	return &ElectronicSignature{
		ID:           es.ID,
		Status:       es.Status,
		DocumentType: es.DocumentType,
		ConsentText:  es.ConsentText,
		LastError:    es.LastError,
		SignedAt:     es.SignedAt,
		CreatedAt:    es.CreatedAt,
		UpdatedAt:    es.UpdatedAt,
	}
}

func NewElectronicSignatureEvent(ev *coredata.ElectronicSignatureEvent) *ElectronicSignatureEvent {
	return &ElectronicSignatureEvent{
		ID:             ev.ID,
		EventType:      ev.EventType,
		EventSource:    ev.EventSource,
		ActorEmail:     ev.ActorEmail,
		ActorIPAddress: ev.ActorIPAddress,
		ActorUserAgent: ev.ActorUserAgent,
		OccurredAt:     ev.OccurredAt,
		CreatedAt:      ev.CreatedAt,
	}
}
