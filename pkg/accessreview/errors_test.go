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

package accessreview_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCampaignClientErrors(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	campaignID := gid.New(tenantID, coredata.AccessReviewCampaignEntityType)

	tests := []struct {
		name     string
		err      error
		wantText string
		sentinel error
	}{
		{
			name: "missing sources",
			err:  accessreview.NewCampaignMissingSourcesError(campaignID),
			wantText: fmt.Sprintf(
				"access review campaign %q cannot be started: no scope sources configured",
				campaignID,
			),
			sentinel: accessreview.ErrCampaignMissingSources,
		},
		{
			name:     "not draft",
			err:      accessreview.NewCampaignNotDraftError(campaignID),
			wantText: fmt.Sprintf("access review campaign %q is not in draft", campaignID),
			sentinel: accessreview.ErrCampaignNotDraft,
		},
		{
			name: "not deletable",
			err:  accessreview.NewCampaignNotDeletableError(campaignID),
			wantText: fmt.Sprintf(
				"access review campaign %q cannot be deleted while it is in progress",
				campaignID,
			),
			sentinel: accessreview.ErrCampaignNotDeletable,
		},
		{
			name: "not pending actions",
			err:  accessreview.NewCampaignNotPendingActionsError(campaignID),
			wantText: fmt.Sprintf(
				"access review campaign %q cannot be closed unless it is pending actions",
				campaignID,
			),
			sentinel: accessreview.ErrCampaignNotPendingActions,
		},
		{
			name:     "completed",
			err:      accessreview.NewCampaignCompletedError(campaignID),
			wantText: fmt.Sprintf("access review campaign %q is already completed", campaignID),
			sentinel: accessreview.ErrCampaignCompleted,
		},
		{
			name:     "cancelled",
			err:      accessreview.NewCampaignCancelledError(campaignID),
			wantText: fmt.Sprintf("access review campaign %q is already cancelled", campaignID),
			sentinel: accessreview.ErrCampaignCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantText, tt.err.Error())
			assert.ErrorIs(t, tt.err, tt.sentinel)
		})
	}
}

func TestMissingOAuthScopesError(t *testing.T) {
	t.Parallel()

	err := accessreview.NewMissingOAuthScopesError([]string{
		"https://graph.microsoft.com/AuditLog.Read.All",
		"openid",
	})

	assert.Equal(
		t,
		"Missing required OAuth scopes: AuditLog.Read.All, openid",
		err.Error(),
	)
	assert.ErrorIs(t, err, accessreview.ErrMissingOAuthScopes)
}

func TestProbeUnauthorized(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "refused operation",
			err: accessreview.NewProbeError(
				coredata.ConnectorProviderLangfuse,
				&provider.CredentialRejectedError{StatusCode: http.StatusForbidden, OperationRefused: true},
			),
			want: true,
		},
		{
			name: "refused credential",
			err: accessreview.NewProbeError(
				coredata.ConnectorProviderLangfuse,
				&provider.CredentialRejectedError{StatusCode: http.StatusUnauthorized},
			),
		},
		{
			name: "a host answering with a page is not an authorization verdict",
			err: accessreview.NewProbeError(
				coredata.ConnectorProviderMetabase,
				&provider.NotAnAPIEndpointError{StatusCode: http.StatusOK},
			),
		},
		{name: "no failure at all", err: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, accessreview.IsProbeOperationRefused(tc.err))
		})
	}
}
