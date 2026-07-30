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
	"go.probo.inc/probo/pkg/page"
)

func NewMailingList(ml *coredata.MailingList) *MailingList {
	if ml == nil {
		return nil
	}

	return &MailingList{
		ID:             ml.ID,
		OrganizationID: ml.OrganizationID,
		ReplyTo:        ml.ReplyTo,
		CreatedAt:      ml.CreatedAt,
		UpdatedAt:      ml.UpdatedAt,
	}
}

func NewMailingListSubscriber(s *coredata.MailingListSubscriber) *MailingListSubscriber {
	return &MailingListSubscriber{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		MailingListID:  s.MailingListID,
		FullName:       s.FullName,
		Email:          s.Email,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func NewListMailingListSubscribersOutput(
	p *page.Page[*coredata.MailingListSubscriber, coredata.MailingListSubscriberOrderField],
) ListMailingListSubscribersOutput {
	subscribers := make([]*MailingListSubscriber, 0, len(p.Data))
	for _, s := range p.Data {
		subscribers = append(subscribers, NewMailingListSubscriber(s))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListMailingListSubscribersOutput{
		NextCursor:             nextCursor,
		MailingListSubscribers: subscribers,
	}
}

func NewMailingListUpdate(u *coredata.MailingListUpdate) *MailingListUpdate {
	return &MailingListUpdate{
		ID:             u.ID,
		OrganizationID: u.OrganizationID,
		MailingListID:  u.MailingListID,
		Title:          u.Title,
		Body:           u.Body,
		Status:         u.Status,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func NewListMailingListUpdatesOutput(
	p *page.Page[*coredata.MailingListUpdate, coredata.MailingListUpdateOrderField],
) ListMailingListUpdatesOutput {
	updates := make([]*MailingListUpdate, 0, len(p.Data))
	for _, u := range p.Data {
		updates = append(updates, NewMailingListUpdate(u))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListMailingListUpdatesOutput{
		NextCursor:         nextCursor,
		MailingListUpdates: updates,
	}
}
