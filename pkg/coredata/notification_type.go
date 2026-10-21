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
	"encoding"
	"fmt"
)

type (
	NotificationType string
)

const (
	NotificationTypeDocumentSignatureRequested NotificationType = "DOCUMENT_SIGNATURE_REQUESTED"
	NotificationTypeDocumentApprovalRequested  NotificationType = "DOCUMENT_APPROVAL_REQUESTED"
	NotificationTypeTaskAssigned               NotificationType = "TASK_ASSIGNED"
	NotificationTypeTaskMentioned              NotificationType = "TASK_MENTIONED"
)

var (
	_ fmt.Stringer             = NotificationType("")
	_ encoding.TextMarshaler   = NotificationType("")
	_ encoding.TextUnmarshaler = (*NotificationType)(nil)
)

func NotificationTypes() []NotificationType {
	return []NotificationType{
		NotificationTypeDocumentSignatureRequested,
		NotificationTypeDocumentApprovalRequested,
		NotificationTypeTaskAssigned,
		NotificationTypeTaskMentioned,
	}
}

func (v NotificationType) IsValid() bool {
	switch v {
	case NotificationTypeDocumentSignatureRequested,
		NotificationTypeDocumentApprovalRequested,
		NotificationTypeTaskAssigned,
		NotificationTypeTaskMentioned:
		return true
	}

	return false
}

func (v NotificationType) String() string {
	return string(v)
}

func (v NotificationType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *NotificationType) UnmarshalText(text []byte) error {
	val := NotificationType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid NotificationType value: %q", string(text))
	}

	*v = val

	return nil
}
