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

export const WEBHOOK_EVENT_TYPES = [
  { value: "THIRD_PARTY_CREATED", label: "third-party:created" },
  { value: "THIRD_PARTY_UPDATED", label: "third-party:updated" },
  { value: "THIRD_PARTY_DELETED", label: "third-party:deleted" },
  { value: "USER_CREATED", label: "user:created" },
  { value: "USER_UPDATED", label: "user:updated" },
  { value: "USER_DELETED", label: "user:deleted" },
  { value: "OBLIGATION_CREATED", label: "obligation:created" },
  { value: "OBLIGATION_UPDATED", label: "obligation:updated" },
  { value: "OBLIGATION_DELETED", label: "obligation:deleted" },
  { value: "RIGHT_REQUEST_CREATED", label: "right-request:created" },
  { value: "RIGHT_REQUEST_UPDATED", label: "right-request:updated" },
  { value: "RIGHT_REQUEST_DELETED", label: "right-request:deleted" },
  { value: "DOCUMENT_CREATED", label: "document:created" },
  { value: "DOCUMENT_UPDATED", label: "document:updated" },
  { value: "DOCUMENT_ARCHIVED", label: "document:archived" },
  { value: "DOCUMENT_UNARCHIVED", label: "document:unarchived" },
  { value: "DOCUMENT_DELETED", label: "document:deleted" },
  { value: "DOCUMENT_VERSION_CREATED", label: "document-version:created" },
  { value: "DOCUMENT_VERSION_UPDATED", label: "document-version:updated" },
  { value: "DOCUMENT_VERSION_PUBLISHED", label: "document-version:published" },
  { value: "DOCUMENT_VERSION_REJECTED", label: "document-version:rejected" },
  { value: "DOCUMENT_VERSION_DELETED", label: "document-version:deleted" },
  { value: "DOCUMENT_VERSION_SIGNATURE_REQUESTED", label: "document-version-signature:requested" },
  { value: "DOCUMENT_VERSION_SIGNATURE_SIGNED", label: "document-version-signature:signed" },
  { value: "DOCUMENT_VERSION_SIGNATURE_CANCELLED", label: "document-version-signature:cancelled" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_REQUESTED", label: "document-version-approval-quorum:requested" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_UPDATED", label: "document-version-approval-quorum:updated" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_APPROVED", label: "document-version-approval-quorum:approved" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_REJECTED", label: "document-version-approval-quorum:rejected" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_VOIDED", label: "document-version-approval-quorum:voided" },
] as const;

export type WebhookEventTypeValue = (typeof WEBHOOK_EVENT_TYPES)[number]["value"];

export function webhookEventTypeLabel(value: string): string {
  return WEBHOOK_EVENT_TYPES.find(event => event.value === value)?.label ?? value;
}

export function filterWebhookEventTypes(query: string) {
  const needle = query.trim().toLowerCase();
  if (needle === "") {
    return WEBHOOK_EVENT_TYPES;
  }

  return WEBHOOK_EVENT_TYPES.filter(event =>
    event.label.includes(needle) || event.value.toLowerCase().includes(needle),
  );
}
