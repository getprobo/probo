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

package iam

import (
	"bytes"
	"encoding/json"
	"strings"
)

const redactedAuditValue = "[redacted:write-time]"

func redactAuditMetadata(raw []byte) []byte {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("{}")) || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return raw
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return raw
	}

	redactAuditValue(value)

	redacted, err := json.Marshal(value)
	if err != nil {
		return raw
	}

	return redacted
}

func redactAuditValue(value any) {
	object, ok := value.(map[string]any)
	if !ok {
		items, ok := value.([]any)
		if !ok {
			return
		}

		for _, item := range items {
			redactAuditValue(item)
		}

		return
	}

	for key, child := range object {
		if auditMetadataKeyIsSecret(key) {
			object[key] = redactedAuditValue
			continue
		}

		redactAuditValue(child)
	}
}

func auditMetadataKeyIsSecret(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	if normalized == "key" || strings.HasSuffix(normalized, "_key") {
		return true
	}

	return strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "authorization")
}
