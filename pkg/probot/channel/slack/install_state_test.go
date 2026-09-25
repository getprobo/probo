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

package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestInstallStateRoundTrip(t *testing.T) {
	t.Parallel()

	organizationID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)

	state, err := newInstallState(
		"state-secret",
		organizationID,
		identityID,
		"/organizations/example/settings",
	)
	require.NoError(t, err)

	payload, err := validateInstallState("state-secret", state)
	require.NoError(t, err)
	assert.Equal(t, organizationID, payload.Data.OrganizationID)
	assert.Equal(t, identityID, payload.Data.IdentityID)
	assert.NotEmpty(t, payload.Data.Nonce)
	assert.Equal(t, "/organizations/example/settings", payload.Data.ContinueURL)
}

func TestInstallStateRejectsWrongSecret(t *testing.T) {
	t.Parallel()

	state, err := newInstallState(
		"state-secret",
		gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
		gid.New(gid.NilTenant, coredata.IdentityEntityType),
		"",
	)
	require.NoError(t, err)

	_, err = validateInstallState("wrong-secret", state)
	require.Error(t, err)
}
