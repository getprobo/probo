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

	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

// openSession opens authenticated access to one account of a workload
// identity connector, delegating to the provider that knows which role and
// region its settings name. An empty accountID is the account the connector's
// own settings imply.
//
// This is the only place a cloud session is opened. Campaign fetch, source
// name sync and the connection-status probe all route through it, so an
// account id reaches all of them at once instead of one being quietly left
// behind — and at fetch, a forgotten account id produces a complete,
// plausible, signed-off review of the wrong account.
func openSession(
	ctx context.Context,
	federation *identityfederation.Issuer,
	registry *provider.Registry,
	dbConnector *coredata.Connector,
	accountID string,
) (cloud.Session, error) {
	if federation == nil {
		return nil, fmt.Errorf(
			"cannot reach %s connector: identity federation is not configured in this deployment",
			dbConnector.Provider,
		)
	}

	reg, ok := registry.Get(dbConnector.Provider)
	if !ok || reg.WorkloadIdentity == nil {
		return nil, fmt.Errorf(
			"cannot reach %s connector: provider offers no workload identity path",
			dbConnector.Provider,
		)
	}

	session, err := reg.WorkloadIdentity.NewSession(ctx, federation, dbConnector, accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot open cloud session for %s connector: %w", dbConnector.Provider, err)
	}

	return session, nil
}

// connectorAccountExternalID resolves the vendor identifier of the account a
// source reviews, for handing to openSession.
//
// A source with no account is CSV. A NULL external id is a credential the
// vendor never names, and resolves to the empty account id that means
// "whatever the settings imply" — which for such a connector is the only
// account there is.
//
// The account id on the source is a GID and must never reach ARN or project
// construction, which is why this resolution exists rather than passing the
// column through.
func connectorAccountExternalID(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	connectorAccountID *gid.GID,
) (string, error) {
	if connectorAccountID == nil {
		return "", nil
	}

	account := &coredata.ConnectorAccount{}
	if err := account.LoadByID(ctx, conn, scope, *connectorAccountID); err != nil {
		return "", fmt.Errorf("cannot load connector account: %w", err)
	}

	return ref.UnrefOrZero(account.ExternalAccountID), nil
}
