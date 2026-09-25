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

package mcp_v1

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

// connectorConnectionStatus reports whether the connector's credential is
// usable right now, by probing the provider and comparing the granted OAuth
// scopes against the current registration.
//
// MCP has no lazy field resolution, so this runs eagerly wherever a Connector
// is built. A provider that accepted the credential and refused the operation
// reports NOT_AUTHORIZED; every other failure collapses to DISCONNECTED,
// because the underlying error wraps provider-controlled text and
// customer-chosen hosts, and stays in the logs rather than travelling to the
// client.
func (r *Resolver) connectorConnectionStatus(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) types.ConnectorConnectionStatus {
	if err := r.accessReview.ProbeConnector(ctx, scope, connectorID); err != nil {
		field := log.String("connector_id", connectorID.String())

		if probeErr, ok := errors.AsType[*accessreview.ProbeError](err); ok && probeErr != nil {
			status := types.ConnectorConnectionStatusDISCONNECTED
			if accessreview.IsProbeOperationRefused(err) {
				status = types.ConnectorConnectionStatusNOTAUTHORIZED
			}

			// Not Probo's failure, so not in the error budget.
			r.logger.WarnCtx(
				ctx,
				"connector credential probe failed",
				field,
				log.String("provider", probeErr.Provider.String()),
				log.String("probe_failure", accessreview.ProbeFailureCode(err)),
				log.String("connection_status", string(status)),
			)

			return status
		}

		r.logger.ErrorCtx(
			ctx,
			"cannot probe connector, reporting disconnected",
			field,
			log.Error(err),
		)

		return types.ConnectorConnectionStatusDISCONNECTED
	}

	needsReconnect, err := r.accessReview.SourceNeedsReconnect(ctx, scope, connectorID)
	if err != nil {
		r.logger.ErrorCtx(
			ctx,
			"cannot determine connector reconnect requirement, reporting disconnected",
			log.String("connector_id", connectorID.String()),
			log.Error(err),
		)

		return types.ConnectorConnectionStatusDISCONNECTED
	}

	if needsReconnect {
		return types.ConnectorConnectionStatusRECONNECTREQUIRED
	}

	return types.ConnectorConnectionStatusCONNECTED
}
