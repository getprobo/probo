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
	"encoding/json"
)

// AccessInventoryGrantAttributeSnapshot is the before/after payload stored on
// change events (entitlement display name + presence).
type AccessInventoryGrantAttributeSnapshot struct {
	EntitlementDisplayName string                       `json:"entitlement_display_name"`
	Presence               AccessInventoryGrantPresence `json:"presence"`
}

// DecodeAccessInventoryGrantAttributeSnapshot unmarshals a change-event
// before/after JSONB payload. Invalid or missing payloads return nil.
func DecodeAccessInventoryGrantAttributeSnapshot(
	raw *json.RawMessage,
) *AccessInventoryGrantAttributeSnapshot {
	if raw == nil {
		return nil
	}

	var snapshot AccessInventoryGrantAttributeSnapshot
	if err := json.Unmarshal(*raw, &snapshot); err != nil {
		return nil
	}

	if !snapshot.Presence.IsValid() {
		return nil
	}

	return &snapshot
}
