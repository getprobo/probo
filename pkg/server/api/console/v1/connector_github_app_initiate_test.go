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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestValidateGitHubAppReconnectConnector_CrossOrganizationIsNotFound(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	requestedOrganizationID := gid.New(tenantID, 1)
	cnnctr := &coredata.Connector{
		OrganizationID: gid.New(tenantID, 2),
		Provider:       coredata.ConnectorProviderGitHub,
		Protocol:       coredata.ConnectorProtocolGitHubApp,
	}

	err := validateGitHubAppReconnectConnector(cnnctr, requestedOrganizationID)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	require.False(t, errors.Is(err, errInvalidReconnectConnector))
}
