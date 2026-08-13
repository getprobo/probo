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

package complianceportal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestRendererOwnsAccessRequestPresentation(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	portalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
	messageID := gid.New(tenantID, coredata.AgentRunEntityType)
	intent, err := NewRenderer("https://app.example.com").RenderMessage(
		t.Context(),
		bot.Message{
			ID:             messageID,
			OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
			Type:           AccessMessageType,
			Attributes: map[string]any{
				CompliancePortalIDAttribute: portalID.String(),
				RequesterEmailAttribute:     "requester@example.com",
				DocumentsAttribute: []MessageResource{{
					ID:     gid.New(tenantID, coredata.DocumentEntityType).String(),
					Title:  "Security policy",
					Status: "REQUESTED",
				}},
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, intent.Cards, 2)
	assert.Equal(t, AccessCapability+".approve_all", intent.Cards[0].Actions[0].ID)
	assert.Equal(t, AccessCapability+".approve_item", intent.Cards[1].Actions[0].ID)
}
