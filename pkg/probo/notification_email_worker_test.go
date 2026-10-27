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

package probo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestNotificationEmailSkip_ExpiresAfterBackoff(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 2, 16, 0, 0, 0, time.UTC)

	var (
		organizationID gid.GID
		profileID      gid.GID
	)

	organizationID[0] = 1
	profileID[0] = 2

	fresh := notificationEmailGroupKey{
		organizationID:     organizationID,
		recipientProfileID: profileID,
		notificationType:   coredata.NotificationTypeTaskAssigned,
	}
	stale := notificationEmailGroupKey{
		organizationID:     organizationID,
		recipientProfileID: profileID,
		notificationType:   coredata.NotificationTypeTaskMentioned,
	}
	handler := &notificationEmailHandler{
		deferredGroups: map[notificationEmailGroupKey]time.Time{},
	}

	handler.deferNotificationEmailGroup(fresh, now)
	handler.deferNotificationEmailGroup(stale, now.Add(-notificationEmailGroupBackoff-time.Second))

	skip := handler.notificationEmailSkip(now)

	require.Len(t, skip.OrganizationIDs, 1)
	assert.Equal(t, organizationID.String(), skip.OrganizationIDs[0])
	assert.Equal(t, profileID.String(), skip.RecipientProfileIDs[0])
	assert.Equal(t, coredata.NotificationTypeTaskAssigned.String(), skip.Types[0])

	skip = handler.notificationEmailSkip(now.Add(notificationEmailGroupBackoff))

	assert.Empty(t, skip.OrganizationIDs)
	assert.Empty(t, skip.RecipientProfileIDs)
	assert.Empty(t, skip.Types)
}
