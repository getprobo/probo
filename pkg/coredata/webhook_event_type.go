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

package coredata

import (
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type WebhookEventType string

const (
	WebhookEventTypeThirdPartyCreated   WebhookEventType = "third-party:created"
	WebhookEventTypeThirdPartyUpdated   WebhookEventType = "third-party:updated"
	WebhookEventTypeThirdPartyDeleted   WebhookEventType = "third-party:deleted"
	WebhookEventTypeUserCreated         WebhookEventType = "user:created"
	WebhookEventTypeUserUpdated         WebhookEventType = "user:updated"
	WebhookEventTypeUserDeleted         WebhookEventType = "user:deleted"
	WebhookEventTypeObligationCreated   WebhookEventType = "obligation:created"
	WebhookEventTypeObligationUpdated   WebhookEventType = "obligation:updated"
	WebhookEventTypeObligationDeleted   WebhookEventType = "obligation:deleted"
	WebhookEventTypeRightRequestCreated WebhookEventType = "right-request:created"
	WebhookEventTypeRightRequestUpdated WebhookEventType = "right-request:updated"
	WebhookEventTypeRightRequestDeleted WebhookEventType = "right-request:deleted"

	WebhookEventTypeDocumentCreated    WebhookEventType = "document:created"
	WebhookEventTypeDocumentUpdated    WebhookEventType = "document:updated"
	WebhookEventTypeDocumentArchived   WebhookEventType = "document:archived"
	WebhookEventTypeDocumentUnarchived WebhookEventType = "document:unarchived"
	WebhookEventTypeDocumentDeleted    WebhookEventType = "document:deleted"

	WebhookEventTypeDocumentVersionCreated   WebhookEventType = "document-version:created"
	WebhookEventTypeDocumentVersionUpdated   WebhookEventType = "document-version:updated"
	WebhookEventTypeDocumentVersionPublished WebhookEventType = "document-version:published"
	WebhookEventTypeDocumentVersionRejected  WebhookEventType = "document-version:rejected"
	WebhookEventTypeDocumentVersionDeleted   WebhookEventType = "document-version:deleted"

	WebhookEventTypeDocumentVersionSignatureRequested WebhookEventType = "document-version-signature:requested"
	WebhookEventTypeDocumentVersionSignatureSigned    WebhookEventType = "document-version-signature:signed"
	WebhookEventTypeDocumentVersionSignatureCancelled WebhookEventType = "document-version-signature:cancelled"

	WebhookEventTypeDocumentVersionApprovalQuorumRequested WebhookEventType = "document-version-approval-quorum:requested"
	WebhookEventTypeDocumentVersionApprovalQuorumUpdated   WebhookEventType = "document-version-approval-quorum:updated"
	WebhookEventTypeDocumentVersionApprovalQuorumApproved  WebhookEventType = "document-version-approval-quorum:approved"
	WebhookEventTypeDocumentVersionApprovalQuorumRejected  WebhookEventType = "document-version-approval-quorum:rejected"
	WebhookEventTypeDocumentVersionApprovalQuorumVoided    WebhookEventType = "document-version-approval-quorum:voided"
)

var (
	_ fmt.Stringer             = WebhookEventType("")
	_ encoding.TextMarshaler   = WebhookEventType("")
	_ encoding.TextUnmarshaler = (*WebhookEventType)(nil)
)

func (v WebhookEventType) IsValid() bool {
	switch v {
	case
		WebhookEventTypeThirdPartyCreated,
		WebhookEventTypeThirdPartyUpdated,
		WebhookEventTypeThirdPartyDeleted,
		WebhookEventTypeUserCreated,
		WebhookEventTypeUserUpdated,
		WebhookEventTypeUserDeleted,
		WebhookEventTypeObligationCreated,
		WebhookEventTypeObligationUpdated,
		WebhookEventTypeObligationDeleted,
		WebhookEventTypeRightRequestCreated,
		WebhookEventTypeRightRequestUpdated,
		WebhookEventTypeRightRequestDeleted,
		WebhookEventTypeDocumentCreated,
		WebhookEventTypeDocumentUpdated,
		WebhookEventTypeDocumentArchived,
		WebhookEventTypeDocumentUnarchived,
		WebhookEventTypeDocumentDeleted,
		WebhookEventTypeDocumentVersionCreated,
		WebhookEventTypeDocumentVersionUpdated,
		WebhookEventTypeDocumentVersionPublished,
		WebhookEventTypeDocumentVersionRejected,
		WebhookEventTypeDocumentVersionDeleted,
		WebhookEventTypeDocumentVersionSignatureRequested,
		WebhookEventTypeDocumentVersionSignatureSigned,
		WebhookEventTypeDocumentVersionSignatureCancelled,
		WebhookEventTypeDocumentVersionApprovalQuorumRequested,
		WebhookEventTypeDocumentVersionApprovalQuorumUpdated,
		WebhookEventTypeDocumentVersionApprovalQuorumApproved,
		WebhookEventTypeDocumentVersionApprovalQuorumRejected,
		WebhookEventTypeDocumentVersionApprovalQuorumVoided:
		return true
	}

	return false
}

func (v WebhookEventType) String() string {
	return string(v)
}

func (v WebhookEventType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *WebhookEventType) UnmarshalText(text []byte) error {
	val := WebhookEventType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid WebhookEventType value: %q", string(text))
	}

	*v = val

	return nil
}

type WebhookEventTypes []WebhookEventType

func (s *WebhookEventTypes) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return s.scanFromString(v)
	case []byte:
		return s.scanFromString(string(v))
	default:
		return fmt.Errorf("unsupported type for WebhookEventTypes: %T", value)
	}
}

func (s *WebhookEventTypes) scanFromString(str string) error {
	str = strings.TrimSpace(str)
	if str == "{}" || str == "" {
		*s = []WebhookEventType{}
		return nil
	}

	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	}

	parts := strings.Split(str, ",")
	result := make([]WebhookEventType, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}

		var et WebhookEventType
		if err := et.UnmarshalText([]byte(part)); err != nil {
			return fmt.Errorf("invalid webhook event type in array: %s", part)
		}

		result[i] = et
	}

	*s = result

	return nil
}

func (s WebhookEventTypes) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "{}", nil
	}

	values := make([]string, len(s))
	for i, et := range s {
		values[i] = et.String()
	}

	return "{" + strings.Join(values, ",") + "}", nil
}
