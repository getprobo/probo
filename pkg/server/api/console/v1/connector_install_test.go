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

package console_v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
)

// TestInstallClaimedWorkFitsClaimWindow pins the ordering the whole single-use
// design rests on. The callback holds a claim across the vendor call, the
// second authorization and the write; a claim untouched for longer than
// InstallStateStaleAfter is reclaimable by the next request. If that work could
// outlive the window, a customer's own refresh would reclaim the claim from
// under the in-flight request and the ceremony would run twice.
func TestInstallClaimedWorkFitsClaimWindow(t *testing.T) {
	t.Parallel()

	assert.Less(
		t,
		installClaimedWorkTimeout,
		coredata.InstallStateStaleAfter,
		"work that can outlive its claim would be reclaimed mid-flight and run twice",
	)
}

// TestInstallAuthorizationFailure pins that a denial and a storage fault are
// told apart. The authorizer reports both through the same return, so answering
// 403 for a Postgres failure would tell the caller a decision was taken that
// never was — and would hide an outage behind a permissions message.
func TestInstallAuthorizationFailure(t *testing.T) {
	t.Parallel()

	t.Run("a policy denial is the caller's answer", func(t *testing.T) {
		t.Parallel()

		denied := iam.NewInsufficientPermissionsError(
			gid.New(gid.NewTenantID(), coredata.IdentityEntityType),
			gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
			probo.ActionConnectorInitiate,
		)

		status, rendered := installAuthorizationFailure(denied)
		assert.Equal(t, http.StatusForbidden, status)
		assert.Equal(t, errInstallForbiddenCallback, rendered)
	})

	t.Run("a denial wrapped on the way up is still a denial", func(t *testing.T) {
		t.Parallel()

		denied := iam.NewInsufficientPermissionsError(
			gid.New(gid.NewTenantID(), coredata.IdentityEntityType),
			gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
			probo.ActionConnectorInitiate,
		)

		status, rendered := installAuthorizationFailure(
			fmt.Errorf("cannot authorize connector install: %w", denied),
		)
		assert.Equal(t, http.StatusForbidden, status)
		assert.Equal(t, errInstallForbiddenCallback, rendered)
	})

	t.Run("anything else is an internal fault", func(t *testing.T) {
		t.Parallel()

		status, rendered := installAuthorizationFailure(errors.New("connection refused"))
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, errInstallInternal, rendered)
	})
}

// TestRedirectInstallOutcome pins that an authorized customer whose install
// failed lands back on their own connections page with a readable reason,
// rather than on a JSON error body mid-top-level-navigation. The page toasts
// `error` when no connector_id came with it.
func TestRedirectInstallOutcome(t *testing.T) {
	t.Parallel()

	organizationID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)
	recorder := httptest.NewRecorder()
	safeRedirect := saferedirect.New(func(_ context.Context, host string) bool {
		return host == "console.example"
	})

	redirectInstallOutcome(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/console/v1/connectors/install/crisp/complete", nil),
		log.NewLogger(log.WithName("test")),
		baseurl.MustParse("https://console.example"),
		safeRedirect,
		organizationID,
		installMessageAlreadyUsed,
	)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)

	location, err := url.Parse(recorder.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://console.example/organizations/"+organizationID.String()+"/access-reviews/connections",
		location.Scheme+"://"+location.Host+location.Path,
	)
	assert.Equal(t, installMessageAlreadyUsed, location.Query().Get("error"))

	// No connector_id: there is no connector to select, and the page uses its
	// absence to decide between toasting and adopting the result.
	assert.Empty(t, location.Query().Get("connector_id"))
}

// TestInstallSentinelsCarryNoIdentifiers pins that the public callback's
// answers stay coarse. The route is reachable with no credentials and
// httpserver.RenderError serializes the error straight into the response body,
// so a sentinel that ever grew a GID would hand an anonymous prober the
// organization or identity the state named.
func TestInstallSentinelsCarryNoIdentifiers(t *testing.T) {
	t.Parallel()

	for _, sentinel := range []error{
		errInstallNotFound,
		errInstallInvalidCallback,
		errInstallUnauthorizedCallback,
		errInstallForbiddenCallback,
		errInstallInternal,
	} {
		assert.NotContains(t, sentinel.Error(), "gid:")
	}
}
