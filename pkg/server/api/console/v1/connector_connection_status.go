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
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

// connectorConnectionStatus reports whether the connector's credential is
// usable right now, by probing the provider and comparing the granted OAuth
// scopes against the current registration.
//
// Obtaining a credential may succeed even when it is expired or invalid (no
// refresh token available, a dead API key, a role Probo can no longer assume),
// so the probe is what makes the answer trustworthy. coredata.ErrResourceNotFound
// is propagated; every caller decides for itself what a missing connector means.
func (r *Resolver) connectorConnectionStatus(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (types.ConnectorConnectionStatus, error) {
	if err := r.accessReview.ProbeConnector(ctx, scope, connectorID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return "", err
		}

		// A code, not the error: probe errors wrap provider-controlled text
		// and customer-chosen hosts, which logging.md keeps out of the logs.
		field := log.String("connector_id", connectorID.String())

		if probeErr, ok := errors.AsType[*accessreview.ProbeError](err); ok && probeErr != nil {
			// The provider took the credential and refused the operation: the
			// customer has a plan or a role to fix, not a key to re-paste.
			status := types.ConnectorConnectionStatusDisconnected
			if accessreview.IsProbeOperationRefused(err) {
				status = types.ConnectorConnectionStatusNotAuthorized
			}

			// Not Probo's failure, so not in the error budget. The status is
			// logged rather than spelled into the message, which would then
			// have to agree with what is returned below.
			r.logger.WarnCtx(
				ctx,
				"connector credential probe failed",
				field,
				log.String("provider", probeErr.Provider.String()),
				log.String("probe_failure", accessreview.ProbeFailureCode(err)),
				log.String("connection_status", string(status)),
			)

			return status, nil
		}

		// A database read, a decrypt, a deployment without identity
		// federation: ours, so it keeps its message in the logs.
		r.logger.ErrorCtx(
			ctx,
			"cannot probe connector, reporting disconnected",
			field,
			log.Error(err),
		)

		return types.ConnectorConnectionStatusDisconnected, nil
	}

	needsReconnect, err := r.accessReview.SourceNeedsReconnect(ctx, scope, connectorID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return "", err
		}

		r.logger.ErrorCtx(ctx, "cannot determine connector reconnect requirement", log.Error(err))

		return "", gqlutils.Internal(ctx)
	}

	if needsReconnect {
		return types.ConnectorConnectionStatusReconnectRequired, nil
	}

	return types.ConnectorConnectionStatusConnected, nil
}
