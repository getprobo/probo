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

package accessreview

import (
	"context"
	"fmt"

	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

// OpenSession opens authenticated access to a workload identity connector.
// An empty accountID, or the initial account's own id, keeps that account.
// Every consumer (fetch, name resolve, source probe, connector probe) must
// go through this helper so an organization install cannot leave one of them
// on the hub.
func (s *Service) OpenSession(
	ctx context.Context,
	dbConnector *coredata.Connector,
	accountID string,
) (cloud.Session, error) {
	return openSession(ctx, s.federation, s.providerRegistry, dbConnector, accountID)
}

func (h *sourceNameHandler) openSession(
	ctx context.Context,
	dbConnector *coredata.Connector,
	accountID string,
) (cloud.Session, error) {
	return openSession(ctx, h.federation, h.providerRegistry, dbConnector, accountID)
}

func openSession(
	ctx context.Context,
	federation *identityfederation.Issuer,
	providerRegistry *provider.Registry,
	dbConnector *coredata.Connector,
	accountID string,
) (cloud.Session, error) {
	reg, ok := providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.WorkloadIdentity == nil {
		return nil, fmt.Errorf(
			"cannot open cloud session: provider %s offers no workload identity path",
			dbConnector.Provider,
		)
	}

	// The initial account's own id is empty too: AWS would assume the member
	// role instead of the configured RoleARN.
	if accountID != "" {
		initialID, _, err := reg.ResolveInitialAccount(dbConnector)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve initial account: %w", err)
		}

		if accountID == initialID {
			accountID = ""
		}
	}

	session, err := reg.WorkloadIdentity.NewSession(ctx, federation, dbConnector, accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot open cloud session for %s connector: %w", dbConnector.Provider, err)
	}

	return session, nil
}

func (s *Service) initialAccount(c *coredata.Connector) (string, string, error) {
	if s.providerRegistry == nil || c == nil {
		return "", "", nil
	}

	reg, ok := s.providerRegistry.Get(c.Provider)
	if !ok {
		return "", "", nil
	}

	externalID, name, err := reg.ResolveInitialAccount(c)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve initial account: %w", err)
	}

	return externalID, name, nil
}

func (s *Service) DiscoverAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) ([]provider.DiscoveredAccount, error) {
	dbConnector, err := s.loadConfiguredConnector(ctx, scope, connectorID)
	if err != nil {
		return nil, err
	}

	reg, ok := s.providerRegistry.Get(dbConnector.Provider)
	if !ok || !reg.SupportsOrganizationInstall() {
		return []provider.DiscoveredAccount{}, nil
	}

	session, err := s.OpenSession(ctx, dbConnector, "")
	if err != nil {
		return nil, err
	}

	accounts, err := s.providerRegistry.DiscoverAccounts(ctx, session, dbConnector)
	if err != nil {
		return nil, fmt.Errorf("cannot discover accounts for %s connector: %w", dbConnector.Provider, err)
	}

	return accounts, nil
}
