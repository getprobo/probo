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
	"go.probo.inc/probo/pkg/coredata"
)

func NewElectronicSignature(es *coredata.ElectronicSignature) *ElectronicSignature {
	signature := &ElectronicSignature{
		ID:           es.ID,
		Status:       es.Status,
		DocumentType: es.DocumentType,
		ConsentText:  es.ConsentText,
		LastError:    es.LastError,
		SignedAt:     es.SignedAt,
		CreatedAt:    es.CreatedAt,
		UpdatedAt:    es.UpdatedAt,
	}

	if es.CertificateFileID != nil {
		signature.Certificate = &File{ID: *es.CertificateFileID}
	}

	return signature
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
