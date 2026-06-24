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

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type Payload struct {
	EventID        string          `json:"eventId"`
	SubscriptionID string          `json:"subscriptionId"`
	OrganizationID string          `json:"organizationId"`
	EventType      string          `json:"eventType"`
	CreatedAt      time.Time       `json:"createdAt"`
	Data           json.RawMessage `json:"data"`
}

func InsertData(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	eventType coredata.WebhookEventType,
	data any,
) error {
	var configs coredata.WebhookSubscriptions

	exists, err := configs.ExistsByOrganizationIDAndEventType(ctx, tx, scope, organizationID, eventType)
	if err != nil {
		return fmt.Errorf("cannot check webhook subscriptions: %w", err)
	}

	if !exists {
		return nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("cannot marshal webhook event data: %w", err)
	}

	webhookData := &coredata.WebhookData{
		ID:             gid.New(scope.GetTenantID(), coredata.WebhookDataEntityType),
		OrganizationID: organizationID,
		EventType:      eventType,
		Data:           raw,
		CreatedAt:      time.Now(),
	}

	if err = webhookData.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert webhook data: %w", err)
	}

	return nil
}
