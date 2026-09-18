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

	"go.gearno.de/kit/log"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func (r *Resolver) organizationConnectorSettings(
	ctx context.Context,
	input types.CreateOrganizationConnectorInput,
) ([]byte, error) {
	raw, err := probo.MarshalOrganizationSettings(
		probo.OrganizationSettingsInput{
			Provider:                    input.Provider,
			AWSRoleARN:                  ref.UnrefOrZero(input.AWSRoleArn),
			AWSMemberRoleName:           ref.UnrefOrZero(input.AWSMemberRoleName),
			GCPWorkloadIdentityProvider: ref.UnrefOrZero(input.GCPWorkloadIdentityProvider),
			GCPServiceAccountEmail:      ref.UnrefOrZero(input.GCPServiceAccountEmail),
			GCPParent:                   ref.UnrefOrZero(input.GCPParent),
			AzureTenantID:               ref.UnrefOrZero(input.AzureTenantID),
			AzureClientID:               ref.UnrefOrZero(input.AzureClientID),
			AzureEnvironment:            string(ref.UnrefOrZero(input.AzureEnvironment)),
		},
	)
	if err != nil {
		return nil, gqlutils.Invalid(ctx, err)
	}

	return raw, nil
}

// discoveredConnectorAccounts asks the vendor what the credential can reach
// and marks which of those Probo has already recorded, so a client can render
// a discover list without a second round trip.
func (r *Resolver) discoveredConnectorAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) ([]*types.DiscoveredConnectorAccount, error) {
	accounts, err := r.accessReview.DiscoverAccounts(ctx, scope, connectorID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, gqlutils.NotFound(ctx, err)
		}

		// The message wraps provider-controlled text and customer-chosen
		// hosts, which logging keeps out of the response.
		r.logger.WarnCtx(
			ctx,
			"cannot discover connector accounts",
			log.String("connector_id", connectorID.String()),
			log.Error(err),
		)

		return nil, gqlutils.Invalidf(ctx, "cannot reach the provider to list accounts")
	}

	enabled, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.ConnectorAccountOrderField]{
			Field:     coredata.ConnectorAccountOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(
			ctx context.Context,
			cursor *page.Cursor[coredata.ConnectorAccountOrderField],
		) ([]*coredata.ConnectorAccount, error) {
			p, err := r.probo.ConnectorAccounts.ListForConnectorID(ctx, scope, connectorID, cursor)
			if err != nil {
				return nil, err
			}

			return p.Data, nil
		},
	)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot list connector accounts", log.Error(err))

		return nil, gqlutils.Internal(ctx)
	}

	enabledIDs := make(map[string]struct{}, len(enabled))
	for _, account := range enabled {
		if account.ExternalAccountID != nil {
			enabledIDs[*account.ExternalAccountID] = struct{}{}
		}
	}

	discovered := make([]*types.DiscoveredConnectorAccount, 0, len(accounts))

	for _, account := range accounts {
		_, isEnabled := enabledIDs[account.ID]

		discovered = append(discovered, &types.DiscoveredConnectorAccount{
			ExternalAccountID: account.ID,
			Name:              account.Name,
			Enabled:           isEnabled,
		})
	}

	return discovered, nil
}
