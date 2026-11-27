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
	"encoding/json"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewAssetChangeEvent(event *coredata.AssetChangeEvent) *AssetChangeEvent {
	return &AssetChangeEvent{
		ID:             event.ID,
		OrganizationID: event.OrganizationID,
		AssetID:        event.AssetID,
		ObservedAt:     event.ObservedAt,
		AssetSyncID:    event.AssetSyncID,
		ActorID:        event.ActorID,
		ChangeType:     event.ChangeType,
		Before:         rawJSONText(event.Before),
		After:          rawJSONText(event.After),
	}
}

func NewListAssetChangeEventsOutput(
	eventPage *page.Page[*coredata.AssetChangeEvent, coredata.AssetChangeEventOrderField],
) ListAssetChangeEventsOutput {
	events := make([]*AssetChangeEvent, 0, len(eventPage.Data))
	for _, event := range eventPage.Data {
		events = append(events, NewAssetChangeEvent(event))
	}

	var nextCursor *page.CursorKey

	if eventPage.Info.HasNext && len(eventPage.Data) > 0 {
		cursorKey := eventPage.Data[len(eventPage.Data)-1].CursorKey(eventPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAssetChangeEventsOutput{
		NextCursor:        nextCursor,
		AssetChangeEvents: events,
	}
}

func rawJSONText(raw json.RawMessage) *string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	value := string(raw)

	return &value
}
